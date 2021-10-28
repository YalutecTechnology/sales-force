package manage

import (
	"bytes"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"yalochat.com/salesforce-integration/base/cache"
	"yalochat.com/salesforce-integration/base/clients/chat"
	"yalochat.com/salesforce-integration/base/helpers"
)

const (
	userID        = "55125421545"
	affinityToken = "affinityToken"
	sessionKey    = "sessionKey"
	sessionID     = "sessionID"
	successState  = "from-sf-success"
	timeoutState  = "from-sf-timeout"
)

func TestHandleLongPolling_test(t *testing.T) {
	m, s := cache.CreateRedisServer()
	defer m.Close()
	defer s.Close()

	config := &ManagerOptions{
		AppName: "salesforce-integration",
		RedisOptions: cache.RedisOptions{
			FailOverOptions: &redis.FailoverOptions{
				MasterName:    s.MasterInfo().Name,
				SentinelAddrs: []string{s.Addr()},
			},
			SessionsTTL: time.Second,
		},
		SpecSchedule: "@every 1h30m",
	}
	manager := CreateManager(config)
	SuccessState = map[string]string{
		string(WhatsappProvider): successState,
		string(FacebookProvider): successState,
	}
	TimeoutState = map[string]string{
		string(WhatsappProvider): timeoutState,
		string(FacebookProvider): timeoutState,
	}
	interconnection := &Interconnection{
		UserID:               userID,
		BotSlug:              botSlug,
		SessionID:            sessionID,
		SessionKey:           sessionKey,
		AffinityToken:        affinityToken,
		Status:               OnHold,
		Provider:             provider,
		finishChannel:        manager.finishInterconnection,
		integrationsChannel:  manager.integrationsChannel,
		salesforceChannel:    manager.salesforceChannel,
		interconnectionCache: manager.interconnectionsCache,
	}
	manager.interconnectionsCache.StoreInterconnection(NewInterconectionCache(interconnection))

	t.Run("Handle 204 not content", func(t *testing.T) {
		expectedLog := "Not content events"
		mock := new(SalesforceServiceInterface)
		botrunnerMock := new(BotRunnerInterface)
		mock.On("GetMessages", affinityToken, sessionKey).Return(&chat.MessagesResponse{}, &helpers.ErrorResponse{
			StatusCode: http.StatusNoContent,
		}).Once()
		mock.On("GetMessages", affinityToken, sessionKey).Return(&chat.MessagesResponse{
			Messages: []chat.MessageObject{
				{
					Type: chat.ChatEnded,
				},
			},
		}, nil).Once()

		botrunnerMock.On("SendTo", map[string]interface{}{"botSlug": botSlug, "message": "", "state": successState, "userId": userID}).
			Return(true, nil).Once()

		interconnection.SalesforceService = mock
		interconnection.BotrunnnerClient = botrunnerMock

		var buf bytes.Buffer
		logrus.SetOutput(&buf)
		interconnection.handleLongPolling()
		logs := buf.String()
		if !strings.Contains(logs, expectedLog) {
			t.Fatalf("Logs should contain <%s>, but this was found <%s>", expectedLog, logs)
		}
	})

	t.Run("Handle 204 not content studiong success", func(t *testing.T) {
		expectedLog := "Not content events"
		mock := new(SalesforceServiceInterface)
		studioNGMock := new(StudioNGInterface)
		mock.On("GetMessages", affinityToken, sessionKey).Return(&chat.MessagesResponse{}, &helpers.ErrorResponse{
			StatusCode: http.StatusNoContent,
		}).Once()
		mock.On("GetMessages", affinityToken, sessionKey).Return(&chat.MessagesResponse{
			Messages: []chat.MessageObject{
				{
					Type: chat.ChatEnded,
				},
			},
		}, nil).Once()

		studioNGMock.On("SendTo", successState, userID).
			Return(nil).Once()

		interconnection.SalesforceService = mock
		interconnection.BotrunnnerClient = nil
		interconnection.isStudioNGFlow = true
		interconnection.StudioNG = studioNGMock

		var buf bytes.Buffer
		logrus.SetOutput(&buf)
		interconnection.handleLongPolling()
		logs := buf.String()
		if !strings.Contains(logs, expectedLog) {
			t.Fatalf("Logs should contain <%s>, but this was found <%s>", expectedLog, logs)
		}
	})

	t.Run("Handle 204 not content studiong error client", func(t *testing.T) {
		expectedLog := "Not content events"
		mock := new(SalesforceServiceInterface)
		studioNGMock := new(StudioNGInterface)
		mock.On("GetMessages", affinityToken, sessionKey).Return(&chat.MessagesResponse{}, &helpers.ErrorResponse{
			StatusCode: http.StatusNoContent,
		}).Once()
		mock.On("GetMessages", affinityToken, sessionKey).Return(&chat.MessagesResponse{
			Messages: []chat.MessageObject{
				{
					Type: chat.ChatEnded,
				},
			},
		}, nil).Once()

		studioNGMock.On("SendTo", successState, userID).
			Return(assert.AnError).Once()

		interconnection.SalesforceService = mock
		interconnection.BotrunnnerClient = nil
		interconnection.isStudioNGFlow = true
		interconnection.StudioNG = studioNGMock
		defer func() {
			interconnection.isStudioNGFlow = false
		}()

		var buf bytes.Buffer
		logrus.SetOutput(&buf)
		interconnection.handleLongPolling()
		logs := buf.String()
		if !strings.Contains(logs, expectedLog) {
			t.Fatalf("Logs should contain <%s>, but this was found <%s>", expectedLog, logs)
		}
	})

	t.Run("Handle Status Forbidden", func(t *testing.T) {
		expectedLog := "StatusForbidden"
		mock := new(SalesforceServiceInterface)
		interconnection.SalesforceService = mock
		mock.On("GetMessages", affinityToken, sessionKey).Return(&chat.MessagesResponse{}, &helpers.ErrorResponse{
			StatusCode: http.StatusForbidden,
		}).Once()

		botrunnerMock := new(BotRunnerInterface)
		botrunnerMock.On("SendTo", map[string]interface{}{"botSlug": botSlug, "message": "", "state": timeoutState, "userId": userID}).
			Return(true, nil).Once()
		interconnection.BotrunnnerClient = botrunnerMock

		var buf bytes.Buffer
		logrus.SetOutput(&buf)
		interconnection.handleLongPolling()
		logs := buf.String()
		if !strings.Contains(logs, expectedLog) {
			t.Fatalf("Logs should contain <%s>, but this was found <%s>", expectedLog, logs)
		}

		assert.Equal(t, Closed, interconnection.Status)
		assert.Equal(t, false, interconnection.runnigLongPolling)
	})

	t.Run("Handle Status Service Unavailable", func(t *testing.T) {
		expectedLog := "StatusServiceUnavailable"
		mock := new(SalesforceServiceInterface)
		interconnection.SalesforceService = mock

		mock.On("GetMessages", affinityToken, sessionKey).Return(&chat.MessagesResponse{}, &helpers.ErrorResponse{
			StatusCode: http.StatusServiceUnavailable,
		}).Once()

		botrunnerMock := new(BotRunnerInterface)
		botrunnerMock.On("SendTo", map[string]interface{}{"botSlug": botSlug, "message": "", "state": timeoutState, "userId": userID}).
			Return(true, nil).Once()
		interconnection.BotrunnnerClient = botrunnerMock

		var buf bytes.Buffer
		logrus.SetOutput(&buf)
		interconnection.handleLongPolling()
		logs := buf.String()
		if !strings.Contains(logs, expectedLog) {
			t.Fatalf("Logs should contain <%s>, but this was found <%s>", expectedLog, logs)
		}
		assert.Equal(t, Closed, interconnection.Status)
		assert.Equal(t, false, interconnection.runnigLongPolling)
	})

	t.Run("Handle Other error", func(t *testing.T) {
		expectedLog := "Exists error in long polling"
		mock := new(SalesforceServiceInterface)
		interconnection.SalesforceService = mock

		mock.On("GetMessages", affinityToken, sessionKey).Return(&chat.MessagesResponse{}, &helpers.ErrorResponse{
			StatusCode: http.StatusBadGateway,
			Error:      assert.AnError,
		}).Once()

		var buf bytes.Buffer
		logrus.SetOutput(&buf)
		interconnection.handleLongPolling()
		logs := buf.String()
		if !strings.Contains(logs, expectedLog) {
			t.Fatalf("Logs should contain <%s>, but this was found <%s>", expectedLog, logs)
		}
		assert.Equal(t, Closed, interconnection.Status)
		assert.Equal(t, false, interconnection.runnigLongPolling)
	})

	// t.Run("Handle Messages from salesforce", func(t *testing.T) {
	// 	expectedLog := "Get Messages sucessfully"
	// 	mock := new(SalesforceServiceInterface)
	// 	interconnection.SalesforceService = mock

	// 	mock.On("GetMessages", affinityToken, sessionKey).Return(&chat.MessagesResponse{
	// 		Messages: []chat.MessageObject{
	// 			{
	// 				Type: chat.ChatEnded,
	// 			},
	// 		},
	// 	}, nil).Once()

	// 	var buf bytes.Buffer
	// 	logrus.SetOutput(&buf)
	// 	interconnection.handleLongPolling()
	// 	logs := buf.String()
	// 	if !strings.Contains(logs, expectedLog) {
	// 		t.Fatalf("Logs should contain <%s>, but this was found <%s>", expectedLog, logs)
	// 	}
	// })
}

func TestCheckEvent_test(t *testing.T) {
	m, s := cache.CreateRedisServer()
	defer m.Close()
	defer s.Close()

	config := &ManagerOptions{
		AppName: "salesforce-integration",
		RedisOptions: cache.RedisOptions{
			FailOverOptions: &redis.FailoverOptions{
				MasterName:    s.MasterInfo().Name,
				SentinelAddrs: []string{s.Addr()},
			},
			SessionsTTL: time.Second,
		},
		SpecSchedule: "@every 1h30m",
	}
	manager := CreateManager(config)
	interconnection := &Interconnection{
		UserID:               userID,
		SessionID:            sessionID,
		SessionKey:           sessionKey,
		AffinityToken:        affinityToken,
		Status:               OnHold,
		Provider:             provider,
		finishChannel:        manager.finishInterconnection,
		integrationsChannel:  manager.integrationsChannel,
		salesforceChannel:    manager.salesforceChannel,
		interconnectionCache: manager.interconnectionsCache,
	}
	manager.interconnectionsCache.StoreInterconnection(NewInterconectionCache(interconnection))

	t.Run("Chat request success event received", func(t *testing.T) {
		expectedLog := chat.ChatRequestSuccess
		event := chat.MessageObject{
			Type:    chat.ChatRequestSuccess,
			Message: chat.Message{},
		}

		var buf bytes.Buffer
		logrus.SetOutput(&buf)
		interconnection.checkEvent(&event)
		logs := buf.String()
		if !strings.Contains(logs, expectedLog) {
			t.Fatalf("Logs should contain <%s>, but this was found <%s>", expectedLog, logs)
		}
	})

	t.Run("Chat Established event received", func(t *testing.T) {
		expectedLog := chat.ChatEstablished
		mock := new(SalesforceServiceInterface)
		mock.On("SendMessage", affinityToken, sessionKey, chat.MessagePayload{Text: interconnection.Context}).
			Return(true, nil).Once()
		mock.On("SendMessage", affinityToken, sessionKey, chat.MessagePayload{Text: fmt.Sprintf("Hola soy %s y necesito ayuda", interconnection.Name)}).
			Return(false, assert.AnError).Once()

		interconnection.SalesforceService = mock
		event := chat.MessageObject{
			Type: chat.ChatEstablished,
			Message: chat.Message{
				Name:                "Name Agent",
				UserId:              "142451",
				ChasitorIdleTimeout: map[string]interface{}{"isEnabled": false},
				GeoLocation:         chat.GeoLocation{},
			},
		}

		var buf bytes.Buffer
		logrus.SetOutput(&buf)
		interconnection.checkEvent(&event)
		logs := buf.String()
		if !strings.Contains(logs, expectedLog) {
			t.Fatalf("Logs should contain <%s>, but this was found <%s>", expectedLog, logs)
		}

		assert.Equal(t, Active, interconnection.Status)
	})

	t.Run("Chat Message event received", func(t *testing.T) {
		expectedMessage := "Message from salesforce"
		event := chat.MessageObject{
			Type: chat.ChatMessage,
			Message: chat.Message{
				Name:        "Name Agent",
				AgentId:     "142451",
				Text:        expectedMessage,
				Schedule:    map[string]interface{}{"responseDelayMilliseconds": 0},
				GeoLocation: chat.GeoLocation{},
			},
		}

		var buf bytes.Buffer
		logrus.SetOutput(&buf)
		interconnection.checkEvent(&event)
		logs := buf.String()
		if !strings.Contains(logs, expectedMessage) {
			t.Fatalf("Logs should contain <%s>, but this was found <%s>", expectedMessage, logs)
		}
	})

	t.Run("Queue update event received", func(t *testing.T) {
		expectedLog := chat.QueueUpdate
		event := chat.MessageObject{
			Type: chat.QueueUpdate,
			Message: chat.Message{
				QueuePosition:     1,
				EstimatedWaitTime: 10,
			},
		}

		var buf bytes.Buffer
		logrus.SetOutput(&buf)
		interconnection.checkEvent(&event)
		logs := buf.String()
		if !strings.Contains(logs, expectedLog) {
			t.Fatalf("Logs should contain <%s>, but this was found <%s>", expectedLog, logs)
		}

		assert.Equal(t, Active, interconnection.Status)
	})

	t.Run("Chat fail event received", func(t *testing.T) {
		expectedLog := chat.ChatRequestFail

		botrunnerMock := new(BotRunnerInterface)
		botrunnerMock.On("SendTo", map[string]interface{}{"botSlug": botSlug, "message": "", "state": timeoutState, "userId": userID}).
			Return(true, nil).Once()
		interconnection.BotrunnnerClient = botrunnerMock
		interconnection.BotSlug = botSlug
		TimeoutState = map[string]string{
			string(WhatsappProvider): timeoutState,
			string(FacebookProvider): timeoutState,
		}

		event := chat.MessageObject{
			Type: chat.ChatRequestFail,
			Message: chat.Message{
				Name:                "Name Agent",
				UserId:              "142451",
				ChasitorIdleTimeout: map[string]interface{}{"isEnabled": false},
				GeoLocation:         chat.GeoLocation{},
			},
		}

		var buf bytes.Buffer
		logrus.SetOutput(&buf)
		interconnection.checkEvent(&event)
		logs := buf.String()
		if !strings.Contains(logs, expectedLog) {
			t.Fatalf("Logs should contain <%s>, but this was found <%s>", expectedLog, logs)
		}

		assert.Equal(t, Failed, interconnection.Status)
	})
}
