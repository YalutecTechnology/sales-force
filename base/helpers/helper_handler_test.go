package helpers

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

type destination struct {
	ID string `json:"id"`
}

func TestReadAndUnmarshal(t *testing.T) {
	dest := destination{}
	type args struct {
		rc          io.ReadCloser
		destination interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				rc:          ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd"}`))),
				destination: &dest,
			},
		},
		{
			name: "error readAll",
			args: args{
				rc:          ioutil.NopCloser(bytes.NewReader(nil)),
				destination: &dest,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ReadAndUnmarshal(tt.args.rc, tt.args.destination); (err != nil) != tt.wantErr {
				t.Errorf("ReadAndUnmarshal() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_writeTo(t *testing.T) {
	response := destination{}
	var ErrorResponse chan int
	type args struct {
		w          http.ResponseWriter
		statusCode int
		response   interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "success",
			args: args{
				w:          httptest.NewRecorder(),
				statusCode: http.StatusOK,
				response:   &response,
			},
		},
		{
			name: "error",
			args: args{
				w:          httptest.NewRecorder(),
				statusCode: http.StatusOK,
				response:   ErrorResponse,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			writeTo(tt.args.w, tt.args.statusCode, tt.args.response)
		})
	}
}

func TestWriteSuccessResponse(t *testing.T) {
	response := destination{}
	type args struct {
		w        http.ResponseWriter
		response interface{}
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "success",
			args: args{
				w:        httptest.NewRecorder(),
				response: &response,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			WriteSuccessResponse(tt.args.w, tt.args.response)
		})
	}
}

func TestWriteFailedResponse(t *testing.T) {
	type args struct {
		w                http.ResponseWriter
		responseCode     int
		errorDescription string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "success",
			args: args{
				w:                httptest.NewRecorder(),
				responseCode:     http.StatusInternalServerError,
				errorDescription: "internal error",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			WriteFailedResponse(tt.args.w, tt.args.responseCode, tt.args.errorDescription)
		})
	}
}

func TestGetPaginationValues(t *testing.T) {
	req, _ := http.NewRequest("POST", "/v1/integrations/whatsapp/webhook", nil)
	reqValues, _ := http.NewRequest("POST", "/v1/integrations/whatsapp/webhook?page=2&size=100", nil)
	reqErrorSize, _ := http.NewRequest("POST", "/v1/integrations/whatsapp/webhook?page=2&size=dsdsad", nil)
	reqErrorPage, _ := http.NewRequest("POST", "/v1/integrations/whatsapp/webhook?page=sads&size=100", nil)
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		want1   int64
		wantErr bool
	}{
		{
			name: "success default values",
			args: args{
				r: req,
			},
			want:    1,
			want1:   10,
			wantErr: false,
		},
		{
			name: "success get values",
			args: args{
				r: reqValues,
			},
			want:    2,
			want1:   100,
			wantErr: false,
		},
		{
			name: "error size",
			args: args{
				r: reqErrorSize,
			},
			want:    0,
			want1:   0,
			wantErr: true,
		},
		{
			name: "error page",
			args: args{
				r: reqErrorPage,
			},
			want:    0,
			want1:   0,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := GetPaginationValues(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPaginationValues() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetPaginationValues() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("GetPaginationValues() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestGetFilterAndShort(t *testing.T) {
	req, _ := http.NewRequest("POST", "/v1/integrations/whatsapp/webhook", nil)
	reqSort, _ := http.NewRequest("POST", "/v1/integrations/whatsapp/webhook?sort=1&dateStart=2021-10-07%2010:00:00", nil)
	reqSortEndDate, _ := http.NewRequest("POST", "/v1/integrations/whatsapp/webhook?sort=1&dateEnd=2021-10-07%2011:00:00", nil)
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name  string
		args  args
		want  map[string]interface{}
		want1 map[string]interface{}
	}{
		{
			name: "success default sort",
			args: args{
				r: req,
			},
			want: map[string]interface{}{},
			want1: map[string]interface{}{
				"inserted_at": -1,
			},
		},
		{
			name: "success sort",
			args: args{
				r: reqSort,
			},
			want: map[string]interface{}{
				"inserted_at": bson.M{"$gte": time.Date(2021, 10, 7, 10, 0, 0, 0, time.UTC)},
			},
			want1: bson.M{"inserted_at": int64(1)},
		},

		{
			name: "success sort endDate",
			args: args{
				r: reqSortEndDate,
			},
			want: map[string]interface{}{
				"inserted_at": bson.M{"$lte": time.Date(2021, 10, 7, 11, 0, 0, 0, time.UTC)},
			},
			want1: bson.M{"inserted_at": int64(1)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1 := GetFilterAndShort(tt.args.r)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetFilterAndShort() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("GetFilterAndShort() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestToSnakeCase(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "success",
			args: args{
				str: "HelloWorld",
			},
			want: "hello_world",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToSnakeCase(tt.args.str); got != tt.want {
				t.Errorf("ToSnakeCase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestErrorResponseMap(t *testing.T) {
	type args struct {
		body           io.ReadCloser
		unmarshalError string
		statusCode     int
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "success",
			args: args{
				body:           ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd"}`))),
				unmarshalError: "unmarshall error",
				statusCode:     http.StatusNotFound,
			},
			wantErr: true,
		},
		{
			name: "Error unmarshall",
			args: args{
				body:           ioutil.NopCloser(bytes.NewReader([]byte(``))),
				unmarshalError: "unmarshall error",
				statusCode:     http.StatusNotFound,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ErrorResponseMap(tt.args.body, tt.args.unmarshalError, tt.args.statusCode); (err != nil) != tt.wantErr {
				t.Errorf("ErrorResponseMap() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetErrorResponse(t *testing.T) {
	type args struct {
		body           io.ReadCloser
		unmarshalError string
		statusCode     int
	}
	tests := []struct {
		name string
		args args
		want *ErrorResponse
	}{
		{
			name: "success",
			args: args{
				body:           ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd"}`))),
				unmarshalError: "error",
				statusCode:     http.StatusNotFound,
			},
			want: &ErrorResponse{
				StatusCode: http.StatusNotFound,
				Error:      errors.New("Error call with status : 404"),
			},
		},
		{
			name: "error unmarshal",
			args: args{
				body:           ioutil.NopCloser(bytes.NewReader([]byte(``))),
				unmarshalError: "error",
				statusCode:     http.StatusNotFound,
			},
			want: &ErrorResponse{
				StatusCode: http.StatusNotFound,
				Error:      errors.New("error : unexpected end of JSON input"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetErrorResponse(tt.args.body, tt.args.unmarshalError, tt.args.statusCode); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetErrorResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetErrorResponseArrayMap(t *testing.T) {
	type args struct {
		body           io.ReadCloser
		unmarshalError string
		statusCode     int
	}
	tests := []struct {
		name string
		args args
		want *ErrorResponse
	}{
		{
			name: "success",
			args: args{
				body:           ioutil.NopCloser(bytes.NewReader([]byte(`[1,2,3,4]`))),
				unmarshalError: "Error unmarshall",
				statusCode:     http.StatusInternalServerError,
			},
			want: &ErrorResponse{
				StatusCode: http.StatusInternalServerError,
				Error:      errors.New("Error call with status : 500"),
			},
		},

		{
			name: "error unmarshal",
			args: args{
				body:           ioutil.NopCloser(bytes.NewReader([]byte(``))),
				unmarshalError: "Error unmarshall",
				statusCode:     http.StatusInternalServerError,
			},
			want: &ErrorResponse{
				StatusCode: http.StatusInternalServerError,
				Error:      errors.New("Error unmarshall : unexpected end of JSON input"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetErrorResponseArrayMap(tt.args.body, tt.args.unmarshalError, tt.args.statusCode); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetErrorResponseArrayMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestErrorMessage(t *testing.T) {
	type args struct {
		messageTitle string
		err          error
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "success",
			args: args{
				messageTitle: "message",
				err:          assert.AnError,
			},
			want: "message : assert.AnError general error for testing",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ErrorMessage(tt.args.messageTitle, tt.args.err); got != tt.want {
				t.Errorf("ErrorMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}
