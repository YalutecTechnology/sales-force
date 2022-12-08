package manage

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"yalochat.com/salesforce-integration/base/constants"
	"yalochat.com/salesforce-integration/base/events"
	"yalochat.com/salesforce-integration/base/subscribers"
	"yalochat.com/salesforce-integration/base/subscribers/kafka"

	"github.com/sirupsen/logrus"
	"yalochat.com/salesforce-integration/app/services"
	"yalochat.com/salesforce-integration/base/cache"
	"yalochat.com/salesforce-integration/base/clients/botrunner"
	"yalochat.com/salesforce-integration/base/clients/chat"
	"yalochat.com/salesforce-integration/base/clients/integrations"
	"yalochat.com/salesforce-integration/base/clients/studiong"
	"yalochat.com/salesforce-integration/base/helpers"
)

// Status that an interconnection can have
type InterconnectionStatus string
type Provider string

const (
	Failed           InterconnectionStatus = "FAILED"
	OnHold           InterconnectionStatus = "ON_HOLD"
	Active           InterconnectionStatus = "ACTIVE"
	Closed           InterconnectionStatus = "CLOSED"
	WhatsappProvider Provider              = "whatsapp"
	FacebookProvider Provider              = "facebook"
)

// Interconnection struct represents a connection between userBotYalo and salesforce agent
type Interconnection struct {
	UserID               string                              `json:"userId"`
	Client               string                              `json:"client"`
	SessionID            string                              `json:"sessionId"`
	SessionKey           string                              `json:"sessionKey"`
	AffinityToken        string                              `json:"affinityToken"`
	Status               InterconnectionStatus               `json:"status"`
	Timestamp            time.Time                           `json:"timestamp"`
	Provider             Provider                            `json:"provider"`
	BotSlug              string                              `json:"botSlug"`
	BotID                string                              `json:"botId"`
	Name                 string                              `json:"name"`
	Email                string                              `json:"email"`
	PhoneNumber          string                              `json:"phoneNumber"`
	CaseID               string                              `json:"caseId"`
	Context              string                              `json:"-"`
	ExtraData            map[string]interface{}              `json:"extraData"`
	finishChannel        chan *Interconnection               `json:"-"`
	BotrunnnerClient     botrunner.BotRunnerInterface        `json:"-"`
	SalesforceService    services.SalesforceServiceInterface `json:"-"`
	IntegrationsClient   integrations.IntegrationInterface   `json:"-"`
	interconnectionCache cache.IInterconnectionCache         `json:"-"`
	runnigLongPolling    bool                                `json:"-"`
	// This field helps us reconnect the chat in Salesforce.
	offset           int `json:"offset"`
	StudioNG         studiong.StudioNGInterface
	isStudioNGFlow   bool
	kafkaProducer    subscribers.Producer
	KafkaTopic       string
	SleepLongPolling time.Duration
	// ack is a sequencing mechanism that allows you to poll for messages on the Live Agent server
	ack int
}

type InterconnectionMessageQueue struct {
	ID        string       `json:"id"`
	EventType string       `json:"event_type"`
	Params    MessageQueue `json:"params"`
	TraceID   string       `json:"trace_id"`
}

type MessageQueue struct {
	Message `json:"message"`
	Client  string `json:"client"`
}

// Message represents the messages that will be sent through the chat
type Message struct {
	ID            string      `json:"id"`
	MainSpan      tracer.Span `json:"-"`
	Text          string      `json:"text"`
	ImageUrl      string      `json:"imageUrl"`
	UserID        string      `json:"userID"`
	SessionKey    string      `json:"sessionKey"`
	AffinityToken string      `json:"affinityToken"`
	Provider      Provider    `json:"provider"`
}

type NewInterconnectionParams struct {
	UserID      string
	Name        string
	Provider    Provider
	BotSlug     string
	BotID       string
	PhoneNumber string
	Email       string
	ExtraData   map[string]interface{}
	Client      string
}

func NewInterconnection(p *NewInterconnectionParams) *Interconnection {
	i := &Interconnection{
		UserID:      p.UserID,
		Name:        p.Name,
		Provider:    p.Provider,
		BotSlug:     p.BotSlug,
		BotID:       p.BotID,
		PhoneNumber: p.PhoneNumber,
		Email:       p.Email,
		ExtraData:   p.ExtraData,
		Client:      p.Client,
	}
	i.ack = constants.InitialAck
	return i
}

func NewIntegrationsMessage(mainSpan tracer.Span, id, userID, text string, provider Provider) *Message {
	return &Message{
		ID:       id,
		MainSpan: mainSpan,
		UserID:   userID,
		Text:     text,
		Provider: provider,
	}
}

func NewSfMessage(mainSpan tracer.Span, affinityToken, key, text, userID string) *Message {
	return &Message{
		MainSpan:      mainSpan,
		AffinityToken: affinityToken,
		SessionKey:    key,
		Text:          text,
		UserID:        userID,
	}
}

// handleLongPolling It will be a gorutine that is active during the conversation between an end user and a Salesforce
// agent, it launches requests to the endpoint `{{sfChatApi}}/chat/rest/System/Messages`, with which we will know
// the status of the chat, and we will receive messages from the agents to send to the end user
func (in *Interconnection) handleLongPolling() {
	// datadog tracing and logging
	mainSpan := tracer.StartSpan("interconnection.handleLongPolling")
	mainSpan.SetTag(ext.ResourceName, fmt.Sprintf("%s %s", "GET", "/chat/rest/System/Messages"))
	mainSpan.SetTag(ext.AnalyticsEvent, true)
	mainSpan.SetTag(events.Interconnection, fmt.Sprintf("%#v", in))
	mainSpan.SetTag(events.UserID, in.UserID)
	mainSpan.SetTag(events.Client, in.Client)
	mainSpan.SetTag("sessionID", in.SessionID)
	defer mainSpan.Finish()
	logFields := logrus.Fields{
		constants.TraceIdKey: mainSpan.Context().TraceID(),
		constants.SpanIdKey:  mainSpan.Context().SpanID(),
		events.UserID:        in.UserID,
	}
	logrus.WithFields(logFields).Info("Starting long polling service from salesforce...")
	// ---

	in.runnigLongPolling = true
	for in.runnigLongPolling {
		response, errorResponse := in.SalesforceService.
			GetMessages(mainSpan, in.AffinityToken, in.SessionKey, in.ack)
		if errorResponse != nil {
			// fmt.Println("interconnection.errorResponse: ", errorResponse.Error.Error())
			switch errorResponse.StatusCode {
			case http.StatusNoContent:
				logrus.WithFields(logFields).Info("Not content events")
			case http.StatusConflict:
				logrus.WithFields(logFields).Info("Duplicate Long Polling")
				<-time.After(in.SleepLongPolling)
			case http.StatusForbidden:
				go ChangeToState(in.UserID, in.BotSlug, TimeoutState[string(in.Provider)], in.BotrunnnerClient, BotrunnerTimeout, StudioNGTimeout, in.StudioNG, in.isStudioNGFlow)
				in.finishLongPolling(Closed)
				logrus.WithFields(logFields).Error("StatusForbidden")
				mainSpan.SetTag(ext.Error, errorResponse.Error)
				mainSpan.SetTag(events.StatusSalesforce, errorResponse.StatusCode)
			case http.StatusServiceUnavailable:
				logrus.WithFields(logFields).Error("StatusServiceUnavailable")
				mainSpan.SetTag("errorSalesforce", errorResponse.Error.Error())
				mainSpan.SetTag(events.StatusSalesforce, errorResponse.StatusCode)

				reconnect, err := in.SalesforceService.ReconnectSession(in.SessionKey, strconv.Itoa(in.offset))
				if err != nil {
					go ChangeToState(in.UserID, in.BotSlug, TimeoutState[string(in.Provider)], in.BotrunnnerClient, BotrunnerTimeout, StudioNGTimeout, in.StudioNG, in.isStudioNGFlow)
					in.finishLongPolling(Closed)
					logrus.WithFields(logFields).WithError(err).Error("Reconnect session failed")
					mainSpan.SetTag(ext.Error, err)
					continue
				}

				logrus.WithFields(logFields).Info("Reconnect session on long polling")
				in.AffinityToken = reconnect.Messages[0].Message.AffinityToken
				in.updateAffinityTokenRedis(reconnect.Messages[0].Message.AffinityToken)
			default:
				if strings.Contains(errorResponse.Error.Error(), "Client.Timeout exceeded while awaiting headers") {
					//fmt.Println("interconnection.timeOutExceeded: ", errorResponse.Error.Error())
					mainSpan.SetTag("Client.Timeout exceeded while awaiting headers", errorResponse.Error.Error())
					continue
				}

				logrus.WithFields(logFields).Errorf("Exists error in long polling : %s", errorResponse.Error.Error())

				go ChangeToState(
					in.UserID,
					in.BotSlug,
					TimeoutState[string(in.Provider)],
					in.BotrunnnerClient,
					BotrunnerTimeout,
					StudioNGTimeout,
					in.StudioNG,
					in.isStudioNGFlow,
				)
				in.finishLongPolling(Closed)
				mainSpan.SetTag(ext.Error, errorResponse.Error)
				mainSpan.SetTag(events.StatusSalesforce, errorResponse.StatusCode)
			}
			continue
		}

		in.offset = response.Offset
		if response.Sequence != 0 {
			in.ack = response.Sequence
		}
		go func(span tracer.Span) {
			for _, event := range response.Messages {
				in.checkEvent(span, &event)
			}
		}(mainSpan)
		<-time.After(time.Millisecond * 100)
	}
}

func (in *Interconnection) checkEvent(mainSpan tracer.Span, event *chat.MessageObject) {
	// datadog tracing
	spanContext := events.GetSpanContextFromSpan(mainSpan)
	span := tracer.StartSpan(fmt.Sprintf("%s.handleLongPolling.checkEvent", in.UserID), tracer.ChildOf(spanContext))
	span.SetTag(ext.AnalyticsEvent, true)
	span.SetTag("eventType", event.Type)
	span.SetTag("event", fmt.Sprintf("%#v", event))
	defer span.Finish()

	logFields := logrus.Fields{
		constants.TraceIdKey: span.Context().TraceID(),
		constants.SpanIdKey:  span.Context().SpanID(),
		events.UserID:        in.UserID,
	}
	//fmt.Println("Interconnection.checkEvent: ", event.Type)
	switch event.Type {
	case chat.ChatRequestFail:
		logrus.WithFields(logFields).Infof("Event [%s] : [%s]", chat.ChatRequestFail, event.Message.Reason)
		mainSpan.SetTag(ext.Error, fmt.Errorf("event [%s] : [%s]", chat.ChatRequestFail, event.Message.Reason))
		go ChangeToState(in.UserID, in.BotSlug, TimeoutState[string(in.Provider)], in.BotrunnnerClient, BotrunnerTimeout, StudioNGTimeout, in.StudioNG, in.isStudioNGFlow)
		in.finishLongPolling(Failed)
	case chat.ChatRequestSuccess:
		logrus.WithFields(logFields).Infof("Event [%s]", chat.ChatRequestSuccess)
		if Messages.WaitAgent != "" {
			in.sendMessageToQueue(span,
				helpers.RandomString(36),
				Messages.WaitAgent,
				constants.SendMessageToUser)
		}
		//in.integrationsChannel <- NewIntegrationsMessage(in.UserID, fmt.Sprintf("%s : %v", Messages.QueuePosition, event.Message.QueuePosition), in.Provider)
		//in.integrationsChannel <- NewIntegrationsMessage(in.UserID, fmt.Sprintf("%s : %vs", Messages.WaitTime, event.Message.EstimatedWaitTime), in.Provider)
	case chat.ChatEstablished:
		logrus.WithFields(logFields).Infof("Event [%s]", event.Type)
		in.ActiveChat(span)
	case chat.ChatMessage:
		logrus.WithFields(logFields).Infof("Message from salesforce : %s", event.Message.Text)
		in.sendMessageToQueue(span,
			helpers.RandomString(36),
			event.Message.Text,
			constants.SendMessageToUser)
	case chat.QueueUpdate:
		logrus.WithFields(logFields).Infof("Event [%s]", chat.QueueUpdate)
		/*if event.Message.QueuePosition > 0 {
			in.integrationsChannel <- NewIntegrationsMessage(in.UserID, fmt.Sprintf("%s : %v", Messages.QueuePosition, event.Message.QueuePosition), in.Provider)
			in.integrationsChannel <- NewIntegrationsMessage(in.UserID, fmt.Sprintf("%s : %vs", Messages.WaitTime, event.Message.EstimatedWaitTime), in.Provider)
		}*/
	case chat.ChatEnded:
		go ChangeToState(in.UserID, in.BotSlug, SuccessState[string(in.Provider)], in.BotrunnnerClient, 0, 0, in.StudioNG, in.isStudioNGFlow)
		in.finishLongPolling(Closed)
	default:
		logrus.WithFields(logFields).Infof("Event [%s]", event.Type)
	}
}

func (in *Interconnection) ActiveChat(mainSpan tracer.Span) {
	in.Status = Active
	in.updateStatusRedis(string(in.Status))
	if in.Context != "" {
		in.sendMessageToSalesforce(NewSfMessage(mainSpan, in.AffinityToken, in.SessionKey, in.Context, in.UserID))
		mainSpan.SetTag("SendContext", in.Context != "")
		in.Context = ""
	}
	if Messages.WelcomeTemplate != "" {
		in.sendMessageToSalesforce(NewSfMessage(mainSpan, in.AffinityToken, in.SessionKey, fmt.Sprintf(Messages.WelcomeTemplate, in.Name), in.UserID))
	}
}

func convertInterconnectionCacheToInterconnection(interconnection cache.Interconnection) *Interconnection {
	return &Interconnection{
		UserID:        interconnection.UserID,
		Client:        interconnection.Client,
		SessionID:     interconnection.SessionID,
		SessionKey:    interconnection.SessionKey,
		AffinityToken: interconnection.AffinityToken,
		Status:        InterconnectionStatus(interconnection.Status),
		Timestamp:     interconnection.Timestamp,
		Provider:      Provider(interconnection.Provider),
		BotSlug:       interconnection.BotSlug,
		BotID:         interconnection.BotID,
		Name:          interconnection.Name,
		Email:         interconnection.Email,
		PhoneNumber:   interconnection.PhoneNumber,
		CaseID:        interconnection.CaseID,
		ExtraData:     interconnection.ExtraData,
	}
}

func (in *Interconnection) updateStatusRedis(status string) {
	interconnectionCache, err := in.interconnectionCache.RetrieveInterconnection(cache.Interconnection{UserID: in.UserID, Client: in.Client})
	if err != nil {
		logrus.Errorf("Could not update status in interconnection userID[%s]-client[%s] from redis : [%s]", in.UserID, in.Client, err.Error())
		return
	}

	interconnectionCache.Status = status
	err = in.interconnectionCache.StoreInterconnection(*interconnectionCache)
	if err != nil {
		logrus.Errorf("Could not update status in interconnection userID[%s]-client[%s] from redis : [%s]", in.UserID, in.Client, err.Error())
	}
}

func (in *Interconnection) updateAffinityTokenRedis(affinityToken string) {
	interconnectionCache, err := in.interconnectionCache.RetrieveInterconnection(cache.Interconnection{UserID: in.UserID, Client: in.Client})

	if err != nil {
		logrus.Errorf("Could not update affinity token in interconnection userID[%s]-client[%s] from redis : [%s]", in.UserID, in.Client, err.Error())
		return
	}

	interconnectionCache.AffinityToken = affinityToken
	err = in.interconnectionCache.StoreInterconnection(*interconnectionCache)
	if err != nil {
		logrus.Errorf("Could not update affinity token in interconnection userID[%s]-client[%s] from redis : [%s]", in.UserID, in.Client, err.Error())
	}
}

func (in *Interconnection) sendMessageToSalesforce(message *Message) {
	// datadog tracing
	spanContext := events.GetSpanContextFromSpan(message.MainSpan)
	span := tracer.StartSpan("sendMessageToSalesforce", tracer.ChildOf(spanContext))
	span.SetTag(ext.AnalyticsEvent, true)
	span.SetTag(events.MessageSalesforce, message)
	span.SetTag(events.UserID, message.UserID)
	span.SetTag(events.SendMessage, false)
	defer span.Finish()
	_, err := in.SalesforceService.SendMessage(span, message.AffinityToken, message.SessionKey, chat.MessagePayload{Text: message.Text})
	if err != nil {
		span.SetTag(ext.Error, err)
		logrus.WithField(events.UserID, message.UserID).Error(helpers.ErrorMessage("Error sendMessage", err))
	}
	span.SetTag(events.SendMessage, true)
	logrus.Infof("Send message to agent from salesforce : %s", message.UserID)
}

func (in *Interconnection) sendMessageToQueue(mainSpan tracer.Span, messageID, text, eventType string) {
	spanContext := events.GetSpanContextFromSpan(mainSpan)
	span := tracer.StartSpan("send_message_to_queue", tracer.ChildOf(spanContext))
	span.SetTag(ext.AnalyticsEvent, true)
	span.SetTag(events.UserID, in.UserID)
	span.SetTag(events.Client, in.Client)
	span.SetTag(events.Message, text)
	span.SetTag("messageId", messageID)
	span.SetTag(events.EventType, eventType)
	traceID := strconv.FormatUint(span.Context().TraceID(), 10)
	defer span.Finish()

	message := InterconnectionMessageQueue{
		ID:        messageID,
		EventType: eventType,
		Params: MessageQueue{
			Client: in.Client,
			Message: Message{
				Text:          text,
				UserID:        in.UserID,
				SessionKey:    in.SessionKey,
				AffinityToken: in.AffinityToken,
				Provider:      in.Provider,
			},
		},
		TraceID: traceID,
	}
	messageBin, _ := json.Marshal(message)

	messageKafka := kafka.KafkaMessageParams{
		Topic: in.KafkaTopic,
		Msg:   messageBin,
		Key:   constants.DefaultKey,
	}

	span.SetTag(events.MessageKafka, message)

	err := in.kafkaProducer.SendMessage(messageKafka)
	if err != nil {
		span.SetTag(ext.Error, err)
		logrus.WithFields(logrus.Fields{
			"user":            in.UserID,
			"interconnection": in,
		}).WithError(err).Error("error sendMessage to kafka'")

	}
}

func (in *Interconnection) finishLongPolling(status InterconnectionStatus) {
	go in.updateStatusRedis(string(status))
	in.Status = status
	in.runnigLongPolling = false
	in.finishChannel <- in
}
