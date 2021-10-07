package proxy

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestForward(t *testing.T) {
	baseURL := "https://google.com/"
	proxyInstance := NewProxy(baseURL, 5)

	t.Run("Valid forward with valid request", func(t *testing.T) {
		request := &Request{
			Body:   []byte("Hola"),
			Method: "GET",
			URI:    "/search?q=hola",
		}

		_, error := proxyInstance.SendHTTPRequest(request)

		if error != nil {
			t.Errorf("It should finish with success")
		}
	})

	t.Run("Fail forward with valid request", func(t *testing.T) {
		baseURL := "https://go.com/"
		proxyInstance := NewProxy(baseURL, 5)
		expectedError := "Error proxying a request"
		request := &Request{
			Body:   []byte("Hola"),
			Method: "GET",
			URI:    "/",
		}

		_, err := proxyInstance.SendHTTPRequest(request)

		if err == nil {
			t.Errorf("It should finish with error")
		}

		if !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("Error message should contain %s, but this was found <%s>", expectedError, err.Error())
		}
	})

	t.Run("Valid forward with valid request and Header", func(t *testing.T) {
		header := http.Header{}
		header.Add("Content-Type", "application/json")
		requestHeader := make(map[string]string)
		requestHeader["accept"] = "application/json"
		request := &Request{
			Body:      []byte("Hola"),
			Method:    "GET",
			URI:       "/search?q=hola",
			Header:    header,
			HeaderMap: requestHeader,
		}

		_, error := proxyInstance.SendHTTPRequest(request)

		if error != nil {
			t.Errorf("It should finish with success")
		}
	})

	t.Run("Invalid forward without required values in request", func(t *testing.T) {
		request := &Request{
			Body: []byte("Hola"),
		}

		_, err := proxyInstance.SendHTTPRequest(request)

		if err == nil {
			t.Errorf("It should finish with error")
		}
	})

	t.Run("Invalid forward with invalid request.baseURL", func(t *testing.T) {
		emptyProxyInstance := Proxy{}

		request := &Request{
			Body:   []byte("Hola"),
			Method: "NONE",
			URI:    "/search?q=hola",
		}

		_, error := emptyProxyInstance.SendHTTPRequest(request)

		if error == nil {
			t.Errorf("It should finish with error")
		}
	})

	t.Run("Invalid forward with error making a request", func(t *testing.T) {
		expectedError := "Error making a request"
		NewRequest = func(method, url string, body io.Reader) (*http.Request, error) {
			return &http.Request{}, fmt.Errorf("Error")
		}

		request := &Request{
			Body:   []byte("Hola"),
			Method: "NONE",
			URI:    "/-",
		}

		_, err := proxyInstance.SendHTTPRequest(request)

		if err == nil {
			t.Errorf("It should finish with error")
		}

		if !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("Error message should contain %s, but this was found <%s>", expectedError, err.Error())
		}
		NewRequest = http.NewRequest
	})

}
