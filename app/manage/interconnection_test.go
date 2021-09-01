package manage

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"yalochat.com/salesforce-integration/base/clients/chat"
	"yalochat.com/salesforce-integration/base/clients/proxy"
)

const (
	userID        = "55125421545"
	affinityToken = "affinityToken"
	sessionKey    = "sessionKey"
	sessionId     = "sessionId"
)

func TestHandleLongPolling_test(t *testing.T) {
	manager := CreateManager(&ManagerOptions{AppName: "salesforce-integration"})
	interconnection := &Interconnection{
		UserId:              userId,
		SessionId:           sessionId,
		SessionKey:          sessionKey,
		AffinityToken:       affinityToken,
		Status:              OnHold,
		Provider:            provider,
		SfcChatClient:       &chat.SfcChatClient{},
		finishChannel:       manager.finishInterconnection,
		integrationsChannel: manager.integrationsChannel,
		salesforceChannel:   manager.salesforceChannel,
	}

	t.Run("Handle 204 not content", func(t *testing.T) {
		expectedLog := "Not content events"
		mock := &proxy.Mock{}
		interconnection.SfcChatClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusNoContent,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(``))),
		}, nil).Once().On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"messages":[{"type":"ChatEnded","message":{"geoLocation":{}}}]}`))),
		}, nil).Once()

		var buf bytes.Buffer
		logrus.SetOutput(&buf)
		interconnection.handleLongPolling()
		logs := buf.String()
		if !strings.Contains(logs, expectedLog) {
			t.Fatalf("Logs should contain <%s>, but this was found <%s>", expectedLog, logs)
		}
	})

	t.Run("Handle Status Forbidden", func(t *testing.T) {
		expectedLog := "StatusForbidden"
		mock := &proxy.Mock{}
		interconnection.SfcChatClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusForbidden,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(``))),
		}, nil).Once()

		var buf bytes.Buffer
		logrus.SetOutput(&buf)
		interconnection.handleLongPolling()
		logs := buf.String()
		if !strings.Contains(logs, expectedLog) {
			t.Fatalf("Logs should contain <%s>, but this was found <%s>", expectedLog, logs)
		}

		assert.Equal(t, Closed, interconnection.Status)
		assert.Equal(t, false, interconnection.runnigLongPolling)
	})

	t.Run("Handle Status Service Unavailable", func(t *testing.T) {
		expectedLog := "StatusServiceUnavailable"
		mock := &proxy.Mock{}
		interconnection.SfcChatClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusServiceUnavailable,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(``))),
		}, nil).Once()

		var buf bytes.Buffer
		logrus.SetOutput(&buf)
		interconnection.handleLongPolling()
		logs := buf.String()
		if !strings.Contains(logs, expectedLog) {
			t.Fatalf("Logs should contain <%s>, but this was found <%s>", expectedLog, logs)
		}
		assert.Equal(t, Closed, interconnection.Status)
		assert.Equal(t, false, interconnection.runnigLongPolling)
	})

	t.Run("Handle Other error", func(t *testing.T) {
		expectedLog := "Exists error in long polling"
		mock := &proxy.Mock{}
		interconnection.SfcChatClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusBadGateway,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(``))),
		}, nil).Once()

		var buf bytes.Buffer
		logrus.SetOutput(&buf)
		interconnection.handleLongPolling()
		logs := buf.String()
		if !strings.Contains(logs, expectedLog) {
			t.Fatalf("Logs should contain <%s>, but this was found <%s>", expectedLog, logs)
		}
		assert.Equal(t, Closed, interconnection.Status)
		assert.Equal(t, false, interconnection.runnigLongPolling)
	})

	t.Run("Handle Messages from salesforce", func(t *testing.T) {
		expectedLog := "Get Messages sucessfully"
		mock := &proxy.Mock{}
		interconnection.SfcChatClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"messages":[{"type":"ChatEnded","message":{"geoLocation":{}}}]}`))),
		}, nil).Once()

		var buf bytes.Buffer
		logrus.SetOutput(&buf)
		interconnection.handleLongPolling()
		logs := buf.String()
		if !strings.Contains(logs, expectedLog) {
			t.Fatalf("Logs should contain <%s>, but this was found <%s>", expectedLog, logs)
		}
	})
}

func TestCheckEvent_test(t *testing.T) {
	manager := CreateManager(&ManagerOptions{AppName: "salesforce-integration"})
	interconnection := &Interconnection{
		UserId:              userId,
		SessionId:           sessionId,
		SessionKey:          sessionKey,
		AffinityToken:       affinityToken,
		Status:              OnHold,
		Provider:            provider,
		SfcChatClient:       &chat.SfcChatClient{},
		finishChannel:       manager.finishInterconnection,
		integrationsChannel: manager.integrationsChannel,
		salesforceChannel:   manager.salesforceChannel,
	}

	t.Run("Chat request success", func(t *testing.T) {
		expectedLog := "ChatRequestSuccess"
		event := chat.MessageObject{
			Type:    chat.ChatRequestSuccess,
			Message: chat.Message{},
		}

		var buf bytes.Buffer
		logrus.SetOutput(&buf)
		interconnection.checkEvent(&event)
		logs := buf.String()
		if !strings.Contains(logs, expectedLog) {
			t.Fatalf("Logs should contain <%s>, but this was found <%s>", expectedLog, logs)
		}
	})

}
