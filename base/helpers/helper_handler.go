package helpers

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
    "crypto/rand"

	"github.com/go-playground/validator/v10"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"yalochat.com/salesforce-integration/base/constants"
)

// Govalidator to can validate incoming requests
var (
	Govalidator = validator.New
	MarshalJSON = json.Marshal
	alphabet    = "abcdefghijklmnopqrstuvwxyz1234567890"
)

// ValidatePayloadError to validate when there is an error with payload
const (
	InvalidPayload             = "Invalid payload received"
	MissingAttribute           = "Message payload incomplete"
	MissingParam               = "Missing param"
	ValidatePayloadError       = "Error validating payload"
	ValidatePaginationError    = "Error obtain pagination : %s"
	ValidateFilterAndSortError = "Error obtain filter and sort : %s"
	DateFormat                 = "2006-01-02 15:04:05"
	MissingQueryParam          = "A query param is needed to make the request"
	EmptyResponse              = "No response received"
)

// FailedResponse for description error
type FailedResponse struct {
	//CodeError int
	ErrorDescription string
}

// SuccessResponse for return a descriptive message
type SuccessResponse struct {
	Message string
}

// Error Response for return a stutusCode and error
type ErrorResponse struct {
	StatusCode int
	Error      error
}

// ReadAndUnmarshal to help with payloads
func ReadAndUnmarshal(rc io.ReadCloser, destination interface{}) error {
	payloadBytes, readError := ioutil.ReadAll(rc)

	if readError != nil {
		return readError
	}
	decodeError := json.Unmarshal(payloadBytes, destination)

	if decodeError != nil {
		return decodeError
	}
	return nil
}

func writeTo(w http.ResponseWriter, statusCode int, response interface{}) {
	w.Header().Add("Content-type", "application/json: charset=utf-8")

	responseBytes, marshalError := json.Marshal(response)
	if marshalError != nil {
		logrus.WithFields(logrus.Fields{
			"error": marshalError,
		}).Error("Error writting to client")

		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("{\"success\": false}"))
		return
	}

	w.WriteHeader(statusCode)
	w.Write(responseBytes)
}

// WriteSuccessResponse to format a success response
func WriteSuccessResponse(w http.ResponseWriter, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	resultJSON, _ := MarshalJSON(response)
	w.Write(resultJSON)
}

// WriteFailedResponse to format a failed response
func WriteFailedResponse(w http.ResponseWriter, responseCode int, errorDescription string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(responseCode)
	result, _ := MarshalJSON(FailedResponse{ErrorDescription: errorDescription})
	w.Write(result)
}

// GetPaginationValues obtains the pageNumber and pageSize of the request,
// if both values do not come in the request, they are assigned a default value.
// PageNumber = 1 and pageSize = 10
func GetPaginationValues(r *http.Request) (int64, int64, error) {
	if r.URL.Query().Get("page") == "" && r.URL.Query().Get("size") == "" {
		return 1, 10, nil
	}
	page, err := strconv.ParseInt(r.URL.Query().Get("page"), 10, 64)
	if err != nil {
		return 0, 0, err
	}

	size, err := strconv.ParseInt(r.URL.Query().Get("size"), 10, 64)
	if err != nil {
		return 0, 0, err
	}

	return page, size, nil
}

// GetFilterAndShort obtain the filter and sort of the request,
// it returns a default value if they are not found.
// filter: {} and sort: {inserted_at: -1}
// At the moment you can only filter by string and bool types
func GetFilterAndShort(r *http.Request) (map[string]interface{}, map[string]interface{}) {
	var filter map[string]interface{}
	var sort map[string]interface{}
	filter = bson.M{}
	sort = bson.M{"inserted_at": -1}
	params := r.URL.Query()

	if len(params) == 0 {
		return filter, sort
	}

	for key, values := range params {
		if key == "page" || key == "size" || key == "token" {
			continue
		}
		if key == "sort" {
			s, err := strconv.ParseInt(values[0], 10, 64)
			if err == nil {
				sort = bson.M{"inserted_at": s}
			}
			continue
		}
		if values[0] == "true" {
			filter[ToSnakeCase(key)] = true
			continue
		}

		if values[0] == "false" {
			filter[ToSnakeCase(key)] = false
			continue
		}

		if key == "dateStart" {
			start, err := time.Parse(DateFormat, values[0])

			if err != nil {
				logrus.Infof("Date start invalid: %s", err.Error())
			}
			filter["inserted_at"] = bson.M{"$gte": start}
			continue
		}

		if key == "dateEnd" {
			end, err := time.Parse(DateFormat, values[0])

			if err != nil {
				logrus.Infof("Date end invalid: %s", err.Error())
			}
			if date, ok := filter["inserted_at"].(bson.M); ok {
				date["$lte"] = end
				filter["inserted_at"] = date
				continue
			}
			filter["inserted_at"] = bson.M{"$lte": end}
			continue
		}

		filter[ToSnakeCase(key)] = values[0]
	}
	return filter, sort
}

func ToSnakeCase(str string) string {
	var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func ErrorResponseMap(body io.ReadCloser, unmarshalError string, statusCode int) error {
	responseMap := map[string]interface{}{}
	readAndUnmarshalError := ReadAndUnmarshal(body, &responseMap)

	if readAndUnmarshalError != nil {
		errorMessage := fmt.Sprintf("%s : %s", unmarshalError, readAndUnmarshalError.Error())
		logrus.Error(errorMessage)
		return errors.New(errorMessage)
	}
	errorMessage := fmt.Sprintf("%s : %d", constants.StatusError, statusCode)
	logrus.WithFields(logrus.Fields{
		"response": responseMap,
	}).Error(errorMessage)
	return errors.New(errorMessage)
}

func RandomString(size int) string {
    ll := len(alphabet)
    b := make([]byte, size)
    rand.Read(b)
    for i := 0; i < size; i++ {
        b[i] = alphabet[int(b[i])%ll]
    }
    return string(b)
}
