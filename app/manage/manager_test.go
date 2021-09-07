package manage

import (
	"strings"
	"testing"
	"time"

	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"

	"yalochat.com/salesforce-integration/base/cache"
	"yalochat.com/salesforce-integration/base/clients/chat"
	"yalochat.com/salesforce-integration/base/models"
)

const (
	name             = "username"
	email            = "user@exmple.com"
	botSlug          = "coppel-bot"
	botID            = "5514254524"
	provider         = "whatsapp"
	phoneNumber      = "5512454545"
	organizationID   = "organizationID"
	deploymentID     = "deploymentID"
	buttonID         = "buttonID"
	blockedUserState = "from-sf-blocked"
	contactID        = "contactID"
	recordTypeID     = "recordTypeID"
	caseID           = "caseID"
)

func TestCreateManager(t *testing.T) {
	m, s := cache.CreateRedisServer()
	defer m.Close()
	defer s.Close()
	t.Run("Should retrieve a manager instance", func(t *testing.T) {
		expected := &Manager{
			clientName:         "salesforce-integration",
			interconnectionMap: interconnectionCache{interconnections: make(interconnectionMap)},
			SalesforceService:  nil,
			IntegrationsClient: nil,
			BotrunnnerClient:   nil,
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
		actual.IntegrationsClient = nil
		actual.BotrunnnerClient = nil
		expected.integrationsChannel = actual.integrationsChannel
		expected.salesforceChannel = actual.salesforceChannel
		expected.finishInterconnection = actual.finishInterconnection
		expected.cache = actual.cache

		assert.Equal(t, expected, actual)
	})
}

func TestSalesforceService_CreateChat(t *testing.T) {
	m, s := cache.CreateRedisServer()
	defer m.Close()
	defer s.Close()
	t.Run("Create Chat Succesfull", func(t *testing.T) {
		interconnection := &Interconnection{
			UserID:      userID,
			BotSlug:     botSlug,
			BotID:       botID,
			Name:        name,
			Provider:    provider,
			Email:       email,
			PhoneNumber: phoneNumber,
			ExtraData:   map[string]interface{}{"data": "datavalue"},
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
		SfcOrganizationID = organizationID
		SfcDeploymentID = deploymentID
		SfcButtonID = buttonID
		SfcRecordTypeID = recordTypeID
		SfcCustomFieldsCase = []string{"data:data"}
		salesforceMock := new(SalesforceServiceInterface)

		contact := &models.SfcContact{
			Id:          contactID,
			FirstName:   interconnection.Name,
			Email:       interconnection.Email,
			MobilePhone: interconnection.PhoneNumber,
		}
		salesforceMock.On("GetOrCreateContact",
			interconnection.Name,
			interconnection.Email,
			interconnection.PhoneNumber).
			Return(contact, nil).Once()

		salesforceMock.On("CreatCase",
			SfcRecordTypeID,
			contact.Id,
			"Caso levantado por el Bot : ",
			string(interconnection.Provider),
			interconnection.ExtraData,
			SfcCustomFieldsCase).
			Return(caseID, nil).Once()

		salesforceMock.On("CreatChat",
			interconnection.Name,
			SfcOrganizationID,
			SfcDeploymentID,
			SfcButtonID,
			caseID,
			contactID).
			Return(&chat.SessionResponse{
				AffinityToken: affinityToken,
				Key:           sessionKey,
			}, nil).Once()

		salesforceMock.On("GetMessages",
			affinityToken, sessionKey).
			Return(&chat.MessagesResponse{}, nil).Once()

		manager.SalesforceService = salesforceMock

		err := manager.CreateChat(interconnection)
		assert.NoError(t, err)
	})

	t.Run("Change to from-sf-blocked state succesfull", func(t *testing.T) {
		expectedLog := "could not create chat in salesforce"
		interconnection := &Interconnection{
			UserID:      userID,
			BotSlug:     botSlug,
			BotID:       botID,
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
		SfcOrganizationID = organizationID
		SfcDeploymentID = deploymentID
		SfcButtonID = buttonID
		BlockedUserState = "from-sf-blocked"
		contact := &models.SfcContact{
			FirstName:   interconnection.Name,
			LastName:    interconnection.Name,
			Email:       interconnection.Email,
			MobilePhone: interconnection.PhoneNumber,
			Blocked:     true,
		}

		salesforceServiceMock := new(SalesforceServiceInterface)
		botRunnerMock := new(BotRunnerInterface)

		salesforceServiceMock.On("GetOrCreateContact",
			interconnection.Name,
			interconnection.Email,
			interconnection.PhoneNumber).
			Return(contact, nil).Once()
		botRunnerMock.On("SendTo", map[string]interface{}{"botSlug": botSlug, "message": "", "state": BlockedUserState, "userId": userID}).
			Return(true, nil).Once()
		manager.SalesforceService = salesforceServiceMock
		manager.BotrunnnerClient = botRunnerMock
		err := manager.CreateChat(interconnection)
		assert.Error(t, err)

		if !strings.Contains(err.Error(), expectedLog) {
			t.Fatalf("Error message should contain %s, but this was found <%s>", expectedLog, err.Error())
		}
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
			From:      fromUser,
		}
		contextCache.On("StoreContext", ctx).Return(nil).Once()

		manager := &Manager{
			cache: contextCache,
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
			From:      fromBot,
		}
		contextCache.On("StoreContext", ctx).Return(nil)

		manager := &Manager{
			cache: contextCache,
		}

		integrations := &models.IntegrationsRequest{
			ID:        "id",
			Timestamp: "123456789",
			Type:      documentType,
			To:        "55555555555",
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
			From:      fromUser,
		}
		contextCache.On("StoreContext", ctx).Return(nil)

		manager := &Manager{
			cache: contextCache,
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
			From:      fromUser,
		}
		contextCache.On("StoreContext", ctx).Return(nil)

		manager := &Manager{
			cache: contextCache,
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
			From:      fromUser,
		}
		contextCache.On("StoreContext", ctx).Return(assert.AnError)

		manager := &Manager{
			cache: contextCache,
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

	t.Run("Should save context default", func(t *testing.T) {
		contextCache := new(ContextCacheMock)
		ctx := cache.Context{
			UserID:    "55555555555",
			Timestamp: 123456789,
			Text:      "text",
			From:      fromUser,
		}
		contextCache.On("StoreContext", ctx).Return(assert.AnError)

		manager := &Manager{
			cache: contextCache,
		}

		integrations := &models.IntegrationsRequest{
			ID:        "id",
			Timestamp: "123456789",
			Type:      "error",
			From:      "55555555555",
			Text: models.Text{
				Body: "text",
			},
		}
		err := manager.SaveContext(integrations)

		assert.Error(t, err)
	})

	t.Run("Should save context error timestamp", func(t *testing.T) {
		contextCache := new(ContextCacheMock)
		ctx := cache.Context{
			UserID:    "55555555555",
			Timestamp: 123456789,
			Text:      "text",
			From:      fromUser,
		}
		contextCache.On("StoreContext", ctx).Return(assert.AnError)

		manager := &Manager{
			cache: contextCache,
		}

		integrations := &models.IntegrationsRequest{
			ID:        "id",
			Timestamp: "error",
			Type:      "error",
			From:      "55555555555",
			Text: models.Text{
				Body: "text",
			},
		}
		err := manager.SaveContext(integrations)

		assert.Error(t, err)
	})

	t.Run("Should send message to salesforce", func(t *testing.T) {
		contextCache := new(ContextCacheMock)
		salesforceMock := new(SalesforceServiceInterface)
		salesforceMock.On("SendMessage",
			affinityToken, sessionKey, chat.MessagePayload{Text: "message"}).
			Return(false, nil).Once()

		manager := &Manager{
			cache: contextCache,
			interconnectionMap: interconnectionCache{
				interconnections: interconnectionMap{
					"55555555555": &Interconnection{
						Status:        Active,
						AffinityToken: affinityToken,
						SessionKey:    sessionKey,
					},
				},
			},
			SalesforceService: salesforceMock,
		}

		integrations := &models.IntegrationsRequest{
			ID:        "id",
			Timestamp: "123456789",
			Type:      textType,
			From:      "55555555555",
			Text: models.Text{
				Body: "message",
			},
		}
		err := manager.SaveContext(integrations)

		assert.NoError(t, err)
	})
}

func TestManager_getContextByUserID(t *testing.T) {
	t.Run("Should save context voice", func(t *testing.T) {
		contextCache := new(ContextCacheMock)
		ctx := []cache.Context{
			{
				UserID:    "55555555555",
				Timestamp: 1630404180000,
				Text:      "this a test",
				From:      fromUser,
			},
			{
				UserID:    "55555555555",
				Timestamp: 1630404000000,
				Text:      "Hello",
				From:      fromUser,
			},
			{
				UserID:    "55555555555",
				Timestamp: 1630404060000,
				Text:      "Hello I'm a bot",
				From:      fromBot,
			},
			{
				UserID:    "55555555555",
				Timestamp: 1630404240000,
				Text:      "ok.",
				From:      fromBot,
			},
			{
				UserID:    "55555555555",
				Timestamp: 1630404120000,
				Text:      "I need help",
				From:      fromUser,
			},
		}

		userID := "userID"
		contextCache.On("RetrieveContext", userID).Return(ctx)

		manager := &Manager{
			cache: contextCache,
		}

		ctxStr := manager.GetContextByUserID(userID)
		expected := `Cliente [31-08-2021 05:00:00]:Hello
Bot [31-08-2021 05:01:00]:Hello I'm a bot
Cliente [31-08-2021 05:02:00]:I need help
Cliente [31-08-2021 05:03:00]:this a test
Bot [31-08-2021 05:04:00]:ok.
`
		assert.Equal(t, expected, ctxStr)
	})
}
