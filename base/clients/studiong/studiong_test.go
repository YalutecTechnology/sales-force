package studiong

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"yalochat.com/salesforce-integration/base/clients/proxy"
)

const (
	token   = "token"
	userID  = "userID"
	state   = "state"
	timeout = 1
)

func TestCaseClient_CreateContentVersion(t *testing.T) {
	t.Run("send message with studiong success", func(t *testing.T) {
		mock := &proxy.Mock{}
		studioNGClient := NewStudioNGClient("uri", token)
		studioNGClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd"}`))),
		}, nil)

		err := studioNGClient.SendTo(state, userID)
		assert.NoError(t, err)
	})

	t.Run("send message with studiong error code", func(t *testing.T) {
		mock := &proxy.Mock{}
		studioNGClient := NewStudioNGClient("uri", token)
		studioNGClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd"}`))),
		}, nil)

		err := studioNGClient.SendTo(state, userID)
		assert.Error(t, err)
	})

	t.Run("send message with studiong error requiered fields", func(t *testing.T) {
		mock := &proxy.Mock{}
		studioNGClient := NewStudioNGClient("uri", token)
		studioNGClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(``))),
		}, nil)

		err := studioNGClient.SendTo("", userID)
		assert.Error(t, err)
	})

	t.Run("send message with studiong success", func(t *testing.T) {
		mock := &proxy.Mock{}
		studioNGClient := NewStudioNGClient("uri", token)
		studioNGClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd"}`))),
		}, assert.AnError)

		err := studioNGClient.SendTo(state, userID)
		assert.Error(t, err)
	})
}
