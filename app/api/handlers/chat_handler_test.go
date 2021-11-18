package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	ddrouter "gopkg.in/DataDog/dd-trace-go.v1/contrib/julienschmidt/httprouter"
	"yalochat.com/salesforce-integration/app/manage"
)

const (
	name           = "username"
	email          = "user@exmple.com"
	userId         = "userId"
	botSlug        = "coppel-bot"
	botId          = "5514254524"
	provider       = "whatsapp"
	phoneNumber    = "5512454545"
	organizationId = "organizationId"
	deploymentId   = "deploymentId"
	buttonId       = "buttonID"
	requestURL     = "/v1/chats/connect"
	reqFinishURL   = "/v1/chat/finish/5217331175599"
	userID         = "5217331175599"
)

func TestCreateChat(t *testing.T) {
	handler := ddrouter.New(ddrouter.WithServiceName("salesforce-integration.http"))
	handler.POST(requestURL, app.createChat)

	t.Run("Should get a valid response", func(t *testing.T) {
		managerMock := new(ManagerI)

		interconnection := &manage.Interconnection{
			UserID:      "5217331175599",
			BotSlug:     "coppel-bot",
			BotID:       "521554578545",
			Name:        "Eduardo Ochoa",
			Provider:    "whatsapp",
			Email:       "ochoapumas@gmail.com",
			PhoneNumber: "55555555555",
		}
		managerMock.On("CreateChat", mock.Anything, interconnection).Return(nil).Once()
		getApp().ManageManager = managerMock

		interconnectionBin, err := json.Marshal(interconnection)
		assert.NoError(t, err)

		req, _ := http.NewRequest("POST", requestURL, bytes.NewBuffer(interconnectionBin))
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", yaloTokenTest))
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		logrus.Infof("Response : %s", response.Body.String())

		if response.Code != http.StatusOK {
			t.Errorf("Response should be %v, but it answer with %v ", http.StatusOK, response.Code)
		}
	})

	t.Run("Should get a valid response with payload error", func(t *testing.T) {
		managerMock := new(ManagerI)

		interconnection := &manage.Interconnection{
			BotSlug:     "coppel-bot",
			BotID:       "521554578545",
			Name:        "Eduardo Ochoa",
			Provider:    "whatsapp",
			Email:       "ochoapumas@gmail.com",
			PhoneNumber: "55555555555",
		}
		getApp().ManageManager = managerMock

		interconnectionBin, err := json.Marshal(interconnection)
		assert.NoError(t, err)

		var buf bytes.Buffer
		logrus.SetOutput(&buf)

		req, _ := http.NewRequest("POST", requestURL, bytes.NewBuffer(interconnectionBin))
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", yaloTokenTest))
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		logrus.Infof("Response : %s", response.Body.String())

		if response.Code != http.StatusBadRequest {
			t.Errorf("Response should be %v, but it answer with %v ", http.StatusBadRequest, response.Code)
		}
		assert.Equal(t,
			`{"ErrorDescription":"Error validating payload : Key: 'ChatPayload.UserID' Error:Field validation for 'UserID' failed on the 'required' tag"}`,
			response.Body.String())

		logs := buf.String()
		expectedLog := "Error validating payload : Key: 'ChatPayload.UserID' Error:Field validation for 'UserID' failed on the 'required' tag"
		if !strings.Contains(logs, expectedLog) {
			t.Fatalf("Logs should contain <%s>, but this was found <%s>", expectedLog, logs)
		}
	})

	t.Run("Should get a valid response with payload encode error", func(t *testing.T) {
		managerMock := new(ManagerI)

		getApp().ManageManager = managerMock
		var buf bytes.Buffer
		logrus.SetOutput(&buf)

		req, _ := http.NewRequest("POST", requestURL, bytes.NewBuffer([]byte("error")))
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", yaloTokenTest))
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		logrus.Infof("Response : %s", response.Body.String())

		if response.Code != http.StatusBadRequest {
			t.Errorf("Response should be %v, but it answer with %v ", http.StatusBadRequest, response.Code)
		}
		assert.Equal(t,
			`{"ErrorDescription":"Invalid payload received : invalid character 'e' looking for beginning of value"}`,
			response.Body.String())

		logs := buf.String()
		expectedLog := "Invalid payload received : invalid character 'e' looking for beginning of value"
		if !strings.Contains(logs, expectedLog) {
			t.Fatalf("Logs should contain <%s>, but this was found <%s>", expectedLog, logs)
		}
	})

	t.Run("Should get a error manage service", func(t *testing.T) {
		managerMock := new(ManagerI)

		interconnection := &manage.Interconnection{
			UserID:      "5217331175599",
			BotSlug:     "coppel-bot",
			BotID:       "521554578545",
			Name:        "Eduardo Ochoa",
			Provider:    "whatsapp",
			Email:       "ochoapumas@gmail.com",
			PhoneNumber: "55555555555",
		}
		managerMock.On("CreateChat", mock.Anything, interconnection).Return(assert.AnError).Once()
		getApp().ManageManager = managerMock

		interconnectionBin, err := json.Marshal(interconnection)
		assert.NoError(t, err)

		var buf bytes.Buffer
		logrus.SetOutput(&buf)

		req, _ := http.NewRequest("POST", requestURL, bytes.NewBuffer(interconnectionBin))
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", yaloTokenTest))
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		logrus.Infof("Response : %s", response.Body.String())

		if response.Code != http.StatusNotFound {
			t.Errorf("Response should be %v, but it answer with %v ", http.StatusNotFound, response.Code)
		}
		assert.Equal(t,
			`{"ErrorDescription":"assert.AnError general error for testing"}`,
			response.Body.String())

		logs := buf.String()
		expectedLog := "assert.AnError general error for testing"
		if !strings.Contains(logs, expectedLog) {
			t.Fatalf("Logs should contain <%s>, but this was found <%s>", expectedLog, logs)
		}
	})

}

func TestFinishChat(t *testing.T) {
	handler := ddrouter.New(ddrouter.WithServiceName("salesforce-integration.http"))
	handler.DELETE(fmt.Sprintf("%s/chat/finish/:user_id", apiVersion), app.finishChat)

	t.Run("You must close the user side chat successfully", func(t *testing.T) {
		managerMock := new(ManagerI)

		managerMock.On("FinishChat", userID).Return(nil).Once()
		getApp().ManageManager = managerMock

		req, _ := http.NewRequest("DELETE", reqFinishURL, nil)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", yaloTokenTest))
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		logrus.Infof("Response : %s", response.Body.String())
		if response.Code != http.StatusOK {
			t.Errorf("Response should be %v, but it answer with %v ", http.StatusOK, response.Code)
		}
	})

	t.Run("You must close the user side chat error service", func(t *testing.T) {
		managerMock := new(ManagerI)

		managerMock.On("FinishChat", userID).Return(assert.AnError).Once()
		getApp().ManageManager = managerMock

		req, _ := http.NewRequest("DELETE", reqFinishURL, nil)
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", yaloTokenTest))
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		logrus.Infof("Response : %s", response.Body.String())
		if response.Code != http.StatusNotFound {
			t.Errorf("Response should be %v, but it answer with %v ", http.StatusNotFound, response.Code)
		}
		assert.Equal(t,
			`{"ErrorDescription":"assert.AnError general error for testing"}`,
			response.Body.String())
	})
}
