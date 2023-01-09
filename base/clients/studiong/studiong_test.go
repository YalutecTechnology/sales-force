package studiong

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"yalochat.com/salesforce-integration/base/clients/studiong/mocks"
)

const (
	token   = "token"
	userID  = "userID"
	state   = "state"
	timeout = 1
)

func TestCaseClient_CreateContentVersion(t *testing.T) {
	t.Run("send message with studiong success", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		studioNGClient := NewStudioNGClient("uri", token)
		studioNGClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd"}`))),
		}, nil)

		err := studioNGClient.SendTo(state, userID)
		assert.NoError(t, err)
	})

	t.Run("send message with studiong error code", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		studioNGClient := NewStudioNGClient("uri", token)
		studioNGClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd"}`))),
		}, nil)

		err := studioNGClient.SendTo(state, userID)
		assert.Error(t, err)
	})

	t.Run("send message with studiong error requiered fields", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		studioNGClient := NewStudioNGClient("uri", token)
		studioNGClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(``))),
		}, nil)

		err := studioNGClient.SendTo("", userID)
		assert.Error(t, err)
	})

	t.Run("send message with studiong success", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		studioNGClient := NewStudioNGClient("uri", token)
		studioNGClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd"}`))),
		}, assert.AnError)

		err := studioNGClient.SendTo(state, userID)
		assert.Error(t, err)
	})
}
