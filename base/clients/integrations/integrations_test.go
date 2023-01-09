package integrations

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"yalochat.com/salesforce-integration/base/clients/integrations/mocks"
	"yalochat.com/salesforce-integration/base/constants"
)

const (
	url       = "test"
	tokenWA   = "token_wa_test"
	tokenFB   = "token_fb_test"
	channelWA = "channel_wa_test"
	channelFB = "channel_fb_test"
	botWAID   = "botWAID_test"
	botFBID   = "botFBID_test"
	phone     = "+5212222222222"
	webhook   = "https://webhook.site/test"
	userID    = "userID"
)

func TestIntegrationsClient_WebhookRegister(t *testing.T) {

	t.Run("Webhook WA Register Successful", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		client := NewIntegrationsClient(url, tokenWA, tokenFB, channelWA, channelFB, botWAID, botFBID)
		client.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusCreated,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"bot_id": "botWAID_test","channel": "channel_wa_test","webhook": "https://webhook.site/test"}`))),
		}, nil)
		payload := HealthcheckPayload{
			Phone:    phone,
			Webhook:  webhook,
			Provider: constants.WhatsappProvider,
		}
		id, err := client.WebhookRegister(payload)

		if err != nil {
			t.Fatalf("Expected nil error, but retrieved this %#v", err)
		}

		expectedResponse := &HealthcheckResponse{
			BotId:   botWAID,
			Channel: channelWA,
			Webhook: webhook,
		}

		assert.Equal(t, expectedResponse, id)
	})

	t.Run("Webhook FB Register Successful", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		client := NewIntegrationsClient(url, tokenWA, tokenFB, channelWA, channelFB, botWAID, botFBID)
		client.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusCreated,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"bot_id": "botFBID_test","channel": "channel_fb_test","webhook": "https://webhook.site/test"}`))),
		}, nil)
		payload := HealthcheckPayload{
			Phone:    phone,
			Webhook:  webhook,
			Provider: constants.FacebookProvider,
		}
		id, err := client.WebhookRegister(payload)

		if err != nil {
			t.Fatalf("Expected nil error, but retrieved this %#v", err)
		}

		expectedResponse := &HealthcheckResponse{
			BotId:   botFBID,
			Channel: channelFB,
			Webhook: webhook,
		}

		assert.Equal(t, expectedResponse, id)
	})

	t.Run("Webhook Register Error validation payload", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		client := NewIntegrationsClient(url, tokenWA, tokenFB, channelWA, channelFB, botWAID, botFBID)
		client.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusCreated,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"bot_id":"dasfasfasd","channel":"c","webhook":"http://"}`))),
		}, nil)
		payload := HealthcheckPayload{
			Phone:    phone,
			Provider: constants.WhatsappProvider,
		}
		id, err := client.WebhookRegister(payload)

		assert.Error(t, err)
		assert.Empty(t, id)
	})

	t.Run("Webhook Register error SendHTTPRequest", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		client := NewIntegrationsClient(url, tokenWA, tokenFB, channelWA, channelFB, botWAID, botFBID)
		client.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{}, assert.AnError)
		payload := HealthcheckPayload{
			Phone:    phone,
			Webhook:  webhook,
			Provider: constants.WhatsappProvider,
		}
		id, err := client.WebhookRegister(payload)

		assert.Error(t, err)
		assert.Empty(t, id)
	})

	t.Run("Webhook Register error status", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		client := NewIntegrationsClient(url, tokenWA, tokenFB, channelWA, channelFB, botWAID, botFBID)
		client.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"errors": { "message": "XXXX"}`))),
		}, nil)
		payload := HealthcheckPayload{
			Phone:    phone,
			Webhook:  webhook,
			Provider: constants.WhatsappProvider,
		}
		id, err := client.WebhookRegister(payload)

		assert.Error(t, err)
		assert.Empty(t, id)
	})

	t.Run("Webhook Register error response", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		client := NewIntegrationsClient(url, tokenWA, tokenFB, channelWA, channelFB, botWAID, botFBID)
		client.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusCreated,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`ok`))),
		}, nil)
		payload := HealthcheckPayload{
			Phone:    phone,
			Webhook:  webhook,
			Provider: constants.WhatsappProvider,
		}
		id, err := client.WebhookRegister(payload)

		assert.Error(t, err)
		assert.Empty(t, id)
	})
}

func TestIntegrationsClient_WebhookRemove(t *testing.T) {

	t.Run("Webhook Remove Successful", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		client := NewIntegrationsClient(url, tokenWA, tokenFB, channelWA, channelFB, botWAID, botFBID)
		client.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusNoContent,
		}, nil)
		payload := RemoveWebhookPayload{
			Phone:    phone,
			Provider: constants.WhatsappProvider,
		}
		id, err := client.WebhookRemove(payload)

		if err != nil {
			t.Fatalf("Expected nil error, but retrieved this %#v", err)
		}

		assert.Equal(t, true, id)
	})

	t.Run("Webhook Remove Error validation payload", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		client := NewIntegrationsClient(url, tokenWA, tokenFB, channelWA, channelFB, botWAID, botFBID)
		client.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusNoContent,
		}, nil)
		payload := RemoveWebhookPayload{
			Phone:    "",
			Provider: constants.WhatsappProvider,
		}
		id, err := client.WebhookRemove(payload)

		assert.Error(t, err)
		assert.Equal(t, false, id)
	})

	t.Run("Webhook Remove error SendHTTPRequest", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		client := NewIntegrationsClient(url, tokenWA, tokenFB, channelWA, channelFB, botWAID, botFBID)
		client.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{}, assert.AnError)
		payload := RemoveWebhookPayload{
			Phone:    phone,
			Provider: constants.WhatsappProvider,
		}
		id, err := client.WebhookRemove(payload)

		assert.Error(t, err)
		assert.Equal(t, false, id)
	})

	t.Run("Webhook Remove error status", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		client := NewIntegrationsClient(url, tokenWA, tokenFB, channelWA, channelFB, botWAID, botFBID)
		client.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"errors": { "message": "XXXX"}`))),
		}, nil)
		payload := RemoveWebhookPayload{
			Phone:    phone,
			Provider: constants.WhatsappProvider,
		}
		id, err := client.WebhookRemove(payload)

		assert.Error(t, err)
		assert.Equal(t, false, id)
	})

}

func TestIntegrationsClient_SendMessage(t *testing.T) {
	t.Run("Send Text Message Successful", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		client := NewIntegrationsClient(url, tokenWA, tokenFB, channelWA, channelFB, botWAID, botFBID)
		client.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusCreated,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"messages":[{"id": "gBGHUhVRI2ACTwIJQht5EEKBBQyz"}]}`))),
		}, nil)
		payload := &SendTextPayload{
			Type:   "text",
			Text:   TextMessage{Body: "Hola!!"},
			UserID: userID,
		}
		id, err := client.SendMessage(payload, constants.WhatsappProvider)

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
		proxyMock := new(mocks.ProxyInterface)
		client := NewIntegrationsClient(url, tokenWA, tokenFB, channelWA, channelFB, botWAID, botFBID)
		client.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusCreated,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"messages":[{"id": "gBGHUhVRI2ACTwIJQht5EEKBBQyz"}]}`))),
		}, nil)
		payload := &SendImagePayload{
			ID:     "212222222222",
			Type:   "image",
			Image:  Media{Url: "http://miimagen.com/test.jpg", Caption: "Test Image"},
			UserID: userID,
		}
		id, err := client.SendMessage(payload, constants.WhatsappProvider)

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
		proxyMock := new(mocks.ProxyInterface)
		client := NewIntegrationsClient(url, tokenWA, tokenFB, channelWA, channelFB, botWAID, botFBID)
		client.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusCreated,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"messages":[{"id": "gBGHUhVRI2ACTwIJQht5EEKBBQyz"}]}`))),
		}, nil)
		payload := &SendAudioPayload{
			Id:     "212222222222",
			Type:   "audio",
			Audio:  Media{Url: "http://miimagen.com/test.mp3"},
			UserID: userID,
		}
		id, err := client.SendMessage(payload, constants.WhatsappProvider)

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
		proxyMock := new(mocks.ProxyInterface)
		client := NewIntegrationsClient(url, tokenWA, tokenFB, channelWA, channelFB, botWAID, botFBID)
		client.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusCreated,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"messages":[{"id": "gBGHUhVRI2ACTwIJQht5EEKBBQyz"}]}`))),
		}, nil)
		payload := &SendVideoPayload{
			Id:     "212222222222",
			Type:   "video",
			Video:  Media{Url: "http://mitest.com/test.mp4", Caption: "Test video"},
			UserID: userID,
		}
		id, err := client.SendMessage(payload, constants.WhatsappProvider)

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
		proxyMock := new(mocks.ProxyInterface)
		client := NewIntegrationsClient(url, tokenWA, tokenFB, channelWA, channelFB, botWAID, botFBID)
		client.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusCreated,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"messages":[{"id": "gBGHUhVRI2ACTwIJQht5EEKBBQyz"}]}`))),
		}, nil)
		payload := &SendDocumentPayload{
			Id:       "212222222222",
			Type:     "document",
			Document: Media{Url: "http://mitest.com/test.pdf", Caption: "Test document"},
			UserID:   userID,
		}
		id, err := client.SendMessage(payload, constants.WhatsappProvider)

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
		proxyMock := new(mocks.ProxyInterface)
		client := NewIntegrationsClient(url, tokenWA, tokenFB, channelWA, channelFB, botWAID, botFBID)
		client.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"messages":[{"id": "gBGHUhVRI2ACTwIJQht5EEKBBQyz"}]}`))),
		}, nil)

		payload := &SendTextPayload{
			Id:     "212222222222",
			Text:   TextMessage{Body: "Hola!!"},
			UserID: userID,
		}

		id, err := client.SendMessage(payload, constants.WhatsappProvider)

		assert.Error(t, err)
		assert.Empty(t, id)
	})

	t.Run("Webhook Register error SendHTTPRequest", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		client := NewIntegrationsClient(url, tokenWA, tokenFB, channelWA, channelFB, botWAID, botFBID)
		client.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{}, assert.AnError)

		payload := &SendTextPayload{
			Id:     "212222222222",
			Type:   "text",
			Text:   TextMessage{Body: "Hola!!"},
			UserID: userID,
		}
		id, err := client.SendMessage(payload, constants.WhatsappProvider)

		assert.Error(t, err)
		assert.Empty(t, id)
	})

	t.Run("Webhook Register error status", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		client := NewIntegrationsClient(url, tokenWA, tokenFB, channelWA, channelFB, botWAID, botFBID)
		client.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"errors": { "message": "XXXX"}`))),
		}, nil)
		payload := &SendTextPayload{
			Id:     "212222222222",
			Type:   "text",
			Text:   TextMessage{Body: "Hola!!"},
			UserID: userID,
		}
		id, err := client.SendMessage(payload, constants.WhatsappProvider)

		assert.Error(t, err)
		assert.Empty(t, id)
	})

	t.Run("Send Message error response", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		client := NewIntegrationsClient(url, tokenWA, tokenFB, channelWA, channelFB, botWAID, botFBID)
		client.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusCreated,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`ok`))),
		}, nil)
		payload := &SendTextPayload{
			Id:     "212222222222",
			Type:   "text",
			Text:   TextMessage{Body: "Hola!!"},
			UserID: userID,
		}
		id, err := client.SendMessage(payload, constants.WhatsappProvider)

		assert.Error(t, err)
		assert.Empty(t, id)
	})
}
