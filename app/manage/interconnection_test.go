package manage

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"yalochat.com/salesforce-integration/app/manage/mocks"
	"yalochat.com/salesforce-integration/base/helpers"
	"yalochat.com/salesforce-integration/base/models"

	"github.com/go-redis/redis"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"yalochat.com/salesforce-integration/base/cache"
	"yalochat.com/salesforce-integration/base/clients/chat"
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
		interconnectionCache: manager.interconnectionsCache,
		offset:               10000,
		ack:                  -1,
	}
	manager.interconnectionsCache.StoreInterconnection(NewInterconectionCache(interconnection))
	waitCheckEvent = 2 * time.Second

	t.Run("Handle 204 no content", func(t *testing.T) {
		expectedLog := "Not content events"
		mockSalesforceServiceInterface := new(mocks.SalesforceServiceInterface)
		botrunnerMock := new(mocks.BotRunnerInterface)
		mockSalesforceServiceInterface.On("GetMessages", mock.Anything, affinityToken, sessionKey, -1).Return(&chat.MessagesResponse{}, &helpers.ErrorResponse{
			StatusCode: http.StatusNoContent,
		}).Once()
		mockSalesforceServiceInterface.On("GetMessages", mock.Anything, affinityToken, sessionKey, -1).Return(&chat.MessagesResponse{
			Messages: []chat.MessageObject{
				{
					Type: chat.ChatRequestFail,
				},
			},
		}, nil).Once()

		botrunnerMock.On("SendTo", map[string]interface{}{"botSlug": botSlug, "message": "", "state": timeoutState, "userId": userID}).
			Return(true, nil).Once()

		interconnection.SalesforceService = mockSalesforceServiceInterface
		interconnection.BotrunnnerClient = botrunnerMock

		var buf bytes.Buffer
		logrus.SetOutput(&buf)
		interconnection.handleLongPolling()
		logs := buf.String()
		if !strings.Contains(logs, expectedLog) {
			t.Fatalf("Logs should contain <%s>, but this was found <%s>", expectedLog, logs)
		}
	})

	t.Run("Handle 204 no content studiong success", func(t *testing.T) {
		expectedLog := "Not content events"
		mockSalesforceServiceInterface := new(mocks.SalesforceServiceInterface)
		studioNGMock := new(mocks.StudioNGInterface)
		mockSalesforceServiceInterface.On("GetMessages", mock.Anything, affinityToken, sessionKey, -1).Return(&chat.MessagesResponse{}, &helpers.ErrorResponse{
			StatusCode: http.StatusNoContent,
		}).Once()
		mockSalesforceServiceInterface.On("GetMessages", mock.Anything, affinityToken, sessionKey, -1).Return(&chat.MessagesResponse{
			Messages: []chat.MessageObject{
				{
					Type: chat.ChatEnded,
				},
			},
		}, nil).Once()

		studioNGMock.On("SendTo", successState, userID).
			Return(nil).Once()

		interconnection.SalesforceService = mockSalesforceServiceInterface
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

	t.Run("Handle 204 no content studiong error client", func(t *testing.T) {
		expectedLog := "Not content events"
		mockSalesforceServiceInterface := new(mocks.SalesforceServiceInterface)
		studioNGMock := new(mocks.StudioNGInterface)
		mockSalesforceServiceInterface.On("GetMessages", mock.Anything, affinityToken, sessionKey, -1).Return(&chat.MessagesResponse{}, &helpers.ErrorResponse{
			StatusCode: http.StatusNoContent,
		}).Once()
		mockSalesforceServiceInterface.On("GetMessages", mock.Anything, affinityToken, sessionKey, -1).Return(&chat.MessagesResponse{
			Messages: []chat.MessageObject{
				{
					Type: chat.ChatEnded,
				},
			},
		}, nil).Once()

		studioNGMock.On("SendTo", successState, userID).
			Return(assert.AnError).Once()

		interconnection.SalesforceService = mockSalesforceServiceInterface
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
		mockSalesforceServiceInterface := new(mocks.SalesforceServiceInterface)
		interconnection.SalesforceService = mockSalesforceServiceInterface
		mockSalesforceServiceInterface.On("GetMessages", mock.Anything, affinityToken, sessionKey, -1).Return(&chat.MessagesResponse{}, &helpers.ErrorResponse{
			StatusCode: http.StatusForbidden,
		}).Once()

		botrunnerMock := new(mocks.BotRunnerInterface)
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

	t.Run("Handle Reconnect session when response is Status Service Unavailable", func(t *testing.T) {
		expectedLog := "Reconnect session on long polling"
		expectedAffinityToken := "newAffinityToken"
		mockSalesforceServiceInterface := new(mocks.SalesforceServiceInterface)
		interconnection.SalesforceService = mockSalesforceServiceInterface

		mockSalesforceServiceInterface.
			On("GetMessages", mock.Anything, affinityToken, sessionKey, -1).
			Return(
				&chat.MessagesResponse{},
				&helpers.ErrorResponse{
					StatusCode: http.StatusServiceUnavailable,
					Error:      assert.AnError,
				},
			).Once()
		mockSalesforceServiceInterface.
			On("ReconnectSession", sessionKey, strconv.Itoa(interconnection.offset)).
			Return(&chat.MessagesResponse{
				Messages: []chat.MessageObject{
					{
						Type: chat.ReconnectSession,
						Message: chat.Message{
							AffinityToken: expectedAffinityToken,
						},
					},
				},
			}, nil).
			Once()
		mockSalesforceServiceInterface.
			On("GetMessages", mock.Anything, expectedAffinityToken, sessionKey, -1).
			Return(&chat.MessagesResponse{}, &helpers.ErrorResponse{
				StatusCode: http.StatusInternalServerError,
				Error:      assert.AnError,
			}).Once()

		botrunnerMock := new(mocks.BotRunnerInterface)
		botrunnerMock.
			On("SendTo", map[string]interface{}{"botSlug": botSlug, "message": "", "state": timeoutState, "userId": userID}).
			Return(true, nil).
			Once()

		interconnection.BotrunnnerClient = botrunnerMock

		var buf bytes.Buffer
		logrus.SetOutput(&buf)

		go interconnection.handleLongPolling()
		time.Sleep(4 * waitCheckEvent)

		logs := buf.String()
		if !strings.Contains(logs, expectedLog) {
			t.Fatalf("Logs should contain <%s>, but this was found <%s>", expectedLog, logs)
		}
		assert.Equal(t, Closed, interconnection.Status)
		assert.Equal(t, false, interconnection.runnigLongPolling)
		assert.Equal(t, expectedAffinityToken, interconnection.AffinityToken)

		inRedis, _ := interconnection.interconnectionCache.RetrieveInterconnection(cache.Interconnection{UserID: interconnection.UserID, Client: interconnection.Client})
		assert.Equal(t, expectedAffinityToken, inRedis.AffinityToken)
	})

	t.Run("Handle Reconnect Error when response is Status Service Unavailable", func(t *testing.T) {
		interconnection.AffinityToken = affinityToken
		logError := "Reconnect session on long polling"
		expectedLog := "Reconnect session failed"
		mockSalesforceServiceInterface := new(mocks.SalesforceServiceInterface)
		interconnection.SalesforceService = mockSalesforceServiceInterface

		mockSalesforceServiceInterface.On("GetMessages", mock.Anything, affinityToken, sessionKey, -1).Return(&chat.MessagesResponse{}, &helpers.ErrorResponse{
			StatusCode: http.StatusServiceUnavailable,
			Error:      assert.AnError,
		}).Once().
			On("ReconnectSession", sessionKey, strconv.Itoa(interconnection.offset)).Return(&chat.MessagesResponse{}, assert.AnError).Once()

		botrunnerMock := new(mocks.BotRunnerInterface)
		botrunnerMock.On("SendTo", map[string]interface{}{"botSlug": botSlug, "message": "", "state": timeoutState, "userId": userID}).
			Return(true, nil).Once()
		interconnection.BotrunnnerClient = botrunnerMock

		var buf bytes.Buffer
		logrus.SetOutput(&buf)
		interconnection.handleLongPolling()
		logs := buf.String()
		if strings.Contains(logs, logError) {
			t.Fatalf("Logs should not contain <%s>, but this was found <%s>", logError, logs)
		}

		if !strings.Contains(logs, expectedLog) {
			t.Fatalf("Logs should contain <%s>, but this was found <%s>", expectedLog, logs)
		}
		assert.Equal(t, Closed, interconnection.Status)
		assert.Equal(t, false, interconnection.runnigLongPolling)
	})

	t.Run("Handle Other error", func(t *testing.T) {
		interconnection.AffinityToken = affinityToken
		expectedLog := "Exists error in long polling"
		mockSalesforceServiceInterface := new(mocks.SalesforceServiceInterface)
		interconnection.SalesforceService = mockSalesforceServiceInterface

		mockSalesforceServiceInterface.On("GetMessages", mock.Anything, affinityToken, sessionKey, -1).Return(&chat.MessagesResponse{}, &helpers.ErrorResponse{
			StatusCode: http.StatusBadGateway,
			Error:      assert.AnError,
		}).Once()

		botrunnerMock := new(mocks.BotRunnerInterface)
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

	t.Run("Handle StatusConflict  error client", func(t *testing.T) {
		expectedLog := "Duplicate Long Polling"
		mockSalesforceServiceInterface := new(mocks.SalesforceServiceInterface)
		studioNGMock := new(mocks.StudioNGInterface)
		mockSalesforceServiceInterface.On("GetMessages", mock.Anything, affinityToken, sessionKey, -1).Return(&chat.MessagesResponse{}, &helpers.ErrorResponse{
			StatusCode: http.StatusConflict,
		}).Once()
		mockSalesforceServiceInterface.On("GetMessages", mock.Anything, affinityToken, sessionKey, -1).Return(&chat.MessagesResponse{
			Messages: []chat.MessageObject{
				{
					Type: chat.ChatEnded,
				},
			},
		}, nil).Once()

		studioNGMock.On("SendTo", successState, userID).
			Return(assert.AnError).Once()

		interconnection.SalesforceService = mockSalesforceServiceInterface
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

	t.Run("Handle 200 by setting the ack param to te sequence value in the next request", func(t *testing.T) {
		salesforceServiceMock := new(mocks.SalesforceServiceInterface)
		studioNGMock := new(mocks.StudioNGInterface)

		interconnection.SalesforceService = salesforceServiceMock
		interconnection.BotrunnnerClient = nil
		interconnection.isStudioNGFlow = true
		interconnection.StudioNG = studioNGMock
		interconnection.ack = -1

		salesforceServiceMock.
			On("GetMessages", mock.Anything, affinityToken, sessionKey, -1).
			Return(&chat.MessagesResponse{Sequence: 1}, nil).
			Once()

		salesforceServiceMock.
			On("GetMessages", mock.Anything, affinityToken, sessionKey, 1).
			Return(&chat.MessagesResponse{Sequence: 2}, nil).
			Once()

		salesforceServiceMock.
			On("GetMessages", mock.Anything, affinityToken, sessionKey, 2).
			Return(&chat.MessagesResponse{Sequence: 3}, nil).
			Once()

		salesforceServiceMock.
			On("GetMessages", mock.Anything, affinityToken, sessionKey, mock.Anything).
			Return(&chat.MessagesResponse{Sequence: 3}, nil)

		assert.Equal(t, -1, interconnection.ack)
		go interconnection.handleLongPolling()
		time.Sleep(4 * waitCheckEvent)
		interconnection.runnigLongPolling = false
		assert.Equal(t, 3, interconnection.ack)
	})
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
		interconnectionCache: manager.interconnectionsCache,
	}
	manager.interconnectionsCache.StoreInterconnection(NewInterconectionCache(interconnection))

	t.Run("Chat request success event received", func(t *testing.T) {
		Messages = models.MessageTemplate{WaitAgent: "Esperando Agente"}
		expectedLog := chat.ChatRequestSuccess
		event := chat.MessageObject{
			Type:    chat.ChatRequestSuccess,
			Message: chat.Message{},
		}

		producerMock := new(mocks.Producer)
		producerMock.On("SendMessage", mock.Anything).Return(nil).Times(5)
		interconnection.kafkaProducer = producerMock

		var buf bytes.Buffer
		logrus.SetOutput(&buf)
		span, _ := tracer.SpanFromContext(context.Background())
		interconnection.checkEvent(span, &event)
		logs := buf.String()
		if !strings.Contains(logs, expectedLog) {
			t.Fatalf("Logs should contain <%s>, but this was found <%s>", expectedLog, logs)
		}
	})

	t.Run("Chat request success event received error SendMessage", func(t *testing.T) {
		Messages = models.MessageTemplate{WaitAgent: "Esperando Agente"}
		expectedLog := chat.ChatRequestSuccess
		event := chat.MessageObject{
			Type:    chat.ChatRequestSuccess,
			Message: chat.Message{},
		}

		producerMock := new(mocks.Producer)
		producerMock.On("SendMessage", mock.Anything).Return(assert.AnError).Times(5)
		interconnection.kafkaProducer = producerMock

		var buf bytes.Buffer
		logrus.SetOutput(&buf)
		span, _ := tracer.SpanFromContext(context.Background())
		interconnection.checkEvent(span, &event)
		logs := buf.String()
		if !strings.Contains(logs, expectedLog) {
			t.Fatalf("Logs should contain <%s>, but this was found <%s>", expectedLog, logs)
		}
	})

	t.Run("Chat Established event received", func(t *testing.T) {
		Messages = models.MessageTemplate{WelcomeTemplate: "Hola soy %s y necesito ayuda"}
		interconnection.Context = "Contexto"
		expectedLog := chat.ChatEstablished
		mockSalesforceServiceInterface := new(mocks.SalesforceServiceInterface)
		mockSalesforceServiceInterface.On("SendMessage", mock.Anything, affinityToken, sessionKey, chat.MessagePayload{Text: interconnection.Context}).
			Return(true, nil).Once()
		mockSalesforceServiceInterface.On("SendMessage", mock.Anything, affinityToken, sessionKey, chat.MessagePayload{Text: fmt.Sprintf(Messages.WelcomeTemplate, interconnection.Name)}).
			Return(false, assert.AnError).Once()

		interconnection.SalesforceService = mockSalesforceServiceInterface
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
		span, _ := tracer.SpanFromContext(context.Background())
		interconnection.checkEvent(span, &event)
		logs := buf.String()
		if !strings.Contains(logs, expectedLog) {
			t.Fatalf("Logs should contain <%s>, but this was found <%s>", expectedLog, logs)
		}

		assert.Equal(t, Active, interconnection.Status)
	})

	t.Run("Chat Established event received without context and messageTemplate", func(t *testing.T) {
		Messages = models.MessageTemplate{WaitAgent: "Esperando Agente"}
		interconnection.Context = ""
		expectedLog := chat.ChatEstablished
		mockSalesforceServiceInterface := new(mocks.SalesforceServiceInterface)
		mockSalesforceServiceInterface.On("SendMessage", mock.Anything, affinityToken, sessionKey, chat.MessagePayload{Text: interconnection.Context}).
			Return(true, nil).Once()
		mockSalesforceServiceInterface.On("SendMessage", mock.Anything, affinityToken, sessionKey, chat.MessagePayload{Text: fmt.Sprintf(Messages.WelcomeTemplate, interconnection.Name)}).
			Return(false, assert.AnError).Once()

		interconnection.SalesforceService = mockSalesforceServiceInterface
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
		span, _ := tracer.SpanFromContext(context.Background())
		interconnection.checkEvent(span, &event)
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
		span, _ := tracer.SpanFromContext(context.Background())
		interconnection.checkEvent(span, &event)
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
		span, _ := tracer.SpanFromContext(context.Background())
		interconnection.checkEvent(span, &event)
		logs := buf.String()
		if !strings.Contains(logs, expectedLog) {
			t.Fatalf("Logs should contain <%s>, but this was found <%s>", expectedLog, logs)
		}

		assert.Equal(t, Active, interconnection.Status)
	})

	t.Run("Chat fail event received", func(t *testing.T) {
		expectedLog := "Event [ChatRequestFail] : [Unavailable]"

		botrunnerMock := new(mocks.BotRunnerInterface)
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
				Reason:              "Unavailable",
			},
		}

		var buf bytes.Buffer
		logrus.SetOutput(&buf)
		span, _ := tracer.SpanFromContext(context.Background())
		interconnection.checkEvent(span, &event)
		logs := buf.String()
		if !strings.Contains(logs, expectedLog) {
			t.Fatalf("Logs should contain <%s>, but this was found <%s>", expectedLog, logs)
		}

		assert.Equal(t, Failed, interconnection.Status)
	})

	t.Run("Default event received", func(t *testing.T) {
		expectedLog := fmt.Sprintf("Event [%s]", chat.AgentTyping)

		event := chat.MessageObject{
			Type: chat.AgentTyping,
			Message: chat.Message{
				Name: "Name Agent",
			},
		}

		var buf bytes.Buffer
		logrus.SetOutput(&buf)
		span, _ := tracer.SpanFromContext(context.Background())
		interconnection.checkEvent(span, &event)
		logs := buf.String()
		if !strings.Contains(logs, expectedLog) {
			t.Fatalf("Logs should contain <%s>, but this was found <%s>", expectedLog, logs)
		}
	})
}

func TestInterconnection_updateStatusRedis(t *testing.T) {

	t.Run("Update status redis without error ", func(t *testing.T) {
		interconnection := Interconnection{
			Client: client,
			UserID: userID,
			Status: OnHold,
		}
		interconnectionCache := new(mocks.IInterconnectionCache)
		interconnectionCache.On("RetrieveInterconnection", cache.Interconnection{UserID: interconnection.UserID, Client: interconnection.Client}).
			Return(&cache.Interconnection{UserID: interconnection.UserID, Client: interconnection.Client, Status: ""}, nil).Once()
		interconnectionCache.On("StoreInterconnection", mock.Anything).
			Return(nil).Once()
		interconnection.interconnectionCache = interconnectionCache

		var buf bytes.Buffer
		logrus.New().SetOutput(&buf)

		interconnection.updateStatusRedis(string(Active))
		logs := buf.String()
		assert.Zero(t, logs)
	})

	t.Run("Update status redis with error in RetrieveInterconnection", func(t *testing.T) {
		interconnection := Interconnection{
			Client: client,
			UserID: userID,
			Status: OnHold,
		}
		interconnectionCache := new(mocks.IInterconnectionCache)
		expectedLog := "Could not update status in interconnection"
		interconnectionCache.On("RetrieveInterconnection", cache.Interconnection{UserID: userID, Client: client}).
			Return(&cache.Interconnection{UserID: interconnection.UserID, Client: interconnection.Client, Status: ""}, assert.AnError).Once()
		interconnection.interconnectionCache = interconnectionCache

		var buf bytes.Buffer
		logrus.SetOutput(&buf)
		interconnection.updateStatusRedis(string(Active))
		logs := buf.String()
		if !strings.Contains(logs, expectedLog) {
			t.Fatalf("Logs should not contain <%s>, but this was found <%s>", expectedLog, logs)
		}
	})

	t.Run("Update status redis with error in StoreInterconnection", func(t *testing.T) {
		interconnection := Interconnection{
			Client: client,
			UserID: userID,
			Status: OnHold,
		}
		expectedLog := "Could not update status in interconnection"
		interconnectionCache := new(mocks.IInterconnectionCache)
		interconnectionCache.On("RetrieveInterconnection", cache.Interconnection{UserID: userID, Client: client}).
			Return(&cache.Interconnection{UserID: interconnection.UserID, Client: interconnection.Client, Status: ""}, nil).Once()
		interconnectionCache.On("StoreInterconnection", mock.Anything).
			Return(assert.AnError).Once()
		interconnection.interconnectionCache = interconnectionCache

		var buf bytes.Buffer
		logrus.SetOutput(&buf)
		interconnection.updateStatusRedis(string(Active))
		logs := buf.String()
		if !strings.Contains(logs, expectedLog) {
			t.Fatalf("Logs should not contain <%s>, but this was found <%s>", expectedLog, logs)
		}
	})
}

func TestInterconnection_updateAffinityTokenRedis(t *testing.T) {
	interconnection := Interconnection{
		Client: client,
		UserID: userID,
		Status: OnHold,
	}

	t.Run("Update affinity token redis without error ", func(t *testing.T) {
		interconnectionCache := new(mocks.IInterconnectionCache)
		expectedLog := "Could not update affinity token in interconnection"
		interconnectionCache.On("RetrieveInterconnection", cache.Interconnection{UserID: interconnection.UserID, Client: interconnection.Client}).
			Return(&cache.Interconnection{UserID: interconnection.UserID, Client: interconnection.Client, Status: ""}, nil).Once()
		interconnectionCache.On("StoreInterconnection", mock.Anything).
			Return(nil).Once()
		interconnection.interconnectionCache = interconnectionCache

		var buf bytes.Buffer
		logrus.SetOutput(&buf)
		interconnection.updateAffinityTokenRedis(affinityToken)
		logs := buf.String()
		if strings.Contains(logs, expectedLog) {
			t.Fatalf("Logs should not contain <%s>, but this was found <%s>", expectedLog, logs)
		}
	})

	t.Run("Update affinity token redis with error in RetrieveInterconnection", func(t *testing.T) {
		interconnectionCache := new(mocks.IInterconnectionCache)
		expectedLog := "Could not update affinity token in interconnection"
		interconnectionCache.On("RetrieveInterconnection", cache.Interconnection{UserID: userID, Client: client}).
			Return(&cache.Interconnection{UserID: interconnection.UserID, Client: interconnection.Client, Status: ""}, assert.AnError).Once()
		interconnection.interconnectionCache = interconnectionCache

		var buf bytes.Buffer
		logrus.SetOutput(&buf)
		interconnection.updateAffinityTokenRedis(affinityToken)
		logs := buf.String()
		if !strings.Contains(logs, expectedLog) {
			t.Fatalf("Logs should not contain <%s>, but this was found <%s>", expectedLog, logs)
		}
	})

	t.Run("Update affinity token redis with error in StoreInterconnection", func(t *testing.T) {
		expectedLog := "Could not update affinity token in interconnection"
		interconnectionCache := new(mocks.IInterconnectionCache)
		interconnectionCache.On("RetrieveInterconnection", cache.Interconnection{UserID: userID, Client: client}).
			Return(&cache.Interconnection{UserID: interconnection.UserID, Client: interconnection.Client, Status: ""}, nil).Once()
		interconnectionCache.On("StoreInterconnection", mock.Anything).
			Return(assert.AnError).Once()
		interconnection.interconnectionCache = interconnectionCache

		var buf bytes.Buffer
		logrus.SetOutput(&buf)
		interconnection.updateAffinityTokenRedis(affinityToken)
		logs := buf.String()
		if !strings.Contains(logs, expectedLog) {
			t.Fatalf("Logs should not contain <%s>, but this was found <%s>", expectedLog, logs)
		}
	})
}
