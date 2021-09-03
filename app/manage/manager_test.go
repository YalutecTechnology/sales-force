package manage

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
	"yalochat.com/salesforce-integration/app/services"
	"yalochat.com/salesforce-integration/base/cache"
	"yalochat.com/salesforce-integration/base/clients/chat"
	"yalochat.com/salesforce-integration/base/clients/login"
	"yalochat.com/salesforce-integration/base/clients/proxy"
	"yalochat.com/salesforce-integration/base/clients/salesforce"
	"yalochat.com/salesforce-integration/base/models"
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
)

func TestCreateManager(t *testing.T) {
	m, s := cache.CreateRedisServer()
	defer m.Close()
	defer s.Close()
	t.Run("Should retrieve a manager instance", func(t *testing.T) {
		expected := &Manager{
			clientName:         "salesforce-integration",
			interconnectionMap: make(map[string]*models.Interconnection),
			sessionMap:         make(map[string]string),
			sfcContactMap:      make(map[string]*models.SfcContact),
			SalesforceService:  nil,
		}
		config := &ManagerOptions{
			AppName: "salesforce-integration",
			RedisOptions: cache.RedisOptions{
				FailOverOptions: &redis.FailoverOptions{
					MasterName:    s.MasterInfo().Name,
					SentinelAddrs: []string{s.Addr()},
				},
				SessionsTTL: time.Second,
			},
		}
		actual := CreateManager(config)
		actual.SalesforceService = nil

		assert.Equal(t, expected, actual)
	})
}

func TestSalesforceService_CreateChat(t *testing.T) {
	m, s := cache.CreateRedisServer()
	defer m.Close()
	defer s.Close()
	t.Run("Create Chat Succesfull", func(t *testing.T) {
		interconnection := &models.Interconnection{
			UserId:      userId,
			BotSlug:     botSlug,
			BotId:       botId,
			Name:        name,
			Provider:    provider,
			Email:       email,
			PhoneNumber: phoneNumber,
		}
		config := &ManagerOptions{
			AppName: "salesforce-integration",
			RedisOptions: cache.RedisOptions{
				FailOverOptions: &redis.FailoverOptions{
					MasterName:    s.MasterInfo().Name,
					SentinelAddrs: []string{s.Addr()},
				},
				SessionsTTL: time.Second,
			},
		}
		manager := CreateManager(config)
		SfcOrganizationId = organizationId
		SfcDeploymentId = deploymentId
		SfcButtonId = buttonId
		mock := &proxy.Mock{}
		mock2 := &proxy.Mock{}
		salesforceServiceMock := services.NewSalesforceService(login.SfcLoginClient{}, chat.SfcChatClient{}, salesforce.SalesforceClient{})
		salesforceServiceMock.SfcClient.Proxy = mock
		salesforceServiceMock.SfcChatClient.Proxy = mock2
		manager.SalesforceService = salesforceServiceMock
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
		}, nil).Once()

		err := manager.CreateChat(interconnection)
		assert.NoError(t, err)
	})

}

func TestManager_SaveContext(t *testing.T) {

	t.Run("Should save context voice", func(t *testing.T) {
		contextCache := new(ContextCacheMock)
		ctx := cache.Context{
			UserID:    "55555555555",
			Timestamp: 123456789,
			URL:       "uri",
			MIMEType:  "voice",
			Caption:   "caption",
		}
		contextCache.On("StoreContext", ctx).Return(nil)

		manager := &Manager{
			contextCache: contextCache,
		}

		integrations := &models.IntegrationsRequest{
			ID:        "id",
			Timestamp: "123456789",
			Type:      voiceType,
			From:      "55555555555",
			Voice: models.Media{
				URL:      "uri",
				MIMEType: "voice",
				Caption:  "caption",
			},
		}
		err := manager.SaveContext(integrations)

		assert.NoError(t, err)
	})

	t.Run("Should save context document", func(t *testing.T) {
		contextCache := new(ContextCacheMock)
		ctx := cache.Context{
			UserID:    "55555555555",
			Timestamp: 123456789,
			URL:       "uri",
			MIMEType:  "document",
			Caption:   "caption",
		}
		contextCache.On("StoreContext", ctx).Return(nil)

		manager := &Manager{
			contextCache: contextCache,
		}

		integrations := &models.IntegrationsRequest{
			ID:        "id",
			Timestamp: "123456789",
			Type:      documentType,
			From:      "55555555555",
			Document: models.Media{
				URL:      "uri",
				MIMEType: "document",
				Caption:  "caption",
			},
		}
		err := manager.SaveContext(integrations)

		assert.NoError(t, err)
	})

	t.Run("Should save context document", func(t *testing.T) {
		contextCache := new(ContextCacheMock)
		ctx := cache.Context{
			UserID:    "55555555555",
			Timestamp: 123456789,
			URL:       "uri",
			MIMEType:  "image",
			Caption:   "caption",
		}
		contextCache.On("StoreContext", ctx).Return(nil)

		manager := &Manager{
			contextCache: contextCache,
		}

		integrations := &models.IntegrationsRequest{
			ID:        "id",
			Timestamp: "123456789",
			Type:      imageType,
			From:      "55555555555",
			Image: models.Media{
				URL:      "uri",
				MIMEType: "image",
				Caption:  "caption",
			},
		}
		err := manager.SaveContext(integrations)

		assert.NoError(t, err)
	})

	t.Run("Should save context text", func(t *testing.T) {
		contextCache := new(ContextCacheMock)
		ctx := cache.Context{
			UserID:    "55555555555",
			Timestamp: 123456789,
			Text:      "text",
		}
		contextCache.On("StoreContext", ctx).Return(nil)

		manager := &Manager{
			contextCache: contextCache,
		}

		integrations := &models.IntegrationsRequest{
			ID:        "id",
			Timestamp: "123456789",
			Type:      textType,
			From:      "55555555555",
			Text: models.Text{
				Body: "text",
			},
		}
		err := manager.SaveContext(integrations)

		assert.NoError(t, err)
	})

	t.Run("Should save context error", func(t *testing.T) {
		contextCache := new(ContextCacheMock)
		ctx := cache.Context{
			UserID:    "55555555555",
			Timestamp: 123456789,
			Text:      "text",
		}
		contextCache.On("StoreContext", ctx).Return(assert.AnError)

		manager := &Manager{
			contextCache: contextCache,
		}

		integrations := &models.IntegrationsRequest{
			ID:        "id",
			Timestamp: "123456789",
			Type:      textType,
			From:      "55555555555",
			Text: models.Text{
				Body: "text",
			},
		}
		err := manager.SaveContext(integrations)

		assert.Error(t, err)
	})
}
