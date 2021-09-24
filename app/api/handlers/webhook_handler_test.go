package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	ddrouter "gopkg.in/DataDog/dd-trace-go.v1/contrib/julienschmidt/httprouter"
	"yalochat.com/salesforce-integration/app/manage"
	"yalochat.com/salesforce-integration/base/cache"
	"yalochat.com/salesforce-integration/base/helpers"
	"yalochat.com/salesforce-integration/base/models"
)

const (
	testURL   = "/v1/integrations/whatsapp/webhook"
	testURLFB = "/v1/integrations/facebook/webhook"
)

func TestApp_webhook(t *testing.T) {
	handler := ddrouter.New(ddrouter.WithServiceName("webhook.http"))
	m, s := cache.CreateRedisServer()
	defer m.Close()
	defer s.Close()
	API(handler, &manage.ManagerOptions{
		AppName: "webhook",
		RedisOptions: cache.RedisOptions{
			FailOverOptions: &redis.FailoverOptions{
				MasterName:    s.MasterInfo().Name,
				SentinelAddrs: []string{s.Addr()},
			},
			SessionsTTL: time.Second * 1,
		},
	}, ApiConfig{
		IntegrationsSignature: "secret",
	})

	t.Run("Should return success", func(t *testing.T) {
		requestURL := testURL
		body := models.IntegrationsRequest{
			ID:        "id",
			Timestamp: "1234556",
			Type:      "text",
			From:      "5555555555",
			Text: models.Text{
				Body: "Hello",
			},
		}

		binBody, err := json.Marshal(body)
		assert.NoError(t, err)

		req, _ := http.NewRequest("POST", requestURL, bytes.NewBuffer(binBody))
		req.Header.Add("x-yalochat-signature", "secret")
		response := httptest.NewRecorder()
		expectedLog := helpers.SuccessResponse{Message: "insert success"}
		binexpectedLog, err := json.Marshal(expectedLog)
		assert.NoError(t, err)

		handler.ServeHTTP(response, req)

		if response.Code != http.StatusOK {
			t.Errorf("Response should be %v, but it answer with %v ", http.StatusOK, response.Code)
		}

		if !strings.Contains(response.Body.String(), string(binexpectedLog)) {
			t.Errorf("Response should be %v, but it answer with %v ", expectedLog, response.Body.String())
		}

	})

	/*t.Run("Should return error signature requiered", func(t *testing.T) {
		requestURL := testURL
		body := models.IntegrationsRequest{
			ID:        "id",
			Timestamp: "1234556",
			Type:      "text",
			From:      "5555555555",
			Text: models.Text{
				Body: "Hello",
			},
		}

		binBody, err := json.Marshal(body)
		assert.NoError(t, err)

		req, _ := http.NewRequest("POST", requestURL, bytes.NewBuffer(binBody))
		req.Header.Add("x-yalochat-signature", "")
		response := httptest.NewRecorder()
		expectedLog := helpers.FailedResponse{
			ErrorDescription: "x-yalochat-signature required, header invalid.",
		}

		handler.ServeHTTP(response, req)

		if response.Code != http.StatusUnauthorized {
			t.Errorf("Response should be %v, but it answer with %v ", http.StatusUnauthorized, response.Code)
		}

		binexpectedLog, err := json.Marshal(expectedLog)
		assert.NoError(t, err)
		if !strings.Contains(response.Body.String(), string(binexpectedLog)) {
			t.Errorf("Response should be %v, but it answer with %v ", expectedLog, response.Body.String())
		}

	})

	t.Run("Should return error signature invalid", func(t *testing.T) {
		requestURL := testURL
		body := models.IntegrationsRequest{
			ID:        "id",
			Timestamp: "1234556",
			Type:      "text",
			From:      "5555555555",
			Text: models.Text{
				Body: "Hello",
			},
		}

		binBody, err := json.Marshal(body)
		assert.NoError(t, err)

		req, _ := http.NewRequest("POST", requestURL, bytes.NewBuffer(binBody))
		req.Header.Add("x-yalochat-signature", "error")
		response := httptest.NewRecorder()
		expectedLog := helpers.FailedResponse{
			ErrorDescription: "x-yalochat-signature invalid, header invalid.",
		}

		handler.ServeHTTP(response, req)

		if response.Code != http.StatusUnauthorized {
			t.Errorf("Response should be %v, but it answer with %v ", http.StatusUnauthorized, response.Code)
		}

		binexpectedLog, err := json.Marshal(expectedLog)
		assert.NoError(t, err)
		if !strings.Contains(response.Body.String(), string(binexpectedLog)) {
			t.Errorf("Response should be %v, but it answer with %v ", expectedLog, response.Body.String())
		}

	})*/

	t.Run("Should return error validate payload", func(t *testing.T) {
		requestURL := testURL
		body := models.IntegrationsRequest{
			ID:   "id",
			Type: "text",
			From: "5555555555",
			Text: models.Text{
				Body: "Hello",
			},
		}

		binBody, err := json.Marshal(body)
		assert.NoError(t, err)

		req, _ := http.NewRequest("POST", requestURL, bytes.NewBuffer(binBody))
		req.Header.Add("x-yalochat-signature", "secret")
		response := httptest.NewRecorder()
		expectedLog := helpers.FailedResponse{
			ErrorDescription: "Error validating payload : Key: 'IntegrationsRequest.Timestamp' Error:Field validation for 'Timestamp' failed on the 'required' tag",
		}
		binexpectedLog, err := json.Marshal(expectedLog)
		assert.NoError(t, err)

		handler.ServeHTTP(response, req)

		if response.Code != http.StatusBadRequest {
			t.Errorf("Response should be %v, but it answer with %v ", http.StatusBadRequest, response.Code)
		}

		if !strings.Contains(response.Body.String(), string(binexpectedLog)) {
			t.Errorf("Response should be %v, but it answer with %v ", expectedLog, response.Body.String())
		}

	})
}

func TestWebhookFB(t *testing.T) {
	handler := ddrouter.New(ddrouter.WithServiceName("salesforce-integration.http"))
	handler.POST(fmt.Sprintf("%s/integrations_fb/webhook", apiVersion), app.webhookFB)

	t.Run("Should save context", func(t *testing.T) {
		managerMock := new(ManagerI)

		interconnection := &models.IntegrationsFacebook{
			AuthorRole: "user",
			BotID:      botId,
			Message: models.Message{
				Entry: []models.Entry{
					{
						ID: "id",
						Messaging: []models.Messaging{
							{
								Message: models.MessagingMessage{
									Text: "text",
								},
							},
						},
					},
				},
			},
			MsgTracking: models.MsgTracking{},
			Provider:    "facebook",
			Timestamp:   123,
		}
		interconectionBin, err := json.Marshal(interconnection)
		assert.NoError(t, err)
		managerMock.On("SaveContextFB", interconnection).Return(nil).Once()
		getApp().ManageManager = managerMock

		body := interconectionBin
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/integrations_fb/webhook", apiVersion), bytes.NewBuffer(body))
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", yaloTokenTest))
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		logrus.Infof("Response : %s", response.Body.String())

		if response.Code != http.StatusOK {
			t.Errorf("Response should be %v, but it answer with %v ", http.StatusOK, response.Code)
		}
	})

	t.Run("Should save contextError Payload", func(t *testing.T) {
		managerMock := new(ManagerI)

		interconnection := &models.IntegrationsFacebook{
			AuthorRole: "user",
			Message: models.Message{
				Entry: []models.Entry{
					{
						ID: "id",
						Messaging: []models.Messaging{
							{
								Message: models.MessagingMessage{
									Text: "text",
								},
							},
						},
					},
				},
			},
			MsgTracking: models.MsgTracking{},
			Provider:    "facebook",
			Timestamp:   123,
		}
		interconectionBin, err := json.Marshal(interconnection)
		assert.NoError(t, err)
		managerMock.On("SaveContextFB", interconnection).Return(nil).Once()
		getApp().ManageManager = managerMock

		body := interconectionBin
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/integrations_fb/webhook", apiVersion), bytes.NewBuffer(body))
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", yaloTokenTest))
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)

		expectedLog := helpers.FailedResponse{
			ErrorDescription: "Error validating payload : Key: 'IntegrationsFacebook.BotID' Error:Field validation for 'BotID' failed on the 'required' tag",
		}
		binexpectedLog, err := json.Marshal(expectedLog)
		assert.NoError(t, err)

		if response.Code != http.StatusBadRequest {
			t.Errorf("Response should be %v, but it answer with %v ", http.StatusBadRequest, response.Code)
		}

		if !strings.Contains(response.Body.String(), string(binexpectedLog)) {
			t.Errorf("Response should be %v, but it answer with %v ", expectedLog, response.Body.String())
		}
	})
}
