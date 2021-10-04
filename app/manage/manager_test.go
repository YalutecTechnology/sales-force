package manage

import (
	"strings"
	"testing"
	"time"

	"github.com/go-redis/redis"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"yalochat.com/salesforce-integration/app/config/envs"
	"yalochat.com/salesforce-integration/base/cache"
	"yalochat.com/salesforce-integration/base/clients/chat"
	"yalochat.com/salesforce-integration/base/clients/integrations"
	"yalochat.com/salesforce-integration/base/models"
)

const (
	name                = "username"
	email               = "user@exmple.com"
	botSlug             = "coppel-bot"
	botID               = "5514254524"
	provider            = "whatsapp"
	phoneNumber         = "5512454545"
	organizationID      = "organizationID"
	deploymentID        = "deploymentID"
	blockedUserState    = "from-sf-blocked"
	contactID           = "contactID"
	recordTypeID        = "recordTypeID"
	recordAccountTypeID = "recordAccountTypeID"
	caseID              = "caseID"
	messageID           = "messageID"
	ttlMessage          = time.Second * 3
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
			cacheMessage:       nil,
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
		actual.cacheMessage = nil
		expected.integrationsChannel = actual.integrationsChannel
		expected.salesforceChannel = actual.salesforceChannel
		expected.finishInterconnection = actual.finishInterconnection
		expected.contextcache = actual.contextcache
		expected.interconnectionsCache = actual.interconnectionsCache

		assert.Equal(t, expected, actual)
	})
}

func TestSalesforceService_CreateChat(t *testing.T) {
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

		SfcOrganizationID = organizationID
		SfcDeploymentID = deploymentID
		SfcRecordTypeID = recordTypeID
		SfcCustomFieldsCase = map[string]string{"data": "data"}

		salesforceMock := new(SalesforceServiceInterface)
		interconnectionMock := new(InterconnectionCache)

		contact := &models.SfcContact{
			ID:          contactID,
			FirstName:   interconnection.Name,
			Email:       interconnection.Email,
			MobilePhone: interconnection.PhoneNumber,
		}
		salesforceMock.On("GetOrCreateContact",
			interconnection.Name,
			interconnection.Email,
			interconnection.PhoneNumber,
			"").
			Return(contact, nil).Once()

		salesforceMock.On("CreatCase",
			SfcRecordTypeID,
			contact.ID,
			"Caso levantado por el Bot : ",
			"subject",
			string(interconnection.Provider),
			"ownerWAID",
			interconnection.ExtraData).
			Return(caseID, nil).Once()

		salesforceMock.On("CreatChat",
			interconnection.Name,
			SfcOrganizationID,
			SfcDeploymentID,
			"buttonWAID",
			caseID,
			contactID).Return(&chat.SessionResponse{
			AffinityToken: affinityToken,
			Key:           sessionKey,
		}, nil).Once()

		salesforceMock.On("GetMessages",
			affinityToken, sessionKey).
			Return(&chat.MessagesResponse{}, nil).Once()

		interconnectionMock.On("RetrieveInterconnectionActiveByUserId", userID).
			Return(nil, nil).Once()
		interconnectionMock.On("StoreInterconnection", mock.Anything).
			Return(nil).Once()

		cacheContextMock := new(ContextCacheMock)
		cacheContextMock.On("RetrieveContext", userID).
			Return([]cache.Context{
				{
					UserID:    userID,
					Timestamp: 1111111111,
					Text:      "text",
				},
			}).Once()
		manager := &Manager{
			SalesforceService:     salesforceMock,
			interconnectionsCache: interconnectionMock,
			contextcache:          cacheContextMock,
			interconnectionMap: interconnectionCache{
				interconnections: interconnectionMap{
					"55555555555": &Interconnection{
						Status:        Active,
						AffinityToken: affinityToken,
						SessionKey:    sessionKey,
						SessionID:     sessionID,
						UserID:        userID,
					},
				},
			},
			SfcSourceFlowBot: envs.SfcSourceFlowBot{
				defaultFieldCustom: {
					Subject: "subject",
					Providers: map[string]envs.Provider{
						"whatsapp": {
							ButtonID: "buttonWAID",
							OwnerID:  "ownerWAID",
						},
					},
				},
			},
		}

		err := manager.CreateChat(interconnection)
		assert.NoError(t, err)
		manager.EndChat(interconnection)
	})

	t.Run("Create Chat Succesfull with FB provider", func(t *testing.T) {
		interconnection := &Interconnection{
			UserID:      userID,
			BotSlug:     botSlug,
			BotID:       botID,
			Name:        name,
			Provider:    FacebookProvider,
			Email:       email,
			PhoneNumber: phoneNumber,
		}
		SfcOrganizationID = organizationID
		SfcDeploymentID = deploymentID
		salesforceMock := new(SalesforceServiceInterface)

		contact := &models.SfcContact{
			ID:          contactID,
			FirstName:   interconnection.Name,
			Email:       interconnection.Email,
			MobilePhone: interconnection.PhoneNumber,
		}
		salesforceMock.On("GetOrCreateContact",
			interconnection.Name,
			interconnection.Email,
			interconnection.PhoneNumber,
			"").
			Return(contact, nil).Once()

		salesforceMock.On("CreatCase",
			SfcRecordTypeID,
			contact.ID,
			"Caso levantado por el Bot : ",
			"subject",
			string(interconnection.Provider),
			"ownerFBID",
			interconnection.ExtraData).
			Return(caseID, nil).Once()

		salesforceMock.On("CreatChat",
			interconnection.Name,
			SfcOrganizationID,
			SfcDeploymentID,
			"buttonFBID",
			caseID,
			contactID).
			Return(&chat.SessionResponse{
				AffinityToken: affinityToken,
				Key:           sessionKey,
			}, nil).Once()

		salesforceMock.On("GetMessages",
			affinityToken, sessionKey).
			Return(&chat.MessagesResponse{}, nil).Once()

		interconnectionMock := new(InterconnectionCache)
		interconnectionMock.On("RetrieveInterconnectionActiveByUserId", userID).
			Return(nil, nil).Once()
		interconnectionMock.On("StoreInterconnection", mock.Anything).
			Return(nil).Once()

		contextMock := new(ContextCacheMock)
		contextMock.On("RetrieveContext",
			userID).
			Return([]cache.Context{
				{
					UserID:    userID,
					Timestamp: 1111111111,
					Text:      "text",
				},
			}).Once()

		manager := &Manager{
			SalesforceService:     salesforceMock,
			interconnectionsCache: interconnectionMock,
			SfcSourceFlowBot: envs.SfcSourceFlowBot{
				defaultFieldCustom: {
					Subject: "subject",
					Providers: map[string]envs.Provider{
						"facebook": {
							ButtonID: "buttonFBID",
							OwnerID:  "ownerFBID",
						},
					},
				},
			},
			contextcache: contextMock,
			interconnectionMap: interconnectionCache{
				interconnections: make(interconnectionMap),
			},
		}

		err := manager.CreateChat(interconnection)
		assert.NoError(t, err)
		manager.EndChat(interconnection)
	})

	t.Run("Create Chat Succesfull with an account", func(t *testing.T) {
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

		SfcOrganizationID = organizationID
		SfcDeploymentID = deploymentID
		SfcRecordTypeID = recordTypeID
		SfcAccountRecordTypeID = recordAccountTypeID
		//SfcCustomFieldsCase = map[string]string{"data": "data"}

		salesforceMock := new(SalesforceServiceInterface)
		interconnectionMock := new(InterconnectionCache)

		contact := &models.SfcContact{
			AccountID:   recordAccountTypeID,
			FirstName:   interconnection.Name,
			Email:       interconnection.Email,
			MobilePhone: interconnection.PhoneNumber,
			LastName:    interconnection.Name,
			ID:          contactID,
		}
		salesforceMock.On("GetOrCreateContact",
			interconnection.Name,
			interconnection.Email,
			interconnection.PhoneNumber,
			recordAccountTypeID).
			Return(contact, nil).Once()

		salesforceMock.On("CreatCase",
			SfcRecordTypeID,
			contact.ID,
			"Caso levantado por el Bot : ",
			"",
			string(interconnection.Provider),
			"",
			interconnection.ExtraData).
			Return(caseID, nil).Once()

		salesforceMock.On("CreatChat",
			interconnection.Name,
			SfcOrganizationID,
			SfcDeploymentID,
			"",
			caseID,
			contactID).Return(&chat.SessionResponse{
			AffinityToken: affinityToken,
			Key:           sessionKey,
		}, nil).Once()

		salesforceMock.On("GetMessages",
			affinityToken, sessionKey).
			Return(&chat.MessagesResponse{}, nil).Once()

		interconnectionMock.On("RetrieveInterconnectionActiveByUserId", userID).
			Return(nil, nil).Once()
		interconnectionMock.On("StoreInterconnection", mock.Anything).
			Return(nil).Once()

		cacheContextMock := new(ContextCacheMock)
		cacheContextMock.On("RetrieveContext", userID).
			Return([]cache.Context{
				{
					UserID:    userID,
					Timestamp: 1111111111,
					Text:      "text",
				},
			}).Once()
		manager := &Manager{
			SalesforceService:     salesforceMock,
			interconnectionsCache: interconnectionMock,
			contextcache:          cacheContextMock,
			interconnectionMap: interconnectionCache{
				interconnections: interconnectionMap{
					"55555555555": &Interconnection{
						Status:        Active,
						AffinityToken: affinityToken,
						SessionKey:    sessionKey,
						SessionID:     sessionID,
						UserID:        userID,
					},
				},
			},
		}

		err := manager.CreateChat(interconnection)
		assert.NoError(t, err)
		manager.EndChat(interconnection)
	})

	t.Run("Change to from-sf-blocked state succesfull", func(t *testing.T) {
		expectedLog := "could not create chat in salesforce: this contact is blocked"
		interconnection := &Interconnection{
			UserID:      userID,
			BotSlug:     botSlug,
			BotID:       botID,
			Name:        name,
			Provider:    provider,
			Email:       email,
			PhoneNumber: phoneNumber,
		}

		SfcOrganizationID = organizationID
		SfcDeploymentID = deploymentID
		SfcAccountRecordTypeID = ""
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
			interconnection.PhoneNumber,
			"").
			Return(contact, nil).Once()
		botRunnerMock.On("SendTo", map[string]interface{}{"botSlug": botSlug, "message": "", "state": BlockedUserState, "userId": userID}).
			Return(true, nil).Once()

		interconnectionMock := new(InterconnectionCache)
		interconnectionMock.On("RetrieveInterconnectionActiveByUserId", userID).
			Return(nil, nil).Once()
		manager := &Manager{
			SalesforceService:     salesforceServiceMock,
			BotrunnnerClient:      botRunnerMock,
			interconnectionsCache: interconnectionMock,
		}
		err := manager.CreateChat(interconnection)
		assert.Error(t, err)

		if !strings.Contains(err.Error(), expectedLog) {
			t.Fatalf("Error message should contain %s, but this was found <%s>", expectedLog, err.Error())
		}
	})

}

func TestSalesforceService_FinishChat(t *testing.T) {
	m, s := cache.CreateRedisServer()
	defer m.Close()
	defer s.Close()
	t.Run("Finish Chat Succesfull", func(t *testing.T) {

		salesforceMock := new(SalesforceServiceInterface)
        manager := &Manager{
            interconnectionMap: interconnectionCache{
                interconnections: interconnectionMap{
                    "55125421545": &Interconnection{
                        Status:        Active,
                        AffinityToken: affinityToken,
                        SessionKey:    sessionKey,
                    },
                },
            },
            SalesforceService: salesforceMock,
        }

		salesforceMock.On("EndChat",
			affinityToken, sessionKey).
			Return(nil).Once()

		manager.SalesforceService = salesforceMock

		err := manager.FinishChat(userID)
		assert.NoError(t, err)
	})

}

func TestManager_SaveContext(t *testing.T) {
	t.Run("Should save context voice", func(t *testing.T) {
		contextCache := new(ContextCacheMock)
		ctx := cache.Context{
			UserID:    "55555555555",
			Timestamp: 1631202334957,
			URL:       "uri",
			MIMEType:  "voice",
			Caption:   "caption",
			From:      fromUser,
		}
		contextCache.On("StoreContext", ctx).Return(nil).Once()

		cacheMessage := new(IMessageCache)
		cacheMessage.On("IsRepeatedMessage", messageID).Return(false).Once()

		manager := &Manager{
			contextcache: contextCache,
			cacheMessage: cacheMessage,
		}

		integrations := &models.IntegrationsRequest{
			ID:        messageID,
			Timestamp: "1631202334957",
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

		cacheMessage := new(IMessageCache)
		cacheMessage.On("IsRepeatedMessage", messageID).Return(false).Once()

		manager := &Manager{
			contextcache: contextCache,
			cacheMessage: cacheMessage,
		}

		integrations := &models.IntegrationsRequest{
			ID:        messageID,
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

		cacheMessage := new(IMessageCache)
		cacheMessage.On("IsRepeatedMessage", messageID).Return(false).Once()

		manager := &Manager{
			contextcache: contextCache,
			cacheMessage: cacheMessage,
		}

		integrations := &models.IntegrationsRequest{
			ID:        messageID,
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

		cacheMessage := new(IMessageCache)
		cacheMessage.On("IsRepeatedMessage", messageID).Return(false).Once()

		manager := &Manager{
			contextcache: contextCache,
			cacheMessage: cacheMessage,
		}

		integrations := &models.IntegrationsRequest{
			ID:        messageID,
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

		cacheMessage := new(IMessageCache)
		cacheMessage.On("IsRepeatedMessage", messageID).Return(false).Once()

		manager := &Manager{
			contextcache: contextCache,
			cacheMessage: cacheMessage,
		}

		integrations := &models.IntegrationsRequest{
			ID:        messageID,
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

		cacheMessage := new(IMessageCache)
		cacheMessage.On("IsRepeatedMessage", messageID).Return(false).Once()

		manager := &Manager{
			contextcache: contextCache,
			cacheMessage: cacheMessage,
		}

		integrations := &models.IntegrationsRequest{
			ID:        messageID,
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

		cacheMessage := new(IMessageCache)
		cacheMessage.On("IsRepeatedMessage", messageID).Return(false).Once()

		manager := &Manager{
			contextcache: contextCache,
			cacheMessage: cacheMessage,
		}

		integrations := &models.IntegrationsRequest{
			ID:        messageID,
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

	t.Run("Should save context repeated message", func(t *testing.T) {
		contextCache := new(ContextCacheMock)
		ctx := cache.Context{
			UserID:    "55555555555",
			Timestamp: 123456789,
			Text:      "text",
			From:      fromUser,
		}
		contextCache.On("StoreContext", ctx).Return(assert.AnError)

		cacheMessage := new(IMessageCache)
		cacheMessage.On("IsRepeatedMessage", messageID).Return(true).Once()

		manager := &Manager{
			contextcache: contextCache,
			cacheMessage: cacheMessage,
		}

		integrations := &models.IntegrationsRequest{
			ID:        messageID,
			Timestamp: "error",
			Type:      "error",
			From:      "55555555555",
			Text: models.Text{
				Body: "text",
			},
		}
		err := manager.SaveContext(integrations)

		assert.NoError(t, err)
	})

	t.Run("Should send message to salesforce", func(t *testing.T) {
		contextCache := new(ContextCacheMock)
		salesforceMock := new(SalesforceServiceInterface)
		salesforceMock.On("SendMessage",
			affinityToken, sessionKey, chat.MessagePayload{Text: "message"}).
			Return(false, nil).Once()

		channelSaleforce := make(chan *Message)
		channelIntegrations := make(chan *Message)
		channelFinish := make(chan *Interconnection)

		cacheMessage := new(IMessageCache)
		cacheMessage.On("IsRepeatedMessage", messageID).Return(false).Once()

		manager := &Manager{
			contextcache:          contextCache,
			salesforceChannel:     channelSaleforce,
			integrationsChannel:   channelIntegrations,
			finishInterconnection: channelFinish,
			SalesforceService:     salesforceMock,
			cacheMessage:          cacheMessage,
		}
		go manager.handleInterconnection()

		manager.interconnectionMap = interconnectionCache{
			interconnections: interconnectionMap{
				"55555555555": &Interconnection{
					Status:              Active,
					AffinityToken:       affinityToken,
					SessionKey:          sessionKey,
					SessionID:           sessionID,
					UserID:              userID,
					salesforceChannel:   manager.salesforceChannel,
					integrationsChannel: manager.integrationsChannel,
					finishChannel:       manager.finishInterconnection,
				},
			},
		}

		integrations := &models.IntegrationsRequest{
			ID:        messageID,
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

	t.Run("Should send message end chat", func(t *testing.T) {
		contextCache := new(ContextCacheMock)
		interconnectionCacheMock := new(InterconnectionCache)
		salesforceMock := new(SalesforceServiceInterface)
		salesforceMock.On("EndChat",
			affinityToken, sessionKey).
			Return(nil).Once()

		cacheMock := &cache.Interconnection{
			UserID:     userID,
			SessionID:  sessionID,
			SessionKey: sessionID,
			Status:     string(Active),
		}
		interconnectionCacheMock.On("RetrieveInterconnection",
			cache.Interconnection{
				UserID:    userID,
				SessionID: sessionID,
			}).
			Return(cacheMock, nil).Once()
		cacheMock.Status = string(Closed)
		interconnectionCacheMock.On("StoreInterconnection", *cacheMock).
			Return(nil).Once()

		cacheMessage := new(IMessageCache)
		cacheMessage.On("IsRepeatedMessage", messageID).Return(false).Once()

		manager := &Manager{
			contextcache:          contextCache,
			salesforceChannel:     make(chan *Message),
			integrationsChannel:   make(chan *Message),
			finishInterconnection: make(chan *Interconnection),
			SalesforceService:     salesforceMock,
			keywordsRestart:       []string{"restart", "test"},
			interconnectionsCache: interconnectionCacheMock,
			cacheMessage:          cacheMessage,
		}
		go manager.handleInterconnection()

		manager.interconnectionMap = interconnectionCache{
			interconnections: interconnectionMap{
				"55555555555": &Interconnection{
					Status:               Active,
					AffinityToken:        affinityToken,
					SessionKey:           sessionKey,
					SessionID:            sessionID,
					UserID:               userID,
					salesforceChannel:    manager.salesforceChannel,
					integrationsChannel:  manager.integrationsChannel,
					finishChannel:        manager.finishInterconnection,
					interconnectionCache: interconnectionCacheMock,
				},
			},
		}

		integrations := &models.IntegrationsRequest{
			ID:        messageID,
			Timestamp: "123456789",
			Type:      textType,
			From:      "55555555555",
			Text: models.Text{
				Body: "ReStArt",
			},
		}
		err := manager.SaveContext(integrations)

		assert.NoError(t, err)
	})

	t.Run("Should send image to salesforce", func(t *testing.T) {
		contextCache := new(ContextCacheMock)
		salesforceMock := new(SalesforceServiceInterface)
		salesforceMock.On("InsertImageInCase",
			"http://test.com", sessionID, "image/png", "caseID").
			Return(nil).Once()

		salesforceMock.On("SendMessage",
			affinityToken, sessionKey, chat.MessagePayload{Text: messageImageSuccess}).
			Return(false, nil).Once()

		cacheMessage := new(IMessageCache)
		cacheMessage.On("IsRepeatedMessage", messageID).Return(false).Once()

		manager := &Manager{
			contextcache:          contextCache,
			salesforceChannel:     make(chan *Message),
			integrationsChannel:   make(chan *Message),
			finishInterconnection: make(chan *Interconnection),
			SalesforceService:     salesforceMock,
			keywordsRestart:       []string{"restart", "test"},
			cacheMessage:          cacheMessage,
		}
		go manager.handleInterconnection()

		manager.interconnectionMap = interconnectionCache{
			interconnections: interconnectionMap{
				"55555555555": &Interconnection{
					Status:              Active,
					AffinityToken:       affinityToken,
					SessionKey:          sessionKey,
					SessionID:           sessionID,
					UserID:              userID,
					CaseID:              caseID,
					salesforceChannel:   manager.salesforceChannel,
					integrationsChannel: manager.integrationsChannel,
					finishChannel:       manager.finishInterconnection,
				},
			},
		}

		integrations := &models.IntegrationsRequest{
			ID:        messageID,
			Timestamp: "123456789",
			Type:      imageType,
			From:      "55555555555",
			Image: models.Media{
				URL:      "http://test.com",
				MIMEType: "image/png",
				Caption:  "caption",
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
				Timestamp: 1631202337350,
				Text: `this a test
second line
`,
				From: fromUser,
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
			contextcache: contextCache,
		}

		ctxStr := manager.GetContextByUserID(userID)
		expected := `Cliente [31-08-2021 05:00:00]:Hello

Bot [31-08-2021 05:01:00]:Hello I'm a bot

Cliente [31-08-2021 05:02:00]:I need help

Bot [31-08-2021 05:04:00]:ok.

Cliente [09-09-2021 10:45:37]:this a test
second line

`
		assert.Equal(t, expected, ctxStr)
	})
}

func TestManager_SaveContextFB(t *testing.T) {
	t.Run("Should save context text from user", func(t *testing.T) {
		contextCache := new(ContextCacheMock)
		ctx := cache.Context{
			UserID:    userID,
			Timestamp: 1631202334957,
			Text:      "text",
			From:      fromUser,
		}
		contextCache.On("StoreContext", ctx).Return(nil).Once()

		cacheMessage := new(IMessageCache)
		cacheMessage.On("IsRepeatedMessage", messageID).Return(false).Once()

		manager := &Manager{
			contextcache: contextCache,
			cacheMessage: cacheMessage,
		}

		integrations := &models.IntegrationsFacebook{
			AuthorRole: fromUser,
			BotID:      "botID",
			Timestamp:  1631202334957,
			Message: models.Message{
				Entry: []models.Entry{
					{
						ID: "id",
						Messaging: []models.Messaging{
							{
								Message: models.MessagingMessage{
									Mid:  messageID,
									Text: "text",
								},
								Recipient: models.Recipient{
									ID: botID,
								},
								Sender: models.Recipient{
									ID: userID,
								},
								Timestamp: 1631202334957,
							},
						},
						Time: 12345,
					},
				},
				Object: "object",
			},
			Provider:    "facebook",
			MsgTracking: models.MsgTracking{},
		}
		err := manager.SaveContextFB(integrations)

		assert.NoError(t, err)
	})

	t.Run("Should save context repited message", func(t *testing.T) {
		cacheMessage := new(IMessageCache)
		cacheMessage.On("IsRepeatedMessage", messageID).Return(true).Once()

		manager := &Manager{
			cacheMessage: cacheMessage,
		}

		integrations := &models.IntegrationsFacebook{
			AuthorRole: fromUser,
			BotID:      "botID",
			Timestamp:  1631202334957,
			Message: models.Message{
				Entry: []models.Entry{
					{
						ID: "id",
						Messaging: []models.Messaging{
							{
								Message: models.MessagingMessage{
									Mid:  messageID,
									Text: "text",
								},
								Recipient: models.Recipient{
									ID: botID,
								},
								Sender: models.Recipient{
									ID: userID,
								},
								Timestamp: 1631202334957,
							},
						},
						Time: 12345,
					},
				},
				Object: "object",
			},
			Provider:    "facebook",
			MsgTracking: models.MsgTracking{},
		}
		err := manager.SaveContextFB(integrations)

		assert.NoError(t, err)
	})

	t.Run("Should save context text from bot", func(t *testing.T) {
		contextCache := new(ContextCacheMock)
		ctx := cache.Context{
			UserID:    userID,
			Timestamp: 1631202334957,
			Text:      "text",
			From:      fromBot,
		}
		contextCache.On("StoreContext", ctx).Return(nil).Once()

		cacheMessage := new(IMessageCache)
		cacheMessage.On("IsRepeatedMessage", messageID).Return(false).Once()

		manager := &Manager{
			contextcache: contextCache,
			cacheMessage: cacheMessage,
		}

		integrations := &models.IntegrationsFacebook{
			AuthorRole: fromBot,
			BotID:      "botID",
			Timestamp:  163120233495,
			Message: models.Message{
				Entry: []models.Entry{
					{
						ID: "id",
						Messaging: []models.Messaging{
							{
								Message: models.MessagingMessage{
									Mid:  messageID,
									Text: "text",
								},
								Recipient: models.Recipient{
									ID: userID,
								},
								Sender: models.Recipient{
									ID: botID,
								},
								Timestamp: 1631202334957,
							},
						},
						Time: 12345,
					},
				},
				Object: "object",
			},
			Provider:    "facebook",
			MsgTracking: models.MsgTracking{},
		}
		err := manager.SaveContextFB(integrations)

		assert.NoError(t, err)
	})

	t.Run("Should interaction text from user", func(t *testing.T) {
		contextCache := new(ContextCacheMock)
		salesforceMock := new(SalesforceServiceInterface)
		ctx := cache.Context{
			UserID:    userID,
			Timestamp: 1631202334957,
			Text:      "text",
			From:      fromUser,
		}
		contextCache.On("StoreContext", ctx).Return(nil).Once()

		salesforceMock.On("SendMessage",
			affinityToken, sessionKey, chat.MessagePayload{Text: "text"}).
			Return(false, nil).Once()
		cacheMessage := new(IMessageCache)
		cacheMessage.On("IsRepeatedMessage", messageID).Return(false).Once()

		manager := &Manager{
			contextcache:          contextCache,
			SalesforceService:     salesforceMock,
			keywordsRestart:       []string{"restart", "test"},
			salesforceChannel:     make(chan *Message),
			finishInterconnection: make(chan *Interconnection),
			integrationsChannel:   make(chan *Message),
			cacheMessage:          cacheMessage,
		}

		go manager.handleInterconnection()

		manager.interconnectionMap = interconnectionCache{
			interconnections: interconnectionMap{
				userID: &Interconnection{
					Status:              Active,
					AffinityToken:       affinityToken,
					SessionKey:          sessionKey,
					salesforceChannel:   manager.salesforceChannel,
					integrationsChannel: manager.integrationsChannel,
					finishChannel:       manager.finishInterconnection,
				},
			},
		}

		integrations := &models.IntegrationsFacebook{
			AuthorRole: fromUser,
			BotID:      "botID",
			Timestamp:  1631202334957,
			Message: models.Message{
				Entry: []models.Entry{
					{
						ID: "id",
						Messaging: []models.Messaging{
							{
								Message: models.MessagingMessage{
									Mid:  messageID,
									Text: "text",
								},
								Recipient: models.Recipient{
									ID: botID,
								},
								Sender: models.Recipient{
									ID: userID,
								},
								Timestamp: 1631202334957,
							},
						},
						Time: 12345,
					},
				},
				Object: "object",
			},
			Provider:    "facebook",
			MsgTracking: models.MsgTracking{},
		}
		err := manager.SaveContextFB(integrations)

		assert.NoError(t, err)
	})

	t.Run("Should interaction  image from user", func(t *testing.T) {
		contextCache := new(ContextCacheMock)
		salesforceMock := new(SalesforceServiceInterface)
		salesforceMock.On("InsertImageInCase",
			"http://test.com", sessionID, "", "caseID").
			Return(nil).Once()

		salesforceMock.On("SendMessage",
			affinityToken, sessionKey, chat.MessagePayload{Text: messageImageSuccess}).
			Return(false, nil).Once()

		cacheMessage := new(IMessageCache)
		cacheMessage.On("IsRepeatedMessage", messageID).Return(false).Once()

		manager := &Manager{
			contextcache:          contextCache,
			SalesforceService:     salesforceMock,
			keywordsRestart:       []string{"restart", "test"},
			salesforceChannel:     make(chan *Message),
			finishInterconnection: make(chan *Interconnection),
			integrationsChannel:   make(chan *Message),
			cacheMessage:          cacheMessage,
		}

		go manager.handleInterconnection()

		manager.interconnectionMap = interconnectionCache{
			interconnections: interconnectionMap{
				userID: &Interconnection{
					Status:              Active,
					AffinityToken:       affinityToken,
					SessionKey:          sessionKey,
					CaseID:              caseID,
					SessionID:           sessionID,
					salesforceChannel:   manager.salesforceChannel,
					integrationsChannel: manager.integrationsChannel,
					finishChannel:       manager.finishInterconnection,
					SalesforceService:   salesforceMock,
				},
			},
		}

		integrations := &models.IntegrationsFacebook{
			AuthorRole: fromUser,
			BotID:      "botID",
			Timestamp:  1631202334957,
			Message: models.Message{
				Entry: []models.Entry{
					{
						ID: "id",
						Messaging: []models.Messaging{
							{
								Recipient: models.Recipient{
									ID: botID,
								},
								Sender: models.Recipient{
									ID: userID,
								},
								Message: models.MessagingMessage{
									Mid: messageID,
									Attachments: []models.Attachment{
										{
											Payload: models.Payload{
												URL: "http://test.com",
											},
											Type: imageType,
										},
									},
								},
								Timestamp: 1631202334957,
							},
						},
						Time: 12345,
					},
				},
				Object: "object",
			},
			Provider:    "facebook",
			MsgTracking: models.MsgTracking{},
		}
		err := manager.SaveContextFB(integrations)

		assert.NoError(t, err)
	})

	t.Run("Should save context endchat", func(t *testing.T) {
		contextCache := new(ContextCacheMock)
		salesforceMock := new(SalesforceServiceInterface)
		salesforceMock.On("EndChat",
			affinityToken, sessionKey).
			Return(nil).Once()

		cacheMock := &cache.Interconnection{
			UserID:     userID,
			SessionID:  sessionID,
			SessionKey: sessionID,
			Status:     string(Active),
		}
		interconnectionCacheMock := new(InterconnectionCache)
		interconnectionCacheMock.On("RetrieveInterconnection",
			cache.Interconnection{
				UserID:    userID,
				SessionID: sessionID,
			}).
			Return(cacheMock, nil).Once()
		cacheMock.Status = string(Closed)
		interconnectionCacheMock.On("StoreInterconnection", *cacheMock).
			Return(nil).Once()

		cacheMessage := new(IMessageCache)
		cacheMessage.On("IsRepeatedMessage", messageID).Return(false).Once()

		manager := &Manager{
			contextcache:          contextCache,
			SalesforceService:     salesforceMock,
			keywordsRestart:       []string{"restart", "test"},
			salesforceChannel:     make(chan *Message),
			finishInterconnection: make(chan *Interconnection),
			integrationsChannel:   make(chan *Message),
			cacheMessage:          cacheMessage,
		}

		go manager.handleInterconnection()

		manager.interconnectionMap = interconnectionCache{
			interconnections: interconnectionMap{
				userID: &Interconnection{
					Status:               Active,
					AffinityToken:        affinityToken,
					SessionKey:           sessionKey,
					CaseID:               caseID,
					UserID:               userID,
					SessionID:            sessionID,
					salesforceChannel:    manager.salesforceChannel,
					integrationsChannel:  manager.integrationsChannel,
					finishChannel:        manager.finishInterconnection,
					SalesforceService:    salesforceMock,
					interconnectionCache: interconnectionCacheMock,
				},
			},
		}

		integrations := &models.IntegrationsFacebook{
			AuthorRole: fromUser,
			BotID:      "botID",
			Timestamp:  1631202334957,
			Message: models.Message{
				Entry: []models.Entry{
					{
						ID: "id",
						Messaging: []models.Messaging{
							{
								Message: models.MessagingMessage{
									Mid:  messageID,
									Text: "ResTart",
								},
								Recipient: models.Recipient{
									ID: botID,
								},
								Sender: models.Recipient{
									ID: userID,
								},
								Timestamp: 1631202334957,
							},
						},
						Time: 12345,
					},
				},
				Object: "object",
			},
			Provider:    "facebook",
			MsgTracking: models.MsgTracking{},
		}
		err := manager.SaveContextFB(integrations)

		assert.NoError(t, err)
	})

	t.Run("Should interaction  image from user error", func(t *testing.T) {
		contextCache := new(ContextCacheMock)
		salesforceMock := new(SalesforceServiceInterface)
		salesforceMock.On("InsertImageInCase",
			"http://test.com", sessionID, "", "caseID").
			Return(assert.AnError).Once()

		salesforceMock.On("SendMessage",
			affinityToken, sessionKey, chat.MessagePayload{Text: messageImageSuccess}).
			Return(false, nil).Once()

		integrationsIMock := new(IntegrationInterface)
		integrationsIMock.On("SendMessage", integrations.SendTextPayloadFB{
			MessagingType: "RESPONSE",
			Recipient: integrations.Recipient{
				ID: userID,
			},
			Message: integrations.Message{
				Text: messageError,
			},
			Metadata: "YALOSOURCE:FIREHOSE",
		}, string(FacebookProvider)).Return(&integrations.SendMessageResponse{}, nil).Once()

		cacheMessage := new(IMessageCache)
		cacheMessage.On("IsRepeatedMessage", messageID).Return(false).Once()

		manager := &Manager{
			contextcache:          contextCache,
			SalesforceService:     salesforceMock,
			keywordsRestart:       []string{"restart", "test"},
			salesforceChannel:     make(chan *Message),
			finishInterconnection: make(chan *Interconnection),
			integrationsChannel:   make(chan *Message),
			IntegrationsClient:    integrationsIMock,
			cacheMessage:          cacheMessage,
		}
		go manager.handleInterconnection()

		time.Sleep(time.Second)
		manager.interconnectionMap = interconnectionCache{
			interconnections: interconnectionMap{
				userID: &Interconnection{
					Status:              Active,
					AffinityToken:       affinityToken,
					SessionKey:          sessionKey,
					CaseID:              caseID,
					SessionID:           sessionID,
					UserID:              userID,
					salesforceChannel:   manager.salesforceChannel,
					integrationsChannel: manager.integrationsChannel,
					finishChannel:       manager.finishInterconnection,
					SalesforceService:   salesforceMock,
					IntegrationsClient:  integrationsIMock,
				},
			},
		}

		integrations := &models.IntegrationsFacebook{
			AuthorRole: fromUser,
			BotID:      "botID",
			Timestamp:  1631202334957,
			Message: models.Message{
				Entry: []models.Entry{
					{
						ID: "id",
						Messaging: []models.Messaging{
							{
								Recipient: models.Recipient{
									ID: botID,
								},
								Sender: models.Recipient{
									ID: userID,
								},
								Message: models.MessagingMessage{
									Mid: messageID,
									Attachments: []models.Attachment{
										{
											Payload: models.Payload{
												URL: "http://test.com",
											},
											Type: imageType,
										},
									},
								},
								Timestamp: 1631202334957,
							},
						},
						Time: 12345,
					},
				},
				Object: "object",
			},
			Provider:    "facebook",
			MsgTracking: models.MsgTracking{},
		}

		err := manager.SaveContextFB(integrations)

		assert.Error(t, err)
	})
}
