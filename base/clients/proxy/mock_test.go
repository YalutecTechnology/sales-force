package proxy

import (
	"context"
	"fmt"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"net/http"
	"reflect"
	"testing"
)

func TestMock(t *testing.T) {
	var proxy ProxyInterface
	t.Run("Should retrieve response when SendHTTPRequest is invoked", func(t *testing.T) {
		expected := &http.Response{}
		mock := &Mock{}
		mock.On("SendHTTPRequest").Return(expected, nil)
		proxy = mock

		span, _ := tracer.SpanFromContext(context.Background())
		response, err := proxy.SendHTTPRequest(span, &Request{})

		if err != nil {
			t.Fatalf("expected nil error but got %v", err)
		}
		if !reflect.DeepEqual(expected, response) {
			t.Fatalf("expected %v but got %v", expected, response)
		}
	})

	t.Run("Should retrieve error when SendHTTPRequest is invoked", func(t *testing.T) {
		expected := fmt.Errorf("new error")
		mock := &Mock{}
		mock.On("SendHTTPRequest").Return(&http.Response{}, expected)
		proxy = mock

		span, _ := tracer.SpanFromContext(context.Background())
		_, err := proxy.SendHTTPRequest(span, &Request{})

		if expected != err {
			t.Fatalf("expected %v but got %v", expected, err)
		}
	})
}
