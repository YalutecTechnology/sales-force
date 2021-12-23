package proxy

import (
	"bytes"
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
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

		span, _ := tracer.SpanFromContext(context.Background())
		_, error := proxyInstance.SendHTTPRequest(span, request)

		if error != nil {
			t.Errorf("It should finish with success")
		}
	})

	t.Run("Valid forward with status 404", func(t *testing.T) {
		request := &Request{
			Body:   []byte(""),
			Method: "GET",
			URI:    "https://api.twitter.com/2/users/by/username/lalo8a",
		}

		expectedLog := "Response with status 404"
		span, _ := tracer.SpanFromContext(context.Background())
		var buf bytes.Buffer
		logrus.SetOutput(&buf)
		_, error := proxyInstance.SendHTTPRequest(span, request)
		if error != nil {
			t.Errorf("It should finish with success")
		}
		logs := buf.String()
		if !strings.Contains(logs, expectedLog) {
			t.Fatalf("Logs should contain <%s>, but this was found <%s>", expectedLog, logs)
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

		span, _ := tracer.SpanFromContext(context.Background())
		_, err := proxyInstance.SendHTTPRequest(span, request)

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

		span, _ := tracer.SpanFromContext(context.Background())
		_, error := proxyInstance.SendHTTPRequest(span, request)

		if error != nil {
			t.Errorf("It should finish with success")
		}
	})

	t.Run("Invalid forward without required values in request", func(t *testing.T) {
		request := &Request{
			Body: []byte("Hola"),
		}

		span, _ := tracer.SpanFromContext(context.Background())
		_, err := proxyInstance.SendHTTPRequest(span, request)

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

		span, _ := tracer.SpanFromContext(context.Background())
		_, error := emptyProxyInstance.SendHTTPRequest(span, request)

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

		span, _ := tracer.SpanFromContext(context.Background())
		_, err := proxyInstance.SendHTTPRequest(span, request)

		if err == nil {
			t.Errorf("It should finish with error")
		}

		if !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("Error message should contain %s, but this was found <%s>", expectedError, err.Error())
		}
		NewRequest = http.NewRequest
	})

}
