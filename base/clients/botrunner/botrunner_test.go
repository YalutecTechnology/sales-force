package botrunner

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"yalochat.com/salesforce-integration/base/clients/proxy"
)

func TestSendTo(t *testing.T) {
	t.Run("Send Message Succesfull", func(t *testing.T) {
		mock := &proxy.Mock{}
		botrunnerClient := &BotRunner{Proxy: mock}
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte("{}"))),
		}, nil)

		requestBody := make(map[string]interface{})
		requestBody["userId"] = "84456484"
		requestBody["botSlug"] = "umileverBot"
		requestBody["state"] = "welcome"
		requestBody["message"] = "Hi"

		ok, err := botrunnerClient.SendTo(requestBody)

		if err != nil {
			t.Fatalf("Expected nil error, but retrieved this %#v", err)
		}

		if !ok {
			t.Fatalf("Expected true, but retrieved false")
		}
	})

	t.Run("Send Message Succesfull with token", func(t *testing.T) {
		mock := &proxy.Mock{}
		botrunnerClient := &BotRunner{Proxy: mock, Token: "1254512"}
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte("{}"))),
		}, nil)

		requestBody := make(map[string]interface{})
		requestBody["userId"] = "84456484"
		requestBody["botSlug"] = "umileverBot"
		requestBody["state"] = "welcome"
		requestBody["message"] = "Hi"

		ok, err := botrunnerClient.SendTo(requestBody)

		if err != nil {
			t.Fatalf("Expected nil error, but retrieved this %#v", err)
		}

		if !ok {
			t.Fatalf("Expected true, but retrieved false")
		}
	})

	t.Run("Should fail by invalid userID received", func(t *testing.T) {
		mock := &proxy.Mock{}
		expectedError := "Invalid userId received"
		botrunnerClient := &BotRunner{Proxy: mock}
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte("{}"))),
		}, nil)

		requestBody := make(map[string]interface{})
		requestBody["state"] = "welcome"
		requestBody["botSlug"] = "umileverBot"
		requestBody["message"] = "Hi"

		ok, err := botrunnerClient.SendTo(requestBody)

		if err == nil {
			t.Fatalf("Expected error, but retrieved nil")
		}

		if !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("Error message should contain %s, but this was found <%s>", expectedError, err.Error())
		}

		if ok {
			t.Fatalf("Expected false, but retrieved true")
		}
	})

	t.Run("Should fail by invalid state received", func(t *testing.T) {
		mock := &proxy.Mock{}
		expectedError := "Invalid state received"
		botrunnerClient := &BotRunner{Proxy: mock}
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte("{}"))),
		}, nil)

		requestBody := make(map[string]interface{})
		requestBody["userId"] = "555015455"
		requestBody["botSlug"] = "umileverBot"
		requestBody["message"] = "Hi"

		ok, err := botrunnerClient.SendTo(requestBody)

		if err == nil {
			t.Fatalf("Expected error, but retrieved nil")
		}

		if !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("Error message should contain %s, but this was found <%s>", expectedError, err.Error())
		}

		if ok {
			t.Fatalf("Expected false, but retrieved true")
		}
	})

	t.Run("Should fail by invalid message received", func(t *testing.T) {
		mock := &proxy.Mock{}
		expectedError := "Invalid message received"
		botrunnerClient := &BotRunner{Proxy: mock}
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte("{}"))),
		}, nil)

		requestBody := make(map[string]interface{})
		requestBody["state"] = "welcome"
		requestBody["botSlug"] = "unileverBot"
		requestBody["userId"] = "555414444"

		ok, err := botrunnerClient.SendTo(requestBody)

		if err == nil {
			t.Fatalf("Expected error, but retrieved nil")
		}

		if !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("Error message should contain %s, but this was found <%s>", expectedError, err.Error())
		}

		if ok {
			t.Fatalf("Expected false, but retrieved true")
		}
	})

	t.Run("Should fail by invalid botSlug received", func(t *testing.T) {
		mock := &proxy.Mock{}
		expectedError := "Invalid botSlug received"
		botrunnerClient := &BotRunner{Proxy: mock}
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte("{}"))),
		}, nil)

		requestBody := make(map[string]interface{})
		requestBody["state"] = "welcome"
		requestBody["message"] = "Hi"
		requestBody["userId"] = "555414444"

		ok, err := botrunnerClient.SendTo(requestBody)

		if err == nil {
			t.Fatalf("Expected error, but retrieved nil")
		}

		if !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("Error message should contain %s, but this was found <%s>", expectedError, err.Error())
		}

		if ok {
			t.Fatalf("Expected false, but retrieved true")
		}
	})

	t.Run("Should fail by proxyError received", func(t *testing.T) {
		mock := &proxy.Mock{}
		expectedError := forwardError
		botrunnerClient := &BotRunner{Proxy: mock}
		mock.On("SendHTTPRequest").Return(&http.Response{}, fmt.Errorf("Error proxying a request"))

		requestBody := make(map[string]interface{})
		requestBody["state"] = "welcome"
		requestBody["message"] = "Hi"
		requestBody["userId"] = "555414444"
		requestBody["botSlug"] = "unileverSlug"

		ok, err := botrunnerClient.SendTo(requestBody)

		if err == nil {
			t.Fatalf("Expected error, but retrieved nil")
		}

		if !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("Error message should contain %s, but this was found <%s>", expectedError, err.Error())
		}

		if ok {
			t.Fatalf("Expected false, but retrieved true")
		}
	})

	t.Run("Should fail by unmarshall error received", func(t *testing.T) {
		mock := &proxy.Mock{}
		expectedError := unmarshallError
		botrunnerClient := &BotRunner{Proxy: mock}
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte("{Invalid Payload:}"))),
		}, nil)

		requestBody := make(map[string]interface{})
		requestBody["state"] = "welcome"
		requestBody["message"] = "Hi"
		requestBody["userId"] = "555414444"
		requestBody["botSlug"] = "unileverSlug"

		ok, err := botrunnerClient.SendTo(requestBody)

		if err == nil {
			t.Fatalf("Expected error, but retrieved nil")
		}

		if !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("Error message should contain %s, but this was found <%s>", expectedError, err.Error())
		}

		if ok {
			t.Fatalf("Expected false, but retrieved true")
		}
	})

	t.Run("Should fail by status error received", func(t *testing.T) {
		mock := &proxy.Mock{}
		expectedError := fmt.Sprintf("%s-[%d] : %s", statusError, http.StatusInternalServerError, "map[error:Bad Request message:child \"userId\" fails because [\"userId\" must only contain alpha-numeric characters] statusCode:400 validation:map[keys:[userId] source:payload]]")
		botrunnerClient := &BotRunner{Proxy: mock}
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte("{\"error\":\"Bad Request\",\"message\":\"child \\\"userId\\\" fails because [\\\"userId\\\" must only contain alpha-numeric characters]\",\"statusCode\":400,\"validation\":{\"keys\":[\"userId\"],\"source\":\"payload\"}}"))),
		}, nil)

		requestBody := make(map[string]interface{})
		requestBody["state"] = "welcome"
		requestBody["message"] = "Hi"
		requestBody["userId"] = "555414444"
		requestBody["botSlug"] = "unileverSlug"

		ok, err := botrunnerClient.SendTo(requestBody)

		if err == nil {
			t.Fatalf("Expected error, but retrieved nil")
		}

		assert.Equal(t, expectedError, err.Error())

		if ok {
			t.Fatalf("Expected false, but retrieved true")
		}
	})

	t.Run("Send message succesfull with body not empty", func(t *testing.T) {
		mock := &proxy.Mock{}
		botrunnerClient := &BotRunner{Proxy: mock}
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte("{\"clientId\":\"5217331175599\",\"message\":\"Hola\",\"state\":\"welcome\",\"userId\":\"5217331175599\"}"))),
		}, nil)

		requestBody := make(map[string]interface{})
		requestBody["userId"] = "84456484"
		requestBody["botSlug"] = "umileverBot"
		requestBody["state"] = "welcome"
		requestBody["message"] = "Hi"

		ok, err := botrunnerClient.SendTo(requestBody)

		if err != nil {
			t.Fatalf("Expected nil error, but retrieved this %#v", err)
		}

		if !ok {
			t.Fatalf("Expected true, but retrieved false")
		}
	})

}
