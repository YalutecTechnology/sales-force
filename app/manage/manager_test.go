package manage

import (
	"bytes"
	"fmt"
	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"strings"
	"testing"
	"time"

	"yalochat.com/salesforce-integration/app/config/envs"
	"yalochat.com/salesforce-integration/base/cache"
	"yalochat.com/salesforce-integration/base/clients/chat"
	"yalochat.com/salesforce-integration/base/clients/integrations"
	"yalochat.com/salesforce-integration/base/constants"
	"yalochat.com/salesforce-integration/base/models"
)

const (
	name                = "username"
	client              = "client"
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

	interconnection := &Interconnection{
		Client:        client,
		BotID:         "botID",
		BotSlug:       "coppel-bot",
		UserID:        "userID",
		Status:        Active,
		SessionID:     "session",
		SessionKey:    "sessionID",
		AffinityToken: "affinityToken",
		Timestamp:     time.Time{},
		Provider:      "provider",
		Name:          "name",
		Email:         "email",
		PhoneNumber:   "55555555555",
		CaseID:        "caseID",
		ExtraData: map[string]interface{}{
			"data": "data",
		},
	}

	t.Run("Should retrieve a manager instance", func(t *testing.T) {
		configRedis := &cache.RedisOptions{
			FailOverOptions: &redis.FailoverOptions{
				MasterName:    s.MasterInfo().Name,
				SentinelAddrs: []string{s.Addr()},
			},
			SessionsTTL: time.Second,
		}
		rcs, _ := cache.NewRedisCache(configRedis)
		rcs.StoreInterconnection(NewInterconectionCache(interconnection))

		salesforceMock := new(SalesforceServiceInterface)
		salesforceMock.On("GetMessages",
			affinityToken, sessionKey).
			Return(&chat.MessagesResponse{}, nil).Once()

		expected := &Manager{
			client:             client,
			clientName:         "salesforce-integration",
			SalesforceService:  nil,
			IntegrationsClient: nil,
			BotrunnnerClient:   nil,
			cacheMessage:       nil,
			interconnectionMap: nil,
			StudioNG:           nil,
		}
		config := &ManagerOptions{
			AppName:                    "salesforce-integration",
			RedisOptions:               *configRedis,
			Client:                     client,
			BotrunnerUrl:               "uri",
			StudioNGUrl:                "uriStudio",
			SfcDefaultBirthDateAccount: "1999-01-01T00:00:00",
			SpecSchedule:               "@every 1h30m",
			CleanContextSchedule:       "0 9 * * *",
		}
		actual := CreateManager(config)
		actual.SalesforceService = salesforceMock
		actual.IntegrationsClient = nil
		actual.BotrunnnerClient = nil
		actual.StudioNG = nil
		actual.cacheMessage = nil
		expected.SalesforceService = salesforceMock
		expected.integrationsChannel = actual.integrationsChannel
		expected.salesforceChannel = actual.salesforceChannel
		expected.finishInterconnection = actual.finishInterconnection
		expected.contextcache = actual.contextcache
		expected.interconnectionsCache = actual.interconnectionsCache
		expected.isStudioNGFlow = true
		expected.interconnectionMap = actual.interconnectionMap
		actual.EndChat(interconnection)
		assert.Equal(t, expected, actual)
		time.Sleep(1 * time.Second)
	})
}

func TestManager_CreateChat(t *testing.T) {
	interconectionLocal := cache.New()
	t.Run("Create Chat Succesfull", func(t *testing.T) {
		defer interconectionLocal.Clear()
		interconnection := &Interconnection{
			UserID:      userID,
			Client:      client,
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
			interconnection.PhoneNumber).
			Return(contact, nil).Once()

		salesforceMock.On("CreatCase",
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

		interconnectionMock.On("RetrieveInterconnection", cache.Interconnection{UserID: userID, Client: client}).
			Return(nil, nil).Once()
		interconnectionMock.On("StoreInterconnection", mock.Anything).
			Return(nil).Once()

		cacheContextMock := new(ContextCacheMock)
		cacheContextMock.On("RetrieveContextFromSet", client, userID).
			Return([]cache.Context{
				{
					Client:    client,
					UserID:    userID,
					Timestamp: 1111111111,
					Text:      "text",
				},
			}).Once()

		manager := &Manager{
			client:                client,
			SalesforceService:     salesforceMock,
			interconnectionsCache: interconnectionMock,
			contextcache:          cacheContextMock,
			interconnectionMap:    interconectionLocal,
			SfcSourceFlowField:    "data",
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
		defer interconectionLocal.Clear()
		interconnection := &Interconnection{
			UserID:      userID,
			Client:      client,
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
			interconnection.PhoneNumber).
			Return(contact, nil).Once()

		salesforceMock.On("CreatCase",
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
		interconnectionMock.On("RetrieveInterconnection", cache.Interconnection{
			UserID: userID, Client: client,
		}).
			Return(nil, nil).Once()
		interconnectionMock.On("StoreInterconnection", mock.Anything).
			Return(nil).Once()

		contextMock := new(ContextCacheMock)
		contextMock.On("RetrieveContextFromSet",
			client, userID).
			Return([]cache.Context{
				{
					Client:    client,
					UserID:    userID,
					Timestamp: 1111111111,
					Text:      "text",
				},
			}).Once()

		manager := &Manager{
			client:                client,
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
			contextcache:       contextMock,
			interconnectionMap: interconectionLocal,
		}

		err := manager.CreateChat(interconnection)
		assert.NoError(t, err)
		manager.EndChat(interconnection)
	})

	t.Run("Create Chat Succesfull with an account", func(t *testing.T) {
		defer interconectionLocal.Clear()
		interconnection := &Interconnection{
			UserID:      userID,
			Client:      client,
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
			interconnection.PhoneNumber).
			Return(contact, nil).Once()

		salesforceMock.On("CreatCase",
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

		interconnectionMock.On("RetrieveInterconnection", cache.Interconnection{
			UserID: userID, Client: client,
		}).
			Return(nil, nil).Once()
		interconnectionMock.On("StoreInterconnection", mock.Anything).
			Return(nil).Once()

		cacheContextMock := new(ContextCacheMock)
		cacheContextMock.On("RetrieveContextFromSet", client, userID).
			Return([]cache.Context{
				{
					Client:    client,
					UserID:    userID,
					Timestamp: 1111111111,
					Text:      "text",
				},
			}).Once()
		manager := &Manager{
			client:                client,
			SalesforceService:     salesforceMock,
			interconnectionsCache: interconnectionMock,
			contextcache:          cacheContextMock,
			interconnectionMap:    interconectionLocal,
		}

		err := manager.CreateChat(interconnection)
		assert.NoError(t, err)
		manager.EndChat(interconnection)
	})

	t.Run("Change to from-sf-blocked state succesfull", func(t *testing.T) {
		defer interconectionLocal.Clear()
		expectedLog := "could not create chat in salesforce: this contact is blocked"
		interconnection := &Interconnection{
			UserID:      userID,
			Client:      client,
			BotSlug:     botSlug,
			BotID:       botID,
			Name:        name,
			Provider:    provider,
			Email:       email,
			PhoneNumber: phoneNumber,
		}

		SfcOrganizationID = organizationID
		SfcDeploymentID = deploymentID
		BlockedUserState = map[string]string{
			provider:                 "from-sf-blocked",
			string(FacebookProvider): "from-sf-blocked",
		}

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
		botRunnerMock.On("SendTo", map[string]interface{}{"botSlug": botSlug, "message": "", "state": blockedUserState, "userId": userID}).
			Return(true, nil).Once()

		interconnectionMock := new(InterconnectionCache)
		interconnectionMock.On("RetrieveInterconnection", cache.Interconnection{UserID: userID, Client: client}).
			Return(nil, nil).Once()
		manager := &Manager{
			client:                client,
			SalesforceService:     salesforceServiceMock,
			BotrunnnerClient:      botRunnerMock,
			interconnectionsCache: interconnectionMock,
			interconnectionMap:    interconectionLocal,
		}
		err := manager.CreateChat(interconnection)
		assert.Error(t, err)

		if !strings.Contains(err.Error(), expectedLog) {
			t.Fatalf("Error message should contain %s, but this was found <%s>", expectedLog, err.Error())
		}
	})

	t.Run("Create Chat Failed for invalid user", func(t *testing.T) {
		defer interconectionLocal.Clear()
		interconnection := &Interconnection{
			UserID:      userID,
			Client:      client,
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

		interconnectionMock.On("RetrieveInterconnection", cache.Interconnection{UserID: userID, Client: client}).
			Return(&cache.Interconnection{UserID: userID, Status: string(Active)}, nil).Once()

		cacheContextMock := new(ContextCacheMock)

		manager := &Manager{
			client:                client,
			SalesforceService:     salesforceMock,
			interconnectionsCache: interconnectionMock,
			contextcache:          cacheContextMock,
			interconnectionMap:    interconectionLocal,
			SfcSourceFlowField:    "data",
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
		assert.Error(t, err)
		manager.EndChat(interconnection)
	})

}

func TestManager_FinishChat(t *testing.T) {
	interconectionLocal := cache.New()
	t.Run("Finish Chat Succesfull", func(t *testing.T) {
		interconnectionCacheMock := new(InterconnectionCache)
		salesforceMock := new(SalesforceServiceInterface)
		interconnection := &Interconnection{
			Client:               client,
			Status:               Active,
			AffinityToken:        affinityToken,
			SessionKey:           sessionKey,
			interconnectionCache: interconnectionCacheMock}
		interconectionLocal.Set(fmt.Sprintf(constants.UserKey, userID), interconnection, ttlMessage)
		interconectionLocal.Wait()
		interconnectionCache := &cache.Interconnection{
			UserID:     userID,
			SessionID:  sessionID,
			SessionKey: sessionID,
			Status:     string(Active),
		}

		manager := &Manager{
			interconnectionMap:    interconectionLocal,
			SalesforceService:     salesforceMock,
			interconnectionsCache: interconnectionCacheMock,
		}

		interconnectionCacheMock.On("RetrieveInterconnection",
			cache.Interconnection{UserID: interconnection.UserID, Client: client}).
			Return(interconnectionCache, nil).Once()
		interconnectionCache.Status = string(Closed)
		interconnectionCacheMock.On("StoreInterconnection", *interconnectionCache).
			Return(nil).Once()

		salesforceMock.On("EndChat",
			affinityToken, sessionKey).
			Return(nil).Once()

		manager.SalesforceService = salesforceMock

		err := manager.FinishChat(userID)
		assert.NoError(t, err)
	})

}

func TestManager_SaveContext(t *testing.T) {
	interconnectionLocal := cache.New()

	t.Run("Should save context audio", func(t *testing.T) {
		contextCache := new(ContextCacheMock)
		contextCache.On("StoreContextToSet", mock.Anything).Return(nil).Once()

		cacheMessage := new(IMessageCache)
		cacheMessage.On("IsRepeatedMessage", messageID).Return(false).Once()

		manager := &Manager{
			contextcache:       contextCache,
			cacheMessage:       cacheMessage,
			interconnectionMap: interconnectionLocal,
		}

		integrations := &models.IntegrationsRequest{
			ID:        messageID,
			Timestamp: "1631202334956",
			Type:      audioType,
			From:      userID,
			Audio: models.Media{
				URL:      "uri",
				MIMEType: "audio",
			},
		}
		err := manager.SaveContext(integrations)

		assert.NoError(t, err)
	})

	t.Run("Should save context voice", func(t *testing.T) {
		contextCache := new(ContextCacheMock)
		contextCache.On("StoreContextToSet", mock.Anything).Return(nil).Once()

		cacheMessage := new(IMessageCache)
		cacheMessage.On("IsRepeatedMessage", messageID).Return(false).Once()

		manager := &Manager{
			contextcache:       contextCache,
			cacheMessage:       cacheMessage,
			interconnectionMap: interconnectionLocal,
		}

		integrations := &models.IntegrationsRequest{
			ID:        messageID,
			Timestamp: "1631202334957",
			Type:      voiceType,
			From:      userID,
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
		contextCache.On("StoreContextToSet", mock.Anything).Return(nil).Once()

		cacheMessage := new(IMessageCache)
		cacheMessage.On("IsRepeatedMessage", messageID).Return(false).Once()

		manager := &Manager{
			contextcache:       contextCache,
			cacheMessage:       cacheMessage,
			interconnectionMap: interconnectionLocal,
		}

		integrations := &models.IntegrationsRequest{
			ID:        messageID,
			Timestamp: "123456789",
			Type:      documentType,
			To:        userID,
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
		contextCache.On("StoreContextToSet", mock.Anything).Return(nil).Once()

		cacheMessage := new(IMessageCache)
		cacheMessage.On("IsRepeatedMessage", messageID).Return(false).Once()

		manager := &Manager{
			contextcache:       contextCache,
			cacheMessage:       cacheMessage,
			interconnectionMap: interconnectionLocal,
		}

		integrations := &models.IntegrationsRequest{
			ID:        messageID,
			Timestamp: "123456789",
			Type:      imageType,
			From:      userID,
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
		contextCache.On("StoreContextToSet", mock.Anything).Return(nil).Once()

		cacheMessage := new(IMessageCache)
		cacheMessage.On("IsRepeatedMessage", messageID).Return(false).Once()

		manager := &Manager{
			contextcache:       contextCache,
			cacheMessage:       cacheMessage,
			interconnectionMap: interconnectionLocal,
		}

		integrations := &models.IntegrationsRequest{
			ID:        messageID,
			Timestamp: "123456789",
			Type:      textType,
			From:      userID,
			Text: models.Text{
				Body: "text",
			},
		}
		err := manager.SaveContext(integrations)

		assert.NoError(t, err)
	})

	t.Run("Should save context  error StoreContextToSet", func(t *testing.T) {
		contextCache := new(ContextCacheMock)
		contextCache.On("StoreContextToSet", mock.Anything).Return(assert.AnError)

		cacheMessage := new(IMessageCache)
		cacheMessage.On("IsRepeatedMessage", messageID).Return(false).Once()

		manager := &Manager{
			client:             client,
			contextcache:       contextCache,
			cacheMessage:       cacheMessage,
			interconnectionMap: interconnectionLocal,
		}

		integrations := &models.IntegrationsRequest{
			ID:        messageID,
			Timestamp: "123456789",
			Type:      textType,
			From:      userID,
			Text: models.Text{
				Body: "text",
			},
		}
		err := manager.SaveContext(integrations)

		assert.NoError(t, err)
	})

	t.Run("Should save context with error", func(t *testing.T) {
		contextCache := new(ContextCacheMock)
		contextCache.On("StoreContextToSet", mock.Anything).Return(assert.AnError)

		cacheMessage := new(IMessageCache)
		cacheMessage.On("IsRepeatedMessage", messageID).Return(false).Once()

		manager := &Manager{
			contextcache:       contextCache,
			cacheMessage:       cacheMessage,
			interconnectionMap: interconnectionLocal,
		}

		integrations := &models.IntegrationsRequest{
			ID:        messageID,
			Timestamp: "123456789",
			Type:      textType,
			From:      userID,
			Text: models.Text{
				Body: "text",
			},
		}
		err := manager.SaveContext(integrations)

		assert.NoError(t, err)
	})

	t.Run("Should save context default", func(t *testing.T) {
		contextCache := new(ContextCacheMock)
		contextCache.On("StoreContextToSet", mock.Anything).Return(assert.AnError)

		cacheMessage := new(IMessageCache)
		cacheMessage.On("IsRepeatedMessage", messageID).Return(false).Once()

		manager := &Manager{
			contextcache:       contextCache,
			cacheMessage:       cacheMessage,
			interconnectionMap: interconnectionLocal,
		}

		integrations := &models.IntegrationsRequest{
			ID:        messageID,
			Timestamp: "123456789",
			Type:      "error",
			From:      userID,
			Text: models.Text{
				Body: "text",
			},
		}
		err := manager.SaveContext(integrations)

		assert.NoError(t, err)
	})

	t.Run("Should save context error timestamp", func(t *testing.T) {
		contextCache := new(ContextCacheMock)
		contextCache.On("StoreContextToSet", mock.Anything).Return(assert.AnError)

		cacheMessage := new(IMessageCache)
		cacheMessage.On("IsRepeatedMessage", messageID).Return(false).Once()

		manager := &Manager{
			contextcache:       contextCache,
			cacheMessage:       cacheMessage,
			interconnectionMap: interconnectionLocal,
		}

		integrations := &models.IntegrationsRequest{
			ID:        messageID,
			Timestamp: "error",
			Type:      "error",
			From:      userID,
			Text: models.Text{
				Body: "text",
			},
		}
		err := manager.SaveContext(integrations)

		assert.Error(t, err)
	})

	t.Run("Should save context repeated message", func(t *testing.T) {
		contextCache := new(ContextCacheMock)
		contextCache.On("StoreContextToSet", mock.Anything).Return(assert.AnError)

		cacheMessage := new(IMessageCache)
		cacheMessage.On("IsRepeatedMessage", messageID).Return(true).Once()

		manager := &Manager{
			contextcache:       contextCache,
			cacheMessage:       cacheMessage,
			interconnectionMap: interconnectionLocal,
		}

		integrations := &models.IntegrationsRequest{
			ID:        messageID,
			Timestamp: "error",
			Type:      "error",
			From:      userID,
			Text: models.Text{
				Body: "text",
			},
		}
		err := manager.SaveContext(integrations)

		assert.NoError(t, err)
	})

	t.Run("Should send message to salesforce", func(t *testing.T) {
		defer interconnectionLocal.Clear()
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

		interconnectionLocal.Set(fmt.Sprintf(constants.UserKey, userID), &Interconnection{
			Status:              Active,
			AffinityToken:       affinityToken,
			SessionKey:          sessionKey,
			SessionID:           sessionID,
			UserID:              userID,
			salesforceChannel:   manager.salesforceChannel,
			integrationsChannel: manager.integrationsChannel,
			finishChannel:       manager.finishInterconnection,
		}, time.Second)
		interconnectionLocal.Wait()

		manager.interconnectionMap = interconnectionLocal

		integrations := &models.IntegrationsRequest{
			ID:        messageID,
			Timestamp: "123456789",
			Type:      textType,
			From:      userID,
			Text: models.Text{
				Body: "message",
			},
		}
		err := manager.SaveContext(integrations)
		assert.NoError(t, err)
	})

	t.Run("Should send message end chat", func(t *testing.T) {
		defer interconnectionLocal.Clear()
		contextCache := new(ContextCacheMock)
		interconnectionCacheMock := new(InterconnectionCache)
		salesforceMock := new(SalesforceServiceInterface)
		salesforceMock.On("EndChat",
			affinityToken, sessionKey).
			Return(nil).Once()

		cacheMock := &cache.Interconnection{
			UserID:     userID,
			Client:     client,
			SessionID:  sessionID,
			SessionKey: sessionID,
			Status:     string(Active),
		}
		interconnectionCacheMock.On("RetrieveInterconnection",
			cache.Interconnection{
				UserID: userID,
				Client: client,
			}).
			Return(cacheMock, nil).Once()
		cacheMock.Status = string(Closed)
		interconnectionCacheMock.On("StoreInterconnection", *cacheMock).
			Return(nil).Once()

		cacheMessage := new(IMessageCache)
		cacheMessage.On("IsRepeatedMessage", messageID).Return(false).Once()

		manager := &Manager{
			contextcache:          contextCache,
			client:                client,
			salesforceChannel:     make(chan *Message),
			integrationsChannel:   make(chan *Message),
			finishInterconnection: make(chan *Interconnection),
			SalesforceService:     salesforceMock,
			keywordsRestart:       []string{"restart", "test"},
			interconnectionsCache: interconnectionCacheMock,
			cacheMessage:          cacheMessage,
		}
		go manager.handleInterconnection()

		interconnectionLocal.Set(fmt.Sprintf(constants.UserKey, userID), &Interconnection{
			Status:               Active,
			AffinityToken:        affinityToken,
			SessionKey:           sessionKey,
			SessionID:            sessionID,
			UserID:               userID,
			Client:               client,
			salesforceChannel:    manager.salesforceChannel,
			integrationsChannel:  manager.integrationsChannel,
			finishChannel:        manager.finishInterconnection,
			interconnectionCache: interconnectionCacheMock,
		}, time.Second)
		interconnectionLocal.Wait()

		manager.interconnectionMap = interconnectionLocal

		integrations := &models.IntegrationsRequest{
			ID:        messageID,
			Timestamp: "123456789",
			Type:      textType,
			From:      userID,
			Text: models.Text{
				Body: "ReStArt",
			},
		}
		err := manager.SaveContext(integrations)

		assert.NoError(t, err)
	})

	t.Run("Should send message end chat error service", func(t *testing.T) {
		defer interconnectionLocal.Clear()
		contextCache := new(ContextCacheMock)
		interconnectionCacheMock := new(InterconnectionCache)
		salesforceMock := new(SalesforceServiceInterface)
		salesforceMock.On("EndChat",
			affinityToken, sessionKey).
			Return(assert.AnError).Once()

		cacheMock := &cache.Interconnection{
			UserID:     userID,
			Client:     client,
			SessionID:  sessionID,
			SessionKey: sessionID,
			Status:     string(Active),
		}
		interconnectionCacheMock.On("RetrieveInterconnection",
			cache.Interconnection{
				UserID: userID,
				Client: client,
			}).
			Return(cacheMock, nil).Once()
		cacheMock.Status = string(Closed)
		interconnectionCacheMock.On("StoreInterconnection", *cacheMock).
			Return(nil).Once()

		cacheMessage := new(IMessageCache)
		cacheMessage.On("IsRepeatedMessage", messageID).Return(false).Once()

		manager := &Manager{
			contextcache:          contextCache,
			client:                client,
			salesforceChannel:     make(chan *Message),
			integrationsChannel:   make(chan *Message),
			finishInterconnection: make(chan *Interconnection),
			SalesforceService:     salesforceMock,
			keywordsRestart:       []string{"restart", "test"},
			interconnectionsCache: interconnectionCacheMock,
			cacheMessage:          cacheMessage,
		}
		go manager.handleInterconnection()

		interconnectionLocal.Set(fmt.Sprintf(constants.UserKey, userID), &Interconnection{
			Status:               Active,
			AffinityToken:        affinityToken,
			SessionKey:           sessionKey,
			SessionID:            sessionID,
			UserID:               userID,
			Client:               client,
			salesforceChannel:    manager.salesforceChannel,
			integrationsChannel:  manager.integrationsChannel,
			finishChannel:        manager.finishInterconnection,
			interconnectionCache: interconnectionCacheMock,
		}, time.Second)
		interconnectionLocal.Wait()

		manager.interconnectionMap = interconnectionLocal

		integrations := &models.IntegrationsRequest{
			ID:        messageID,
			Timestamp: "123456789",
			Type:      textType,
			From:      userID,
			Text: models.Text{
				Body: "ReStArt",
			},
		}
		err := manager.SaveContext(integrations)

		assert.NoError(t, err)
	})

	t.Run("Should send image to salesforce", func(t *testing.T) {
		defer interconnectionLocal.Clear()
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

		interconnectionLocal.Set(fmt.Sprintf(constants.UserKey, userID), &Interconnection{
			Status:              Active,
			AffinityToken:       affinityToken,
			SessionKey:          sessionKey,
			SessionID:           sessionID,
			UserID:              userID,
			CaseID:              caseID,
			salesforceChannel:   manager.salesforceChannel,
			integrationsChannel: manager.integrationsChannel,
			finishChannel:       manager.finishInterconnection,
		}, time.Second)
		interconnectionLocal.Wait()

		manager.interconnectionMap = interconnectionLocal

		integrations := &models.IntegrationsRequest{
			ID:        messageID,
			Timestamp: "123456789",
			Type:      imageType,
			From:      userID,
			Image: models.Media{
				URL:      "http://test.com",
				MIMEType: "image/png",
				Caption:  "caption",
			},
		}
		err := manager.SaveContext(integrations)

		assert.NoError(t, err)
	})

	t.Run("Should send image to salesforce error service", func(t *testing.T) {
		defer interconnectionLocal.Clear()
		contextCache := new(ContextCacheMock)
		salesforceMock := new(SalesforceServiceInterface)
		salesforceMock.On("InsertImageInCase",
			"http://test.com", sessionID, "image/png", "caseID").
			Return(assert.AnError).Once()

		cacheMessage := new(IMessageCache)
		cacheMessage.On("IsRepeatedMessage", messageID).Return(false).Once()

		integrationsIMock := new(IntegrationInterface)
		integrationsIMock.On("SendMessage", mock.Anything, string(WhatsappProvider)).Return(&integrations.SendMessageResponse{}, nil).Once()

		manager := &Manager{
			contextcache:          contextCache,
			salesforceChannel:     make(chan *Message),
			integrationsChannel:   make(chan *Message),
			finishInterconnection: make(chan *Interconnection),
			SalesforceService:     salesforceMock,
			keywordsRestart:       []string{"restart", "test"},
			cacheMessage:          cacheMessage,
			IntegrationsClient:    integrationsIMock,
		}
		go manager.handleInterconnection()

		interconnectionLocal.Set(fmt.Sprintf(constants.UserKey, userID), &Interconnection{
			Status:              Active,
			AffinityToken:       affinityToken,
			SessionKey:          sessionKey,
			SessionID:           sessionID,
			UserID:              userID,
			CaseID:              caseID,
			salesforceChannel:   manager.salesforceChannel,
			integrationsChannel: manager.integrationsChannel,
			finishChannel:       manager.finishInterconnection,
		}, time.Second)
		interconnectionLocal.Wait()

		manager.interconnectionMap = interconnectionLocal

		integrations := &models.IntegrationsRequest{
			ID:        messageID,
			Timestamp: "123456789",
			Type:      imageType,
			From:      userID,
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
	t.Run("ContextByUserID", func(t *testing.T) {
		contextCache := new(ContextCacheMock)
		ctx := []cache.Context{
			{
				UserID:    userID,
				Client:    client,
				Timestamp: 1631202337350,
				Text: `this a test
second line
`,
				From: fromUser,
				Ttl:  time.Now().Add(2 * time.Minute),
			},
			{
				UserID:    userID,
				Client:    client,
				Timestamp: 1630404000000,
				Ttl:       time.Now().Add(2 * time.Minute * -1),
				Text:      "This message should not appear",
				From:      fromUser,
			},
			{
				UserID:    userID,
				Client:    client,
				Timestamp: 1630404000000,
				Ttl:       time.Now().Add(2 * time.Minute),
				Text:      "Hello",
				From:      fromUser,
			},
			{
				Client:    client,
				UserID:    userID,
				Timestamp: 1630404060000,
				Ttl:       time.Now().Add(2 * time.Minute),
				Text:      "Hello I'm a bot",
				From:      fromBot,
			},
			{
				UserID:    userID,
				Client:    client,
				Timestamp: 1630404240000,
				Ttl:       time.Now().Add(2 * time.Minute),
				Text:      "ok.",
				From:      fromBot,
			},
			{
				UserID:    userID,
				Client:    client,
				Timestamp: 1630404120000,
				Text:      "I need help",
				From:      fromUser,
				Ttl:       time.Now().Add(2 * time.Minute),
			},
		}

		userID := "userID"
		contextCache.On("RetrieveContextFromSet", client, userID).Return(ctx)

		manager := &Manager{
			client:       client,
			contextcache: contextCache,
		}

		ctxStr := manager.getContextByUserID(userID)
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

func TestManager_GetContextByUserID(t *testing.T) {
	t.Run("Should Get context by userID and client", func(t *testing.T) {
		ttlExpected := time.Now().Add(2 * time.Minute)
		contextCache := new(ContextCacheMock)
		ctx := []cache.Context{
			{
				UserID:    userID,
				Timestamp: 1631202337350,
				Client:    client,
				Ttl:       ttlExpected,
				Text: `this a test
second line
`,
				From: fromUser,
			},
			{
				UserID:    userID,
				Client:    client,
				Timestamp: 1630404000000,
				Ttl:       time.Now().Add(2 * time.Minute * -1),
				Text:      "This message should not appear",
				From:      fromUser,
			},
			{
				Client:    client,
				Ttl:       ttlExpected,
				UserID:    userID,
				Timestamp: 1630404000000,
				Text:      "Hello",
				From:      fromUser,
			},
			{
				Client:    client,
				Ttl:       ttlExpected,
				UserID:    userID,
				Timestamp: 1630404060000,
				Text:      "Hello I'm a bot",
				From:      fromBot,
			},
			{
				Client:    client,
				Ttl:       ttlExpected,
				UserID:    userID,
				Timestamp: 1630404240000,
				Text:      "ok.",
				From:      fromBot,
			},
			{
				Client:    client,
				Ttl:       ttlExpected,
				UserID:    userID,
				Timestamp: 1630404120000,
				Text:      "I need help",
				From:      fromUser,
			},
		}

		contextCache.On("RetrieveContextFromSet", client, userID).Return(ctx)

		manager := &Manager{
			client:       client,
			contextcache: contextCache,
		}

		expected := []cache.Context{
			{
				Client:    client,
				Ttl:       ttlExpected,
				UserID:    userID,
				Timestamp: 1630404000000,
				Text:      "Hello",
				From:      fromUser,
			},
			{
				Client: client,
				Ttl:    ttlExpected,
				UserID: userID,

				Timestamp: 1630404060000,
				Text:      "Hello I'm a bot",
				From:      fromBot,
			},
			{
				Client:    client,
				Ttl:       ttlExpected,
				UserID:    userID,
				Timestamp: 1630404120000,
				Text:      "I need help",
				From:      fromUser,
			},
			{
				Client:    client,
				Ttl:       ttlExpected,
				UserID:    userID,
				Timestamp: 1630404240000,
				Text:      "ok.",
				From:      fromBot,
			},
			{
				Client:    client,
				Ttl:       ttlExpected,
				UserID:    userID,
				Timestamp: 1631202337350,
				Text: `this a test
second line
`,
				From: fromUser,
			},
		}

		ctxStr := manager.GetContextByUserID(userID)
		assert.Equal(t, expected, ctxStr)
	})
}

func TestManager_SaveContextFB(t *testing.T) {
	interconnectionLocal := cache.New()
	t.Run("Should save context text from user", func(t *testing.T) {
		contextCache := new(ContextCacheMock)
		contextCache.On("StoreContextToSet", mock.Anything).Return(nil).Once()

		cacheMessage := new(IMessageCache)
		cacheMessage.On("IsRepeatedMessage", messageID).Return(false).Once()

		manager := &Manager{
			contextcache:       contextCache,
			cacheMessage:       cacheMessage,
			interconnectionMap: interconnectionLocal,
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

	t.Run("Should save context text from user error StoreContextToSet", func(t *testing.T) {
		contextCache := new(ContextCacheMock)
		contextCache.On("StoreContextToSet", mock.Anything).Return(assert.AnError).Once()

		cacheMessage := new(IMessageCache)
		cacheMessage.On("IsRepeatedMessage", messageID).Return(false).Once()

		manager := &Manager{
			contextcache:       contextCache,
			cacheMessage:       cacheMessage,
			interconnectionMap: interconnectionLocal,
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
			cacheMessage:       cacheMessage,
			interconnectionMap: interconnectionLocal,
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
		contextCache.On("StoreContextToSet", mock.Anything).Return(nil).Once()

		cacheMessage := new(IMessageCache)
		cacheMessage.On("IsRepeatedMessage", messageID).Return(false).Once()

		manager := &Manager{
			contextcache:       contextCache,
			cacheMessage:       cacheMessage,
			interconnectionMap: interconnectionLocal,
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
		defer interconnectionLocal.Clear()
		contextCache := new(ContextCacheMock)
		salesforceMock := new(SalesforceServiceInterface)
		contextCache.On("StoreContextToSet", mock.Anything).Return(nil).Once()

		salesforceMock.On("SendMessage",
			affinityToken, sessionKey, chat.MessagePayload{Text: "text"}).
			Return(false, nil).Once()
		cacheMessage := new(IMessageCache)
		cacheMessage.On("IsRepeatedMessage", messageID).Return(false).Once()

		manager := &Manager{
			client:                client,
			contextcache:          contextCache,
			SalesforceService:     salesforceMock,
			keywordsRestart:       []string{"restart", "test"},
			salesforceChannel:     make(chan *Message),
			finishInterconnection: make(chan *Interconnection),
			integrationsChannel:   make(chan *Message),
			cacheMessage:          cacheMessage,
		}

		go manager.handleInterconnection()

		interconnectionLocal.Set(fmt.Sprintf(constants.UserKey, userID), &Interconnection{
			Status:              Active,
			AffinityToken:       affinityToken,
			SessionKey:          sessionKey,
			salesforceChannel:   manager.salesforceChannel,
			integrationsChannel: manager.integrationsChannel,
			finishChannel:       manager.finishInterconnection,
		}, time.Second)
		interconnectionLocal.Wait()

		manager.interconnectionMap = interconnectionLocal

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
		defer interconnectionLocal.Clear()
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

		interconnectionLocal.Set(fmt.Sprintf(constants.UserKey, userID), &Interconnection{
			Status:              Active,
			AffinityToken:       affinityToken,
			SessionKey:          sessionKey,
			CaseID:              caseID,
			SessionID:           sessionID,
			salesforceChannel:   manager.salesforceChannel,
			integrationsChannel: manager.integrationsChannel,
			finishChannel:       manager.finishInterconnection,
			SalesforceService:   salesforceMock,
		}, time.Second)
		interconnectionLocal.Wait()

		manager.interconnectionMap = interconnectionLocal

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
		defer interconnectionLocal.Clear()
		contextCache := new(ContextCacheMock)
		salesforceMock := new(SalesforceServiceInterface)
		salesforceMock.On("EndChat",
			affinityToken, sessionKey).
			Return(nil).Once()

		cacheMock := &cache.Interconnection{
			UserID:     userID,
			Client:     client,
			SessionID:  sessionID,
			SessionKey: sessionID,
			Status:     string(Active),
		}
		interconnectionCacheMock := new(InterconnectionCache)
		interconnectionCacheMock.On("RetrieveInterconnection",
			cache.Interconnection{
				UserID: userID,
				Client: client,
			}).
			Return(cacheMock, nil).Once()
		cacheMock.Status = string(Closed)
		interconnectionCacheMock.On("StoreInterconnection", *cacheMock).
			Return(nil).Once()

		cacheMessage := new(IMessageCache)
		cacheMessage.On("IsRepeatedMessage", messageID).Return(false).Once()

		manager := &Manager{
			client:                client,
			contextcache:          contextCache,
			SalesforceService:     salesforceMock,
			keywordsRestart:       []string{"restart", "test"},
			salesforceChannel:     make(chan *Message),
			finishInterconnection: make(chan *Interconnection),
			integrationsChannel:   make(chan *Message),
			cacheMessage:          cacheMessage,
		}

		go manager.handleInterconnection()

		interconnectionLocal.Set(fmt.Sprintf(constants.UserKey, userID), &Interconnection{
			Client:               client,
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
		}, time.Second)
		interconnectionLocal.Wait()

		manager.interconnectionMap = interconnectionLocal

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
		defer interconnectionLocal.Clear()
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

		interconnectionLocal.Set(fmt.Sprintf(constants.UserKey, userID), &Interconnection{
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
		}, time.Second)
		interconnectionLocal.Wait()

		manager.interconnectionMap = interconnectionLocal

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

}

func TestManager_RegisterWebhook(t *testing.T) {
	WebhookBaseUrl = "http://localhost"
	WebhookFacebook = "/v1/integrations/facebook/webhook"
	WebhookWhatsapp = "/v1/integrations/whatsapp/webhook"
	t.Run("Register webhook whastapp Succesfull", func(t *testing.T) {
		WAPhone = "waphone"
		integrationsClientMock := new(IntegrationInterface)

		response := &integrations.HealthcheckResponse{
			BotId:   "botID",
			Channel: "outgoing",
			Webhook: WebhookBaseUrl + WebhookWhatsapp,
		}

		payload := integrations.HealthcheckPayload{
			Phone:    WAPhone,
			Webhook:  fmt.Sprintf("%s/v1/integrations/whatsapp/webhook", WebhookBaseUrl),
			Provider: string(WhatsappProvider),
		}

		integrationsClientMock.On("WebhookRegister", payload).Return(response, nil).Once()

		manager := &Manager{
			IntegrationsClient: integrationsClientMock,
		}

		err := manager.RegisterWebhookInIntegrations("whatsapp")
		assert.NoError(t, err)
	})

	t.Run("Register webhook whastapp Fail", func(t *testing.T) {
		WAPhone = "waphone"
		integrationsClientMock := new(IntegrationInterface)

		payload := integrations.HealthcheckPayload{
			Phone:    WAPhone,
			Webhook:  WebhookBaseUrl + WebhookWhatsapp,
			Provider: string(WhatsappProvider),
		}

		integrationsClientMock.On("WebhookRegister", payload).Return(nil, assert.AnError).Once()

		manager := &Manager{
			IntegrationsClient: integrationsClientMock,
		}

		err := manager.RegisterWebhookInIntegrations("whatsapp")
		assert.Error(t, err)
	})

	t.Run("Register webhook facebook Succesfull", func(t *testing.T) {
		FBPhone = "fbphone"
		integrationsClientMock := new(IntegrationInterface)

		response := &integrations.HealthcheckResponse{
			BotId:   "botID",
			Channel: "outgoing",
			Webhook: WebhookBaseUrl + WebhookFacebook,
		}

		payload := integrations.HealthcheckPayload{
			Phone:    FBPhone,
			Webhook:  fmt.Sprintf("%s/v1/integrations/facebook/webhook", WebhookBaseUrl),
			Version:  3,
			Provider: string(FacebookProvider),
		}

		integrationsClientMock.On("WebhookRegister", payload).Return(response, nil).Once()

		manager := &Manager{
			IntegrationsClient: integrationsClientMock,
		}

		err := manager.RegisterWebhookInIntegrations("facebook")
		assert.NoError(t, err)
	})

	t.Run("Register webhook facebook fail", func(t *testing.T) {
		FBPhone = "fbphone"
		integrationsClientMock := new(IntegrationInterface)

		payload := integrations.HealthcheckPayload{
			Phone:    FBPhone,
			Version:  3,
			Webhook:  WebhookBaseUrl + WebhookFacebook,
			Provider: string(FacebookProvider),
		}

		integrationsClientMock.On("WebhookRegister", payload).Return(nil, assert.AnError).Once()

		manager := &Manager{
			IntegrationsClient: integrationsClientMock,
		}

		err := manager.RegisterWebhookInIntegrations("facebook")
		assert.Error(t, err)
	})

	t.Run("Register webhook default value", func(t *testing.T) {
		FBPhone = "fbphone"
		integrationsClientMock := new(IntegrationInterface)

		manager := &Manager{
			IntegrationsClient: integrationsClientMock,
		}

		err := manager.RegisterWebhookInIntegrations("facebooks")
		assert.Error(t, err)
	})
}

func TestManager_StoreInterconnectionInRedis(t *testing.T) {
	t.Run("Fail store interconnection", func(t *testing.T) {
		interconnection := &Interconnection{
			UserID:    "55125421545",
			SessionID: "sessionId",
		}
		intCache := NewInterconectionCache(interconnection)
		interconnectionCacheMock := new(InterconnectionCache)
		interconnectionCacheMock.On("StoreInterconnection", intCache).
			Return(assert.AnError).Once()
		manager := &Manager{
			interconnectionsCache: interconnectionCacheMock,
		}
		expectedLog := "Could not store interconnection with userID"
		var buf bytes.Buffer
		logrus.SetOutput(&buf)
		manager.storeInterconnectionInRedis(interconnection)
		logs := buf.String()
		if !strings.Contains(logs, expectedLog) {
			t.Fatalf("Logs should contain <%s>, but this was found <%s>", expectedLog, logs)
		}
	})
}

func TestManager_GetContextInterconnection(t *testing.T) {
	t.Run("Should get context from user", func(t *testing.T) {
		interconnection := &Interconnection{
			UserID: "55125421545",
		}
		cacheContextMock := new(ContextCacheMock)
		cacheContextMock.On("RetrieveContextFromSet", client, userID).
			Return([]cache.Context{
				{
					UserID:    userID,
					Client:    client,
					Timestamp: 1111111111,
					Text:      "text",
				},
			}).Once()
		manager := &Manager{
			client:       client,
			contextcache: cacheContextMock,
		}
		expectedLog := "Get context of userID"
		var buf bytes.Buffer
		logrus.SetOutput(&buf)
		manager.getContext(interconnection)
		logs := buf.String()
		if !strings.Contains(logs, expectedLog) {
			t.Fatalf("Logs should contain <%s>, but this was found <%s>", expectedLog, logs)
		}
	})
}

func TestManager_RemoveWebhook(t *testing.T) {
	t.Run("Remove webhook whastapp Succesfull", func(t *testing.T) {
		WAPhone = "waphone"
		WebhookBaseUrl = "http://localhost"
		integrationsClientMock := new(IntegrationInterface)

		payload := integrations.RemoveWebhookPayload{
			Phone:    WAPhone,
			Provider: string(WhatsappProvider),
		}

		integrationsClientMock.On("WebhookRemove", payload).Return(true, nil).Once()

		manager := &Manager{
			IntegrationsClient: integrationsClientMock,
		}

		err := manager.RemoveWebhookInIntegrations("whatsapp")
		assert.NoError(t, err)
	})

	t.Run("Remove webhook whastapp Fail", func(t *testing.T) {
		WAPhone = "waphone"
		WebhookBaseUrl = "http://localhost"
		integrationsClientMock := new(IntegrationInterface)

		payload := integrations.RemoveWebhookPayload{
			Phone:    WAPhone,
			Provider: string(WhatsappProvider),
		}

		integrationsClientMock.On("WebhookRemove", payload).Return(false, assert.AnError).Once()

		manager := &Manager{
			IntegrationsClient: integrationsClientMock,
		}

		err := manager.RemoveWebhookInIntegrations("whatsapp")
		assert.Error(t, err)
	})

	t.Run("Remove webhook facebook Succesfull", func(t *testing.T) {
		FBPhone = "fbphone"
		WebhookBaseUrl = "http://localhost"
		integrationsClientMock := new(IntegrationInterface)

		payload := integrations.RemoveWebhookPayload{
			Phone:    FBPhone,
			Provider: string(FacebookProvider),
		}

		integrationsClientMock.On("WebhookRemove", payload).Return(true, nil).Once()

		manager := &Manager{
			IntegrationsClient: integrationsClientMock,
		}

		err := manager.RemoveWebhookInIntegrations("facebook")
		assert.NoError(t, err)
	})

	t.Run("Remove webhook facebook fail", func(t *testing.T) {
		FBPhone = "fbphone"
		WebhookBaseUrl = "http://localhost"
		integrationsClientMock := new(IntegrationInterface)

		payload := integrations.RemoveWebhookPayload{
			Phone:    FBPhone,
			Provider: string(FacebookProvider),
		}

		integrationsClientMock.On("WebhookRemove", payload).Return(false, assert.AnError).Once()

		manager := &Manager{
			IntegrationsClient: integrationsClientMock,
		}

		err := manager.RemoveWebhookInIntegrations("facebook")
		assert.Error(t, err)
	})

	t.Run("Remove webhook default value", func(t *testing.T) {
		FBPhone = "fbphone"
		WebhookBaseUrl = "http://localhost"
		integrationsClientMock := new(IntegrationInterface)

		manager := &Manager{
			IntegrationsClient: integrationsClientMock,
		}

		err := manager.RemoveWebhookInIntegrations("facebooks")
		assert.Error(t, err)
	})
}

func TestManager_CleanPrefixPhoneNumber(t *testing.T) {
	t.Run("No remove code if SfcCodePhoneRemove is empty", func(t *testing.T) {
		phoneNumberExpected := "521124512451"
		interconnection := &Interconnection{PhoneNumber: phoneNumberExpected}
		cleanPrefixPhoneNumber(interconnection)
		assert.Equal(t, phoneNumberExpected, interconnection.PhoneNumber)
	})

	t.Run("Remove code 521 of interconnection", func(t *testing.T) {
		CodePhoneRemove = []string{"521", "52"}
		phoneNumberExpected := "1245124510"
		interconnection := &Interconnection{PhoneNumber: "5211245124510"}
		cleanPrefixPhoneNumber(interconnection)
		assert.Equal(t, phoneNumberExpected, interconnection.PhoneNumber)
	})

	t.Run("Remove code 52 of interconnection", func(t *testing.T) {
		CodePhoneRemove = []string{"521", "52"}
		phoneNumberExpected := "1245124510"
		interconnection := &Interconnection{PhoneNumber: "521245124510"}
		cleanPrefixPhoneNumber(interconnection)
		assert.Equal(t, phoneNumberExpected, interconnection.PhoneNumber)
	})

	t.Run("No remove code 521 if phoneNumber is less than ten digits in interconnection", func(t *testing.T) {
		CodePhoneRemove = []string{"521", "52"}
		phoneNumberExpected := "52124512451"
		interconnection := &Interconnection{PhoneNumber: phoneNumberExpected}
		cleanPrefixPhoneNumber(interconnection)
		assert.Equal(t, phoneNumberExpected, interconnection.PhoneNumber)
	})
}

func TestManager_sendMessageToSalesforce(t *testing.T) {
	message := &Message{
		Text:          "Hola test",
		UserID:        userID,
		SessionKey:    sessionKey,
		AffinityToken: affinityToken,
	}

	t.Run("Should sent message", func(t *testing.T) {
		expectedLog := "Send message to agent from salesforce"
		salesforceServiceMock := new(SalesforceServiceInterface)
		salesforceServiceMock.On("SendMessage", message.AffinityToken, message.SessionKey, chat.MessagePayload{Text: message.Text}).Return(true, nil).Once()
		manager := Manager{
			SalesforceService: salesforceServiceMock,
		}

		var buf bytes.Buffer
		logrus.SetOutput(&buf)
		manager.sendMessageToSalesforce(message)
		logs := buf.String()

		if !strings.Contains(logs, expectedLog) {
			t.Fatalf("Logs should contain <%s>, but this was found <%s>", expectedLog, logs)
		}

	})

	t.Run("Should retry message one time", func(t *testing.T) {
		expectedLog := "Error sendMessage to salesforce"
		salesforceServiceMock := new(SalesforceServiceInterface)
		salesforceServiceMock.On("SendMessage", message.AffinityToken, message.SessionKey, chat.MessagePayload{Text: message.Text}).Return(false, assert.AnError).Once()
		salesforceServiceMock.On("SendMessage", message.AffinityToken, message.SessionKey, chat.MessagePayload{Text: message.Text}).Return(true, nil).Once()
		manager := Manager{
			maxRetries:        1,
			SalesforceService: salesforceServiceMock,
		}

		var buf bytes.Buffer
		logrus.SetOutput(&buf)
		manager.sendMessageToSalesforce(message)
		logs := buf.String()

		if !strings.Contains(logs, expectedLog) {
			t.Fatalf("Logs should contain <%s>, but this was found <%s>", expectedLog, logs)
		}
	})

	t.Run("Should retry message three times", func(t *testing.T) {
		expectedLog := "Error sendMessage to salesforce, max retries"
		salesforceServiceMock := new(SalesforceServiceInterface)
		salesforceServiceMock.On("SendMessage", message.AffinityToken, message.SessionKey, chat.MessagePayload{Text: message.Text}).Return(false, assert.AnError).Once()
		salesforceServiceMock.On("SendMessage", message.AffinityToken, message.SessionKey, chat.MessagePayload{Text: message.Text}).Return(false, assert.AnError).Once()
		salesforceServiceMock.On("SendMessage", message.AffinityToken, message.SessionKey, chat.MessagePayload{Text: message.Text}).Return(false, assert.AnError).Once()
		salesforceServiceMock.On("SendMessage", message.AffinityToken, message.SessionKey, chat.MessagePayload{Text: message.Text}).Return(false, assert.AnError).Once()
		manager := Manager{
			maxRetries:        3,
			SalesforceService: salesforceServiceMock,
		}

		var buf bytes.Buffer
		logrus.SetOutput(&buf)
		manager.sendMessageToSalesforce(message)
		logs := buf.String()

		if !strings.Contains(logs, expectedLog) {
			t.Fatalf("Logs should contain <%s>, but this was found <%s>", expectedLog, logs)
		}
	})

}

func TestManager_sendMessageToUser(t *testing.T) {
	message := &Message{
		Text:     "Hola test",
		UserID:   userID,
		Provider: WhatsappProvider,
	}

	t.Run("Should sent message whats", func(t *testing.T) {
		expectedLog := "Send message to UserID"
		integrationsClient := new(IntegrationInterface)
		integrationsClient.On("SendMessage", mock.Anything, string(message.Provider)).Return(&integrations.SendMessageResponse{}, nil).Once()
		manager := Manager{
			IntegrationsClient: integrationsClient,
		}

		var buf bytes.Buffer
		logrus.SetOutput(&buf)
		manager.sendMessageToUser(message)
		logs := buf.String()

		if !strings.Contains(logs, expectedLog) {
			t.Fatalf("Logs should contain <%s>, but this was found <%s>", expectedLog, logs)
		}

	})

	t.Run("Should retry message one time whats", func(t *testing.T) {
		expectedLog := "Error sendMessage to user"
		integrationsClient := new(IntegrationInterface)
		integrationsClient.On("SendMessage", mock.Anything, string(message.Provider)).Return(&integrations.SendMessageResponse{}, assert.AnError).Once()
		integrationsClient.On("SendMessage", mock.Anything, string(message.Provider)).Return(&integrations.SendMessageResponse{}, nil).Once()
		manager := Manager{
			maxRetries:         1,
			IntegrationsClient: integrationsClient,
		}

		var buf bytes.Buffer
		logrus.SetOutput(&buf)
		manager.sendMessageToUser(message)
		logs := buf.String()

		if !strings.Contains(logs, expectedLog) {
			t.Fatalf("Logs should contain <%s>, but this was found <%s>", expectedLog, logs)
		}
	})

	t.Run("Should retry message three times whats", func(t *testing.T) {
		expectedLog := "Error sendMessage to user, max retries"
		integrationsClient := new(IntegrationInterface)
		integrationsClient.On("SendMessage", mock.Anything, string(message.Provider)).Return(&integrations.SendMessageResponse{}, assert.AnError).Once()
		integrationsClient.On("SendMessage", mock.Anything, string(message.Provider)).Return(&integrations.SendMessageResponse{}, assert.AnError).Once()
		integrationsClient.On("SendMessage", mock.Anything, string(message.Provider)).Return(&integrations.SendMessageResponse{}, assert.AnError).Once()
		integrationsClient.On("SendMessage", mock.Anything, string(message.Provider)).Return(&integrations.SendMessageResponse{}, assert.AnError).Once()
		manager := Manager{
			maxRetries:         3,
			IntegrationsClient: integrationsClient,
		}

		var buf bytes.Buffer
		logrus.SetOutput(&buf)
		manager.sendMessageToUser(message)
		logs := buf.String()

		if !strings.Contains(logs, expectedLog) {
			t.Fatalf("Logs should contain <%s>, but this was found <%s>", expectedLog, logs)
		}
	})

	message.Provider = FacebookProvider
	t.Run("Should sent message fb", func(t *testing.T) {
		expectedLog := "Send message to UserID"
		integrationsClient := new(IntegrationInterface)
		integrationsClient.On("SendMessage", mock.Anything, string(message.Provider)).Return(&integrations.SendMessageResponse{}, nil).Once()
		manager := Manager{
			IntegrationsClient: integrationsClient,
		}

		var buf bytes.Buffer
		logrus.SetOutput(&buf)
		manager.sendMessageToUser(message)
		logs := buf.String()

		if !strings.Contains(logs, expectedLog) {
			t.Fatalf("Logs should contain <%s>, but this was found <%s>", expectedLog, logs)
		}

	})

	t.Run("Should retry message one time fb", func(t *testing.T) {
		expectedLog := "Error sendMessage to user"
		integrationsClient := new(IntegrationInterface)
		integrationsClient.On("SendMessage", mock.Anything, string(message.Provider)).Return(&integrations.SendMessageResponse{}, assert.AnError).Once()
		integrationsClient.On("SendMessage", mock.Anything, string(message.Provider)).Return(&integrations.SendMessageResponse{}, nil).Once()
		manager := Manager{
			maxRetries:         1,
			IntegrationsClient: integrationsClient,
		}

		var buf bytes.Buffer
		logrus.SetOutput(&buf)
		manager.sendMessageToUser(message)
		logs := buf.String()

		if !strings.Contains(logs, expectedLog) {
			t.Fatalf("Logs should contain <%s>, but this was found <%s>", expectedLog, logs)
		}
	})

	t.Run("Should retry message three times fb", func(t *testing.T) {
		expectedLog := "Error sendMessage to user, max retries"
		integrationsClient := new(IntegrationInterface)
		integrationsClient.On("SendMessage", mock.Anything, string(message.Provider)).Return(&integrations.SendMessageResponse{}, assert.AnError).Once()
		integrationsClient.On("SendMessage", mock.Anything, string(message.Provider)).Return(&integrations.SendMessageResponse{}, assert.AnError).Once()
		integrationsClient.On("SendMessage", mock.Anything, string(message.Provider)).Return(&integrations.SendMessageResponse{}, assert.AnError).Once()
		integrationsClient.On("SendMessage", mock.Anything, string(message.Provider)).Return(&integrations.SendMessageResponse{}, assert.AnError).Once()
		manager := Manager{
			maxRetries:         3,
			IntegrationsClient: integrationsClient,
		}

		var buf bytes.Buffer
		logrus.SetOutput(&buf)
		manager.sendMessageToUser(message)
		logs := buf.String()

		if !strings.Contains(logs, expectedLog) {
			t.Fatalf("Logs should contain <%s>, but this was found <%s>", expectedLog, logs)
		}
	})
}

func TestManager_saveContextInRedis(t *testing.T) {
	t.Run("Should save context in redis without error", func(t *testing.T) {
		expectedLog := "Error store context in set"
		contextCache := new(ContextCacheMock)
		contextCache.On("StoreContextToSet", mock.Anything).Return(nil).Once()
		manager := Manager{
			contextcache: contextCache,
		}

		var buf bytes.Buffer
		logrus.SetOutput(&buf)
		manager.saveContextInRedis(&cache.Context{})
		logs := buf.String()
		if strings.Contains(logs, expectedLog) {
			t.Fatalf("Logs should not contain <%s>, but this was found <%s>", expectedLog, logs)
		}
	})

	t.Run("Should save context in redis with error", func(t *testing.T) {
		expectedLog := "Error store context in set"
		contextCache := new(ContextCacheMock)
		contextCache.On("StoreContextToSet", mock.Anything).Return(assert.AnError).Once()
		manager := Manager{
			contextcache: contextCache,
		}

		var buf bytes.Buffer
		logrus.SetOutput(&buf)
		manager.saveContextInRedis(&cache.Context{})
		logs := buf.String()
		if !strings.Contains(logs, expectedLog) {
			t.Fatalf("Logs should not contain <%s>, but this was found <%s>", expectedLog, logs)
		}
	})
}
