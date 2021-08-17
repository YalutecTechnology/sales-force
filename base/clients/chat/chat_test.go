package chat

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"yalochat.com/salesforce-integration/base/clients/proxy"
)

func TestGetMessages(t *testing.T) {
	messagesResponse := `{"messages":[{"type":"ChatRequestSuccess","message":{"connectionTimeout":150000,"estimatedWaitTime":9,"sensitiveDataRules":[],"transcriptSaveEnabled":false,"url":"","queuePosition":1,"customDetails":[],"visitorId":"e5c7268d-ac63-40af-8d1c-d53879f8e637","geoLocation":{"organization":"Telmex","countryName":"Mexico","latitude":19.43,"countryCode":"MX","longitude":-99.13}}},{"type":"QueueUpdate","message":{"estimatedWaitTime":0,"position":0}},{"type":"ChatEstablished","message":{"name":"Everardo G","userId":"0053g000000usWa","items":[],"sneakPeekEnabled":false,"chasitorIdleTimeout":{"isEnabled":false}}},{"type":"AgentTyping","message":{"name":"Everardo G","agentId":"5b39e61e-1c94-4bab-b07e-6318b8c8f484"}},{"type":"ChatMessage","message":{"text":"Ok","name":"Everardo G","schedule":{"responseDelayMilliseconds":0},"agentId":"0053g000000usWa"}},{"type":"AgentNotTyping","message":{"name":"Everardo G","agentId":"5b39e61e-1c94-4bab-b07e-6318b8c8f484"}},{"type":"AgentTyping","message":{"name":"Everardo G","agentId":"5b39e61e-1c94-4bab-b07e-6318b8c8f484"}},{"type":"AgentTyping","message":{"name":"Everardo G","agentId":"5b39e61e-1c94-4bab-b07e-6318b8c8f484"}},{"type":"ChatMessage","message":{"text":"Lo ayudio","name":"Everardo G","schedule":{"responseDelayMilliseconds":0},"agentId":"0053g000000usWa"}},{"type":"AgentNotTyping","message":{"name":"Everardo G","agentId":"5b39e61e-1c94-4bab-b07e-6318b8c8f484"}},{"type":"AgentTyping","message":{"name":"Everardo G","agentId":"5b39e61e-1c94-4bab-b07e-6318b8c8f484"}},{"type":"AgentTyping","message":{"name":"Everardo G","agentId":"5b39e61e-1c94-4bab-b07e-6318b8c8f484"}},{"type":"ChatMessage","message":{"text":"que necesita","name":"Everardo G","schedule":{"responseDelayMilliseconds":0},"agentId":"0053g000000usWa"}},{"type":"AgentNotTyping","message":{"name":"Everardo G","agentId":"5b39e61e-1c94-4bab-b07e-6318b8c8f484"}},{"type":"AgentTyping","message":{"name":"Everardo G","agentId":"5b39e61e-1c94-4bab-b07e-6318b8c8f484"}},{"type":"AgentTyping","message":{"name":"Everardo G","agentId":"5b39e61e-1c94-4bab-b07e-6318b8c8f484"}},{"type":"ChatMessage","message":{"text":"lo ayudo","name":"Everardo G","schedule":{"responseDelayMilliseconds":0},"agentId":"0053g000000usWa"}},{"type":"AgentNotTyping","message":{"name":"Everardo G","agentId":"5b39e61e-1c94-4bab-b07e-6318b8c8f484"}}],"sequence":17,"offset":1636522046}`

	t.Run("Get Messages Succesfull", func(t *testing.T) {
		mock := &proxy.Mock{}
		salesforceClient := &SfcChatClient{Proxy: mock}
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(messagesResponse))),
		}, nil)

		messages, err := salesforceClient.GetMessages("affinityToken", "key")

		if err != nil {
			t.Fatalf("Expected nil error, but retrieved this %#v", err)
		}

		if len(messages.Messages) != 18 {
			t.Fatalf(`Expected 17 items, but retrieved %d`, len(messages.Messages))
		}
	})

	t.Run("Should fail by status error received", func(t *testing.T) {
		mock := &proxy.Mock{}
		expectedError := proxy.StatusError
		salesforceClient := &SfcChatClient{Proxy: mock}
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusNoContent,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte("{}"))),
		}, nil)

		_, err := salesforceClient.GetMessages("affinityToken", "key")

		if err.StatusCode != http.StatusNoContent {
			t.Fatalf("Expected Status not content, but retrieved %v", err.StatusCode)
		}

		if !strings.Contains(err.Error.Error(), expectedError) {
			t.Fatalf("Error message should contain %s, but this was found <%s>", expectedError, err.Error.Error())
		}
	})

}
