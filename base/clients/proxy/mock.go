package proxy

import (
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
func (m *Mock) SendHTTPRequest(request *Request) (*http.Response, error) {
	args := m.Called()
	return args.Get(0).(*http.Response), args.Error(1)
}
