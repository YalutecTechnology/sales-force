package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	ddrouter "gopkg.in/DataDog/dd-trace-go.v1/contrib/julienschmidt/httprouter"
	"yalochat.com/salesforce-integration/base/cache"
	"yalochat.com/salesforce-integration/base/helpers"
	"yalochat.com/salesforce-integration/base/models"
)

func TestApp_webhook(t *testing.T) {
	handler := ddrouter.New(ddrouter.WithServiceName("salesforce-integration.http"))
	url := fmt.Sprintf("%s/integrations/whatsapp/webhook", apiVersion)
	handler.POST(url, app.webhook)

	t.Run("Should return success", func(t *testing.T) {
		managerMock := new(ManagerI)
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

		managerMock.On("SaveContext", &body).Return(nil).Once()
		getApp().ManageManager = managerMock

		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(binBody))
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

	t.Run("Should return error validate payload", func(t *testing.T) {
		managerMock := new(ManagerI)
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

		managerMock.On("SaveContext", &body).Return(nil).Once()
		getApp().ManageManager = managerMock

		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(binBody))
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

	t.Run("Should return error payload decode", func(t *testing.T) {
		managerMock := new(ManagerI)

		getApp().ManageManager = managerMock

		req, _ := http.NewRequest("POST", url, bytes.NewBuffer([]byte("error")))
		req.Header.Add("x-yalochat-signature", "secret")
		response := httptest.NewRecorder()
		expectedLog := helpers.FailedResponse{
			ErrorDescription: "Invalid payload received : invalid character 'e' looking for beginning of value",
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

	t.Run("Should return error manage", func(t *testing.T) {
		managerMock := new(ManagerI)
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

		managerMock.On("SaveContext", &body).Return(assert.AnError).Once()
		getApp().ManageManager = managerMock

		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(binBody))
		req.Header.Add("x-yalochat-signature", "secret")
		response := httptest.NewRecorder()
		expectedLog := helpers.FailedResponse{
			ErrorDescription: "There was an error inserting integration message: assert.AnError general error for testing",
		}
		binexpectedLog, err := json.Marshal(expectedLog)
		assert.NoError(t, err)

		handler.ServeHTTP(response, req)

		if response.Code != http.StatusNotFound {
			t.Errorf("Response should be %v, but it answer with %v ", http.StatusNotFound, response.Code)
		}

		if !strings.Contains(response.Body.String(), string(binexpectedLog)) {
			t.Errorf("Response should be %v, but it answer with %v ", expectedLog, response.Body.String())
		}

	})
}

func TestWebhookFB(t *testing.T) {
	handler := ddrouter.New(ddrouter.WithServiceName("salesforce-integration.http"))
	url := fmt.Sprintf("%s/integrations/facebook/webhook", apiVersion)
	handler.POST(url, app.webhookFB)

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
		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
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

		getApp().ManageManager = managerMock

		body := interconectionBin
		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
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

	t.Run("Should save contextError Payload decode", func(t *testing.T) {
		managerMock := new(ManagerI)

		getApp().ManageManager = managerMock

		req, _ := http.NewRequest("POST", url, bytes.NewBuffer([]byte("error")))
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", yaloTokenTest))
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)

		expectedLog := helpers.FailedResponse{
			ErrorDescription: "Invalid payload received : invalid character 'e' looking for beginning of value",
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
		managerMock.On("SaveContextFB", interconnection).Return(assert.AnError).Once()
		getApp().ManageManager = managerMock

		body := interconectionBin
		req, _ := http.NewRequest("POST", url, bytes.NewBuffer(body))
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", yaloTokenTest))
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)

		expectedLog := helpers.FailedResponse{
			ErrorDescription: "There was an error inserting integration message: assert.AnError general error for testing",
		}
		binexpectedLog, err := json.Marshal(expectedLog)
		assert.NoError(t, err)

		if response.Code != http.StatusNotFound {
			t.Errorf("Response should be %v, but it answer with %v ", http.StatusNotFound, response.Code)
		}

		if !strings.Contains(response.Body.String(), string(binexpectedLog)) {
			t.Errorf("Response should be %v, but it answer with %v ", expectedLog, response.Body.String())
		}
	})
}

func TestGetContext(t *testing.T) {
	handler := ddrouter.New(ddrouter.WithServiceName("salesforce-integration.http"))
	url := fmt.Sprintf("%s/context/:user_id", apiVersion)

	urlTest := fmt.Sprintf("%s/context/%s", apiVersion, userID)
	handler.GET(url, app.getContext)

	t.Run("Should save context", func(t *testing.T) {
		managerMock := new(ManagerI)

		expected := []cache.Context{
			{
				UserID:    userID,
				Timestamp: 111111,
				Text:      "test",
			},
		}
		managerMock.On("GetContextByUserID", userID).Return(expected).Once()
		getApp().ManageManager = managerMock

		req, _ := http.NewRequest("GET", urlTest, nil)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", yaloTokenTest))
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		logrus.Infof("Response : %s", response.Body.String())

		expectedBin, err := json.Marshal(expected)
		assert.NoError(t, err)
		if response.Code != http.StatusOK {
			t.Errorf("Response should be %v, but it answer with %v ", http.StatusOK, response.Code)
		}

		assert.Equal(t, string(expectedBin), response.Body.String())
	})
}
