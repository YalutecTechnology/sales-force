package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sirupsen/logrus"
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
)

func TestCreateChat(t *testing.T) {
	handler := ddrouter.New(ddrouter.WithServiceName("salesforce-integration.http"))
	handler.POST(fmt.Sprintf("%s/chats/connect", apiVersion), app.createChat)

	t.Run("Should get a valid response with valid line and agent name", func(t *testing.T) {
		managerMock := new(ManagerI)

		interconnection := &manage.Interconnection{
			UserID:   "5217331175599",
			BotSlug:  "coppel-bot",
			BotID:    "521554578545",
			Name:     "Eduardo Ochoa",
			Provider: "whatsapp",
			Email:    "ochoapumas@gmail.com",
		}
		managerMock.On("CreateChat", interconnection).Return(nil).Once()
		getApp().ManageManager = managerMock

		body := `{"userID":"5217331175599","botSlug":"coppel-bot","botId":"521554578545","name":"Eduardo Ochoa","provider":"whatsapp","email":"ochoapumas@gmail.com"}`
		req, _ := http.NewRequest("POST", requestURL, bytes.NewBuffer([]byte(body)))
		req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", yaloTokenTest))
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		logrus.Infof("Response : %s", response.Body.String())

		if response.Code != http.StatusOK {
			t.Errorf("Response should be %v, but it answer with %v ", http.StatusOK, response.Code)
		}
	})

}
