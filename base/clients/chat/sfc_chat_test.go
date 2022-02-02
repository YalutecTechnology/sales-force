package chat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"yalochat.com/salesforce-integration/base/clients/proxy"
	"yalochat.com/salesforce-integration/base/constants"
	"yalochat.com/salesforce-integration/base/helpers"
)

func TestCreateSession(t *testing.T) {
	sessionResponse := `{"key":"ec550263-354e-477c-b773-7747ebce3f5e!1629334776994!TrfoJ67wmtlYiENsWdaUBu0xZ7M=","id":"ec550263-354e-477c-b773-7747ebce3f5e","clientPollTimeout":40,"affinityToken":"878a1fa0"}`
	span, _ := tracer.SpanFromContext(context.Background())
	t.Run("Get CreateSession Successful", func(t *testing.T) {
		mock := &proxy.Mock{}
		salesforceClient := &SfcChatClient{Proxy: mock}
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(sessionResponse))),
		}, nil)

		sessionResponse, err := salesforceClient.CreateSession(span)

		if err != nil {
			t.Fatalf("Expected nil error, but retrieved this %#v", err)
		}

		if sessionResponse.Id == "" {
			t.Fatalf(`Expected sesionId, but retrieved ""`)
		}
	})

	t.Run("Get CreateSession with error SendHTTPRequest", func(t *testing.T) {
		mock := &proxy.Mock{}
		salesforceClient := &SfcChatClient{Proxy: mock}
		mock.On("SendHTTPRequest").Return(&http.Response{}, assert.AnError)

		sessionResponse, err := salesforceClient.CreateSession(span)

		assert.Error(t, err)
		assert.Empty(t, sessionResponse)
	})

	t.Run("Should fail by status error received", func(t *testing.T) {
		mock := &proxy.Mock{}
		expectedError := fmt.Sprintf("[%d] - %s : %s", http.StatusBadRequest, constants.StatusError, "No version header found")
		salesforceClient := &SfcChatClient{Proxy: mock}
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte("No version header found"))),
		}, nil)

		_, err := salesforceClient.CreateSession(span)

		assert.Equal(t, expectedError, err.Error())
	})

}

func TestCreateChat(t *testing.T) {
	span, _ := tracer.SpanFromContext(context.Background())
	t.Run("Get CreateChat Successful", func(t *testing.T) {
		mock := &proxy.Mock{}
		salesforceClient := &SfcChatClient{Proxy: mock}
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte("OK"))),
		}, nil)

		chatRequest := NewChatRequest("organizationId", "deploymentId", "seassionId", "buttonId", "Eduardo")
		ok, err := salesforceClient.CreateChat(span, "affinityToken", "sessionKey", chatRequest)

		if err != nil {
			t.Fatalf("Expected nil error, but retrieved this %#v", err)
		}

		if !ok {
			t.Fatalf(`Expected true, but retrieved false`)
		}
	})

	t.Run("Should fail by invalid payload", func(t *testing.T) {
		expectedError := helpers.InvalidPayload
		salesforceClient := &SfcChatClient{Proxy: &proxy.Proxy{}}
		_, err := salesforceClient.CreateChat(span, "affinityToken", "sessionKey", ChatRequest{OrganizationId: "OrganizationID"})

		if !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("Error message should contain %s, but this was found <%s>", expectedError, err.Error())
		}
	})

	t.Run("Should fail by status error received", func(t *testing.T) {
		mock := &proxy.Mock{}
		expectedError := fmt.Sprintf("[%d] - %s : %s", http.StatusForbidden, constants.StatusError, "Session required but was invalid or not found")
		salesforceClient := &SfcChatClient{Proxy: mock}
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusForbidden,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte("Session required but was invalid or not found"))),
		}, nil)

		chatRequest := NewChatRequest("organizationId", "deploymentId", "seassionId", "buttonId", "Eduardo")
		ok, err := salesforceClient.CreateChat(span, "affinityToken", "sessionKey", chatRequest)

		assert.Equal(t, expectedError, err.Error())

		if ok {
			t.Fatalf(`Expected false, but retrieved true`)
		}
	})

	t.Run("Should fail by proxy error received", func(t *testing.T) {
		mock := &proxy.Mock{}
		expectedError := constants.ForwardError
		salesforceClient := &SfcChatClient{Proxy: mock}
		mock.On("SendHTTPRequest").Return(&http.Response{}, fmt.Errorf("Error proxying a request"))

		chatRequest := NewChatRequest("organizationId", "deploymentId", "seassionId", "buttonId", "Eduardo")
		_, err := salesforceClient.CreateChat(span, "affinityToken", "sessionKey", chatRequest)

		if !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("Error message should contain %s, but this was found <%s>", expectedError, err.Error())
		}
	})

}

func TestGetMessages(t *testing.T) {
	span, _ := tracer.SpanFromContext(context.Background())
	messagesResponse := `{"messages":[{"type":"ChatRequestSuccess","message":{"connectionTimeout":150000,"estimatedWaitTime":9,"sensitiveDataRules":[],"transcriptSaveEnabled":false,"url":"","queuePosition":1,"customDetails":[],"visitorId":"e5c7268d-ac63-40af-8d1c-d53879f8e637","geoLocation":{"organization":"Telmex","countryName":"Mexico","latitude":19.43,"countryCode":"MX","longitude":-99.13}}},{"type":"QueueUpdate","message":{"estimatedWaitTime":0,"position":0}},{"type":"ChatEstablished","message":{"name":"Everardo G","userId":"0053g000000usWa","items":[],"sneakPeekEnabled":false,"chasitorIdleTimeout":{"isEnabled":false}}},{"type":"AgentTyping","message":{"name":"Everardo G","agentId":"5b39e61e-1c94-4bab-b07e-6318b8c8f484"}},{"type":"ChatMessage","message":{"text":"Ok","name":"Everardo G","schedule":{"responseDelayMilliseconds":0},"agentId":"0053g000000usWa"}},{"type":"AgentNotTyping","message":{"name":"Everardo G","agentId":"5b39e61e-1c94-4bab-b07e-6318b8c8f484"}},{"type":"AgentTyping","message":{"name":"Everardo G","agentId":"5b39e61e-1c94-4bab-b07e-6318b8c8f484"}},{"type":"AgentTyping","message":{"name":"Everardo G","agentId":"5b39e61e-1c94-4bab-b07e-6318b8c8f484"}},{"type":"ChatMessage","message":{"text":"Lo ayudio","name":"Everardo G","schedule":{"responseDelayMilliseconds":0},"agentId":"0053g000000usWa"}},{"type":"AgentNotTyping","message":{"name":"Everardo G","agentId":"5b39e61e-1c94-4bab-b07e-6318b8c8f484"}},{"type":"AgentTyping","message":{"name":"Everardo G","agentId":"5b39e61e-1c94-4bab-b07e-6318b8c8f484"}},{"type":"AgentTyping","message":{"name":"Everardo G","agentId":"5b39e61e-1c94-4bab-b07e-6318b8c8f484"}},{"type":"ChatMessage","message":{"text":"que necesita","name":"Everardo G","schedule":{"responseDelayMilliseconds":0},"agentId":"0053g000000usWa"}},{"type":"AgentNotTyping","message":{"name":"Everardo G","agentId":"5b39e61e-1c94-4bab-b07e-6318b8c8f484"}},{"type":"AgentTyping","message":{"name":"Everardo G","agentId":"5b39e61e-1c94-4bab-b07e-6318b8c8f484"}},{"type":"AgentTyping","message":{"name":"Everardo G","agentId":"5b39e61e-1c94-4bab-b07e-6318b8c8f484"}},{"type":"ChatMessage","message":{"text":"lo ayudo","name":"Everardo G","schedule":{"responseDelayMilliseconds":0},"agentId":"0053g000000usWa"}},{"type":"AgentNotTyping","message":{"name":"Everardo G","agentId":"5b39e61e-1c94-4bab-b07e-6318b8c8f484"}}],"sequence":17,"offset":1636522046}`

	t.Run("Get Messages Successful", func(t *testing.T) {
		mock := &proxy.Mock{}
		salesforceClient := &SfcChatClient{Proxy: mock}
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(messagesResponse))),
		}, nil)

		messages, err := salesforceClient.GetMessages(span, "affinityToken", "key")

		if err != nil {
			t.Fatalf("Expected nil error, but retrieved this %#v", err)
		}

		if len(messages.Messages) != 18 {
			t.Fatalf(`Expected 17 items, but retrieved %d`, len(messages.Messages))
		}
	})

	t.Run("Should fail by status error received", func(t *testing.T) {
		mock := &proxy.Mock{}
		expectedError := fmt.Sprintf("[%d] - %s : %s", http.StatusNoContent, constants.StatusError, "{}")
		salesforceClient := &SfcChatClient{Proxy: mock}
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusNoContent,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte("{}"))),
		}, nil)

		_, err := salesforceClient.GetMessages(span, "affinityToken", "key")

		if err.StatusCode != http.StatusNoContent {
			t.Fatalf("Expected Status not content, but retrieved %v", err.StatusCode)
		}

		assert.Equal(t, expectedError, err.Error.Error())
	})

}

func TestSendMessage(t *testing.T) {
	const (
		affinityToken = "affinityToken"
		sessionKey    = "sessionKey"
	)
	span, _ := tracer.SpanFromContext(context.Background())

	t.Run("Send message Successfuly", func(t *testing.T) {
		mock := &proxy.Mock{}
		sfChatClient := &SfcChatClient{Proxy: mock}

		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`OK`))),
		}, nil)
		payload := MessagePayload{
			Text: "A large text",
		}

		response, err := sfChatClient.SendMessage(span, affinityToken, sessionKey, payload)

		if err != nil {
			t.Fatalf("Expected nil error, but retrieved this %#v", err)
		}

		assert.Equal(t, true, response)
	})

	t.Run("Send message error validation payload", func(t *testing.T) {
		mock := &proxy.Mock{}
		sfChatClient := &SfcChatClient{Proxy: mock}

		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`OK`))),
		}, nil)
		payload := MessagePayload{}
		response, err := sfChatClient.SendMessage(span, affinityToken, sessionKey, payload)

		assert.Error(t, err)
		assert.Empty(t, response)
	})

	t.Run("Send message error SendHTTPRequest", func(t *testing.T) {
		mock := &proxy.Mock{}
		sfChatClient := &SfcChatClient{Proxy: mock}
		mock.On("SendHTTPRequest").Return(&http.Response{}, assert.AnError)
		payload := MessagePayload{
			Text: "A large text",
		}

		response, err := sfChatClient.SendMessage(span, affinityToken, sessionKey, payload)

		assert.Error(t, err)
		assert.Empty(t, response)
	})

	t.Run("Send message error status", func(t *testing.T) {
		mock := &proxy.Mock{}
		sfChatClient := &SfcChatClient{Proxy: mock}
		expectedError := fmt.Sprintf("%s-[%d]: %s", constants.StatusError, http.StatusInternalServerError, "Send message error")

		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`Send message error`))),
		}, nil)
		payload := MessagePayload{
			Text: "A large text",
		}

		response, err := sfChatClient.SendMessage(span, affinityToken, sessionKey, payload)

		assert.Error(t, err)
		assert.Empty(t, response)
		assert.Equal(t, expectedError, err.Error())
	})
}

func TestChatClient_ReconnectSession(t *testing.T) {
	t.Run("ReconnectSession case Successful", func(t *testing.T) {
		mock := &proxy.Mock{}
		chat := &SfcChatClient{Proxy: mock}
		expected := MessagesResponse{
			Messages: []MessageObject{
				{
					Type: "type",
					Message: Message{
						ResetSequence: true,
						AffinityToken: "affinity",
					},
				},
			},
		}

		binExpected, err := json.Marshal(&expected)
		assert.NoError(t, err)

		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(binExpected)),
		}, nil)

		session, err := chat.ReconnectSession("token", "key", "offset")

		assert.NoError(t, err)
		assert.Equal(t, &expected, session)
	})

	t.Run("ReconnectSession Error offset", func(t *testing.T) {
		mock := &proxy.Mock{}
		chat := &SfcChatClient{Proxy: mock}
		expected := MessagesResponse{
			Messages: []MessageObject{
				{
					Type: "type",
					Message: Message{
						ResetSequence: true,
						AffinityToken: "affinity",
					},
				},
			},
		}

		binExpected, err := json.Marshal(&expected)
		assert.NoError(t, err)

		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(binExpected)),
		}, nil)

		session, err := chat.ReconnectSession("token", "key", "")

		assert.Error(t, err)
		assert.Nil(t, session)
	})

	t.Run("ReconnectSession error request", func(t *testing.T) {
		mock := &proxy.Mock{}
		chat := &SfcChatClient{Proxy: mock}
		mock.On("SendHTTPRequest").Return(&http.Response{}, assert.AnError)

		session, err := chat.ReconnectSession("token", "key", "offset")

		assert.Error(t, err)
		assert.Nil(t, session)
	})
}

func TestChatEnd(t *testing.T) {
	const (
		affinityToken = "affinityToken"
		sessionKey    = "sessionKey"
	)

	t.Run("Chat end Successfuly", func(t *testing.T) {
		mock := &proxy.Mock{}
		sfChatClient := &SfcChatClient{Proxy: mock}

		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`OK`))),
		}, nil)

		err := sfChatClient.ChatEnd(affinityToken, sessionKey)
		assert.NoError(t, err)
	})

	t.Run("Chat end error SendHTTPRequest", func(t *testing.T) {
		mock := &proxy.Mock{}
		sfChatClient := &SfcChatClient{Proxy: mock}
		mock.On("SendHTTPRequest").Return(&http.Response{}, assert.AnError)

		err := sfChatClient.ChatEnd(affinityToken, sessionKey)

		assert.Error(t, err)
	})

	t.Run("Chat end error status", func(t *testing.T) {
		mock := &proxy.Mock{}
		expectedError := fmt.Sprintf("%s-[%d] : %s", constants.StatusError, http.StatusInternalServerError, "Chat end error")
		sfChatClient := &SfcChatClient{Proxy: mock}
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`Chat end error`))),
		}, nil)

		err := sfChatClient.ChatEnd(affinityToken, sessionKey)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err.Error())
	})
}

func TestCaseClient_UpdateToken(t *testing.T) {

	t.Run("Update token Succesfull", func(t *testing.T) {
		tokenExpected := "14525542211224"
		sfChatClient := &SfcChatClient{AccessToken: "accessToken"}

		sfChatClient.UpdateToken(tokenExpected)
		assert.Equal(t, tokenExpected, sfChatClient.AccessToken)
	})
}
