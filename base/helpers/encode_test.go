package helpers

import (
	"io"
	"net/http"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncode(t *testing.T) {
	type args struct {
		bin []byte
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "success",
			args: args{
				bin: []byte("Hola"),
			},
			want: []byte("SG9sYQ=="),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Encode(tt.args.bin); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Encode() = %v, want %v", got, tt.want)
			}
		})
	}
}

//TODO: Fix this text
func TestGetExportFilename(t *testing.T) {
	type args struct {
		name     string
		mimeType string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		/*{
			name: "success jpg",
			args: args{
				name:     "title",
				mimeType: "image/jpeg",
			},
			want: "title.jpeg",
		},*/
		{
			name: "success png",
			args: args{
				name:     "title",
				mimeType: "image/png",
			},
			want: "title.png",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetExportFilename(tt.args.name, tt.args.mimeType); got != tt.want {
				t.Errorf("GetExportFilename() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetContentAndTypeByReader(t *testing.T) {
	resp, err := http.Get("https://cdn-icons-png.flaticon.com/512/545/545682.png")
	assert.NoError(t, err)
	type args struct {
		reader io.Reader
	}
	tests := []struct {
		name            string
		args            args
		wantContentType string
		wantErr         bool
	}{
		{
			name: "success",
			args: args{
				reader: resp.Body,
			},
			wantContentType: "image/png",
			wantErr:         false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotContentType, _, err := GetContentAndTypeByReader(tt.args.reader)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetContentAndTypeByReader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotContentType != tt.wantContentType {
				t.Errorf("GetContentAndTypeByReader() gotContentType = %v, want %v", gotContentType, tt.wantContentType)
			}

		})
	}
}
