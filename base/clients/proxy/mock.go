package proxy

import (
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"net/http"

	"github.com/stretchr/testify/mock"
)

// Mock mimics the behavior of the proxy implementation
type Mock struct {
	mock.Mock // add a Mock object instance
}

// Set BaseURL
func (m *Mock) SetBaseURL(baseUrl string) {
}

// SendHTTPRequest Sends the HTTP `request` to the `${BaseURL}${uri}` path
func (m *Mock) SendHTTPRequest(mainSpan tracer.Span, request *Request) (*http.Response, error) {
	args := m.Called()
	return args.Get(0).(*http.Response), args.Error(1)
}
