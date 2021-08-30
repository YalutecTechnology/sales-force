package integrations

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"yalochat.com/salesforce-integration/base/clients/proxy"
)

const (
	url     = "test"
	token   = "token_test"
	channel = "channel_test"
	botId   = "botId_test"
	phone   = "+5212222222222"
	webhook = "https://webhook.site/test"
)

func TestIntegrationsClient_WebhookRegister(t *testing.T) {

	t.Run("Webhook Register Successful", func(t *testing.T) {
		mock := &proxy.Mock{}
		client := NewIntegrationsClient(url, token, channel, botId)
		client.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"bot_id": "botId_test","channel": "channel_test","webhook": "https://webhook.site/test"}`))),
		}, nil)
		payload := HealthcheckPayload{
			Phone:   phone,
			Webhook: webhook,
		}
		id, err := client.WebhookRegister(payload)

		if err != nil {
			t.Fatalf("Expected nil error, but retrieved this %#v", err)
		}

		expectedResponse := &HealthcheckResponse{
			BotId:   botId,
			Channel: channel,
			Webhook: webhook,
		}

		assert.Equal(t, expectedResponse, id)
	})

	t.Run("Webhook Register Error validation payload", func(t *testing.T) {
		mock := &proxy.Mock{}
		client := NewIntegrationsClient(url, token, channel, botId)
		client.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"bot_id":"dasfasfasd","channel":"c","webhook":"http://"}`))),
		}, nil)
		payload := HealthcheckPayload{
			Phone: phone,
		}
		id, err := client.WebhookRegister(payload)

		assert.Error(t, err)
		assert.Empty(t, id)
	})

	t.Run("Webhook Register error SendHTTPRequest", func(t *testing.T) {
		mock := &proxy.Mock{}
		client := NewIntegrationsClient(url, token, channel, botId)
		client.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{}, assert.AnError)
		payload := HealthcheckPayload{
			Phone:   phone,
			Webhook: webhook,
		}
		id, err := client.WebhookRegister(payload)

		assert.Error(t, err)
		assert.Empty(t, id)
	})

	t.Run("Webhook Register error status", func(t *testing.T) {
		mock := &proxy.Mock{}
		client := NewIntegrationsClient(url, token, channel, botId)
		client.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"errors": { "message": "XXXX"}`))),
		}, nil)
		payload := HealthcheckPayload{
			Phone:   phone,
			Webhook: webhook,
		}
		id, err := client.WebhookRegister(payload)

		assert.Error(t, err)
		assert.Empty(t, id)
	})

	t.Run("Webhook Register error response", func(t *testing.T) {
		mock := &proxy.Mock{}
		client := NewIntegrationsClient(url, token, channel, botId)
		client.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`ok`))),
		}, nil)
		payload := HealthcheckPayload{
			Phone:   phone,
			Webhook: webhook,
		}
		id, err := client.WebhookRegister(payload)

		assert.Error(t, err)
		assert.Empty(t, id)
	})
}

func TestIntegrationsClient_WebhookRemove(t *testing.T) {

	t.Run("Webhook Remove Successful", func(t *testing.T) {
		mock := &proxy.Mock{}
		client := NewIntegrationsClient(url, token, channel, botId)
		client.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusNoContent,
		}, nil)
		payload := RemoveWebhookPayload{
			Phone: phone,
		}
		id, err := client.WebhookRemove(payload)

		if err != nil {
			t.Fatalf("Expected nil error, but retrieved this %#v", err)
		}

		assert.Equal(t, true, id)
	})

	t.Run("Webhook Remove Error validation payload", func(t *testing.T) {
		mock := &proxy.Mock{}
		client := NewIntegrationsClient(url, token, channel, botId)
		client.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusNoContent,
		}, nil)
		payload := RemoveWebhookPayload{
			Phone: "",
		}
		id, err := client.WebhookRemove(payload)

		assert.Error(t, err)
		assert.Equal(t, false, id)
	})

	t.Run("Webhook Remove error SendHTTPRequest", func(t *testing.T) {
		mock := &proxy.Mock{}
		client := NewIntegrationsClient(url, token, channel, botId)
		client.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{}, assert.AnError)
		payload := RemoveWebhookPayload{
			Phone: phone,
		}
		id, err := client.WebhookRemove(payload)

		assert.Error(t, err)
		assert.Equal(t, false, id)
	})

	t.Run("Webhook Remove error status", func(t *testing.T) {
		mock := &proxy.Mock{}
		client := NewIntegrationsClient(url, token, channel, botId)
		client.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"errors": { "message": "XXXX"}`))),
		}, nil)
		payload := RemoveWebhookPayload{
			Phone: phone,
		}
		id, err := client.WebhookRemove(payload)

		assert.Error(t, err)
		assert.Equal(t, false, id)
	})

}

func TestIntegrationsClient_SendMessage(t *testing.T) {

	t.Run("Send Text Message Successful", func(t *testing.T) {
		mock := &proxy.Mock{}
		client := NewIntegrationsClient(url, token, channel, botId)
		client.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusCreated,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"messages":[{"id": "gBGHUhVRI2ACTwIJQht5EEKBBQyz"}]}`))),
		}, nil)
		payload := &SendTextPayload{
			Type: "text",
			Text: TextMessage{Body: "Hola!!"},
		}
		id, err := client.SendMessage(payload)

		if err != nil {
			t.Fatalf("Expected nil error, but retrieved this %#v", err)
		}

		expectedResponse := &SendMessageResponse{
			Messages: []MessageId{
				{Id: "gBGHUhVRI2ACTwIJQht5EEKBBQyz"},
			},
		}

		assert.Equal(t, expectedResponse, id)
	})

	t.Run("Send Image Message Successful", func(t *testing.T) {
		mock := &proxy.Mock{}
		client := NewIntegrationsClient(url, token, channel, botId)
		client.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusCreated,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"messages":[{"id": "gBGHUhVRI2ACTwIJQht5EEKBBQyz"}]}`))),
		}, nil)
		payload := &SendImagePayload{
			Id:    "212222222222",
			Type:  "image",
			Image: Media{Url: "http://miimagen.com/test.jpg", Caption: "Test Image"},
		}
		id, err := client.SendMessage(payload)

		if err != nil {
			t.Fatalf("Expected nil error, but retrieved this %#v", err)
		}

		expectedResponse := &SendMessageResponse{
			Messages: []MessageId{
				{Id: "gBGHUhVRI2ACTwIJQht5EEKBBQyz"},
			},
		}

		assert.Equal(t, expectedResponse, id)
	})

	t.Run("Send Audio Message Successful", func(t *testing.T) {
		mock := &proxy.Mock{}
		client := NewIntegrationsClient(url, token, channel, botId)
		client.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusCreated,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"messages":[{"id": "gBGHUhVRI2ACTwIJQht5EEKBBQyz"}]}`))),
		}, nil)
		payload := &SendAudioPayload{
			Id:    "212222222222",
			Type:  "audio",
			Audio: Media{Url: "http://miimagen.com/test.mp3"},
		}
		id, err := client.SendMessage(payload)

		if err != nil {
			t.Fatalf("Expected nil error, but retrieved this %#v", err)
		}

		expectedResponse := &SendMessageResponse{
			Messages: []MessageId{
				{Id: "gBGHUhVRI2ACTwIJQht5EEKBBQyz"},
			},
		}

		assert.Equal(t, expectedResponse, id)
	})

	t.Run("Send Video Message Successful", func(t *testing.T) {
		mock := &proxy.Mock{}
		client := NewIntegrationsClient(url, token, channel, botId)
		client.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusCreated,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"messages":[{"id": "gBGHUhVRI2ACTwIJQht5EEKBBQyz"}]}`))),
		}, nil)
		payload := &SendVideoPayload{
			Id:    "212222222222",
			Type:  "video",
			Video: Media{Url: "http://mitest.com/test.mp4", Caption: "Test video"},
		}
		id, err := client.SendMessage(payload)

		if err != nil {
			t.Fatalf("Expected nil error, but retrieved this %#v", err)
		}

		expectedResponse := &SendMessageResponse{
			Messages: []MessageId{
				{Id: "gBGHUhVRI2ACTwIJQht5EEKBBQyz"},
			},
		}

		assert.Equal(t, expectedResponse, id)
	})

	t.Run("Send Document Message Successful", func(t *testing.T) {
		mock := &proxy.Mock{}
		client := NewIntegrationsClient(url, token, channel, botId)
		client.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusCreated,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"messages":[{"id": "gBGHUhVRI2ACTwIJQht5EEKBBQyz"}]}`))),
		}, nil)
		payload := &SendDocumentPayload{
			Id:       "212222222222",
			Type:     "document",
			Document: Media{Url: "http://mitest.com/test.pdf", Caption: "Test document"},
		}
		id, err := client.SendMessage(payload)

		if err != nil {
			t.Fatalf("Expected nil error, but retrieved this %#v", err)
		}

		expectedResponse := &SendMessageResponse{
			Messages: []MessageId{
				{Id: "gBGHUhVRI2ACTwIJQht5EEKBBQyz"},
			},
		}

		assert.Equal(t, expectedResponse, id)
	})

	t.Run("Webhook Register Error validation payload", func(t *testing.T) {
		mock := &proxy.Mock{}
		client := NewIntegrationsClient(url, token, channel, botId)
		client.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"messages":[{"id": "gBGHUhVRI2ACTwIJQht5EEKBBQyz"}]}`))),
		}, nil)

		payload := &SendTextPayload{
			Id:   "212222222222",
			Text: TextMessage{Body: "Hola!!"},
		}

		id, err := client.SendMessage(payload)

		assert.Error(t, err)
		assert.Empty(t, id)
	})

	t.Run("Webhook Register error SendHTTPRequest", func(t *testing.T) {
		mock := &proxy.Mock{}
		client := NewIntegrationsClient(url, token, channel, botId)
		client.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{}, assert.AnError)

		payload := &SendTextPayload{
			Id:   "212222222222",
			Type: "text",
			Text: TextMessage{Body: "Hola!!"},
		}
		id, err := client.SendMessage(payload)

		assert.Error(t, err)
		assert.Empty(t, id)
	})

	t.Run("Webhook Register error status", func(t *testing.T) {
		mock := &proxy.Mock{}
		client := NewIntegrationsClient(url, token, channel, botId)
		client.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"errors": { "message": "XXXX"}`))),
		}, nil)
		payload := &SendTextPayload{
			Id:   "212222222222",
			Type: "text",
			Text: TextMessage{Body: "Hola!!"},
		}
		id, err := client.SendMessage(payload)

		assert.Error(t, err)
		assert.Empty(t, id)
	})

	t.Run("Send Message error response", func(t *testing.T) {
		mock := &proxy.Mock{}
		client := NewIntegrationsClient(url, token, channel, botId)
		client.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusCreated,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`ok`))),
		}, nil)
		payload := &SendTextPayload{
			Id:   "212222222222",
			Type: "text",
			Text: TextMessage{Body: "Hola!!"},
		}
		id, err := client.SendMessage(payload)

		assert.Error(t, err)
		assert.Empty(t, id)
	})
}
