package handlers

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	ddrouter "gopkg.in/DataDog/dd-trace-go.v1/contrib/julienschmidt/httprouter"
	"yalochat.com/salesforce-integration/app/manage"
	"yalochat.com/salesforce-integration/app/services"
	"yalochat.com/salesforce-integration/base/cache"
	"yalochat.com/salesforce-integration/base/clients/chat"
	"yalochat.com/salesforce-integration/base/clients/login"
	"yalochat.com/salesforce-integration/base/clients/proxy"
	"yalochat.com/salesforce-integration/base/clients/salesforce"
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
	m, s := cache.CreateRedisServer()
	defer m.Close()
	defer s.Close()
	config := &manage.ManagerOptions{
		AppName:           "salesforce-integration",
		SfcOrganizationId: organizationId,
		SfcDeploymentId:   deploymentId,
		SfcButtonId:       buttonId,
		RedisOptions: cache.RedisOptions{
			FailOverOptions: &redis.FailoverOptions{
				MasterName:    s.MasterInfo().Name,
				SentinelAddrs: []string{s.Addr()},
			},
			SessionsTTL: time.Second,
		},
	}
	API(handler, config, apiConfig)
	mock := &proxy.Mock{}
	mock2 := &proxy.Mock{}
	salesforceServiceMock := services.NewSalesforceService(login.SfcLoginClient{}, chat.SfcChatClient{}, salesforce.SalesforceClient{})
	salesforceServiceMock.SfcClient.Proxy = mock
	salesforceServiceMock.SfcChatClient.Proxy = mock2

	getApp().ManageManager.SalesforceService = salesforceServiceMock
	mock.On("SendHTTPRequest").Return(&http.Response{
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"totalSize":0,"done":true,"records":[]}`))),
	}, nil).Once().On("SendHTTPRequest").Return(&http.Response{
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"totalSize":0,"done":true,"records":[]}`))),
	}, nil).Once().On("SendHTTPRequest").Return(&http.Response{
		StatusCode: http.StatusCreated,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd","success":true,"errors":[]}`))),
	}, nil).Once()
	sessionResponse := `{"key":"ec550263-354e-477c-b773-7747ebce3f5e!1629334776994!TrfoJ67wmtlYiENsWdaUBu0xZ7M=","id":"ec550263-354e-477c-b773-7747ebce3f5e","clientPollTimeout":40,"affinityToken":"878a1fa0"}`
	mock2.On("SendHTTPRequest").Return(&http.Response{
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte(sessionResponse))),
	}, nil).Once().On("SendHTTPRequest").Return(&http.Response{
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte(sessionResponse))),
	}, nil).Once().On("SendHTTPRequest").Return(&http.Response{
		StatusCode: http.StatusOK,
		Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"messages":[{"type":"ChatRequestFail","message":{"geoLocation":{}}}]}`))),
	}, nil).Once()

	t.Run("Should get a valid response with valid line and agent name", func(t *testing.T) {
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
