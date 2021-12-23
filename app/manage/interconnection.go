package manage

import (
	"errors"
	"fmt"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"net/http"
	"time"
	"yalochat.com/salesforce-integration/base/constants"
	"yalochat.com/salesforce-integration/base/events"

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
	salesforceChannel    chan *Message                       `json:"-"`
	integrationsChannel  chan *Message                       `json:"-"`
	finishChannel        chan *Interconnection               `json:"-"`
	BotrunnnerClient     botrunner.BotRunnerInterface        `json:"-"`
	SalesforceService    services.SalesforceServiceInterface `json:"-"`
	IntegrationsClient   integrations.IntegrationInterface   `json:"-"`
	interconnectionCache cache.InterconnectionCache          `json:"-"`
	runnigLongPolling    bool                                `json:"-"`
	// This field helps us reconnect the chat in Salesforce.
	offset         int `json:"-"`
	StudioNG       studiong.StudioNGInterface
	isStudioNGFlow bool
}

// Message represents the messages that will be sent through the chat
type Message struct {
	MainSpan      tracer.Span `json:"-"`
	Text          string      `json:"text"`
	ImageUrl      string      `json:"imageUrl"`
	UserID        string      `json:"userID"`
	SessionKey    string      `json:"sessionKey"`
	AffinityToken string      `json:"affinityToken"`
	Provider      Provider    `json:"provider"`
}

func NewIntegrationsMessage(mainSpan tracer.Span, userID, text string, provider Provider) *Message {
	return &Message{
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
// agent, it launches requests to the endpoint `{{sfChatApi}}/chat/rest/System/Messages`, with which we will know if
// the status of the chat and we receive messages from the agents to send to the end user
func (in *Interconnection) handleLongPolling() {
	// datadog tracing
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

	logrus.WithFields(logFields).Info("Starting long polling service from salesforce")
	in.runnigLongPolling = true
	for in.runnigLongPolling {
		response, errorResponse := in.SalesforceService.GetMessages(mainSpan, in.AffinityToken, in.SessionKey)

		if errorResponse != nil {
			switch errorResponse.StatusCode {
			case http.StatusNoContent:
				logrus.WithFields(logFields).Info("Not content events")
			case http.StatusConflict:
				logrus.WithFields(logFields).Info("Duplicate Long Polling")
				time.Sleep(time.Second * 5)
			case http.StatusForbidden:
				go ChangeToState(in.UserID, in.BotSlug, TimeoutState[string(in.Provider)], in.BotrunnnerClient, BotrunnerTimeout, StudioNGTimeout, in.StudioNG, in.isStudioNGFlow)
				in.updateStatusRedis(string(Closed))
				in.Status = Closed
				in.runnigLongPolling = false
				logrus.WithFields(logFields).Error("StatusForbidden")
				mainSpan.SetTag(ext.Error, errorResponse.Error)
				mainSpan.SetTag(events.StatusSalesforce, errorResponse.StatusCode)
			case http.StatusServiceUnavailable:
				// TODO: Reconnect Session
				in.updateStatusRedis(string(Closed))
				in.Status = Closed
				in.runnigLongPolling = false
				logrus.WithFields(logFields).Error("StatusServiceUnavailable")
				mainSpan.SetTag(ext.Error, errorResponse.Error)
				mainSpan.SetTag(events.StatusSalesforce, errorResponse.StatusCode)
			default:
				logrus.WithFields(logFields).Errorf("Exists error in long polling : %s", errorResponse.Error.Error())
				go ChangeToState(in.UserID, in.BotSlug, TimeoutState[string(in.Provider)], in.BotrunnnerClient, BotrunnerTimeout, StudioNGTimeout, in.StudioNG, in.isStudioNGFlow)
				in.updateStatusRedis(string(Closed))
				in.Status = Closed
				in.runnigLongPolling = false
				mainSpan.SetTag(ext.Error, errorResponse.Error)
				mainSpan.SetTag(events.StatusSalesforce, errorResponse.StatusCode)
			}
			continue
		}

		in.offset = response.Offset
		go func(span tracer.Span) {
			for _, event := range response.Messages {
				in.checkEvent(span, &event)
			}
		}(mainSpan)

		time.Sleep(time.Second * 5)
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

	switch event.Type {
	case chat.ChatRequestFail:
		logrus.WithFields(logFields).Infof("Event [%s] : [%s]", chat.ChatRequestFail, event.Message.Reason)
		mainSpan.SetTag(ext.Error, errors.New(fmt.Sprintf("Event [%s] : [%s]", chat.ChatRequestFail, event.Message.Reason)))
		go ChangeToState(in.UserID, in.BotSlug, TimeoutState[string(in.Provider)], in.BotrunnnerClient, BotrunnerTimeout, StudioNGTimeout, in.StudioNG, in.isStudioNGFlow)
		in.updateStatusRedis(string(Failed))
		in.runnigLongPolling = false
		in.Status = Failed
	case chat.ChatRequestSuccess:
		logrus.WithFields(logFields).Infof("Event [%s]", chat.ChatRequestSuccess)
		if Messages.WaitAgent != "" {
			in.integrationsChannel <- NewIntegrationsMessage(span, in.UserID, Messages.WaitAgent, in.Provider)
		}
		//in.integrationsChannel <- NewIntegrationsMessage(in.UserID, fmt.Sprintf("%s : %v", Messages.QueuePosition, event.Message.QueuePosition), in.Provider)
		//in.integrationsChannel <- NewIntegrationsMessage(in.UserID, fmt.Sprintf("%s : %vs", Messages.WaitTime, event.Message.EstimatedWaitTime), in.Provider)
	case chat.ChatEstablished:
		logrus.WithFields(logFields).Infof("Event [%s]", event.Type)
		in.ActiveChat(span)
	case chat.ChatMessage:
		logrus.WithFields(logFields).Infof("Message from salesforce : %s", event.Message.Text)
		in.integrationsChannel <- NewIntegrationsMessage(mainSpan, in.UserID, event.Message.Text, in.Provider)
	case chat.QueueUpdate:
		logrus.WithFields(logFields).Infof("Event [%s]", chat.QueueUpdate)
		/*if event.Message.QueuePosition > 0 {
			in.integrationsChannel <- NewIntegrationsMessage(in.UserID, fmt.Sprintf("%s : %v", Messages.QueuePosition, event.Message.QueuePosition), in.Provider)
			in.integrationsChannel <- NewIntegrationsMessage(in.UserID, fmt.Sprintf("%s : %vs", Messages.WaitTime, event.Message.EstimatedWaitTime), in.Provider)
		}*/
	case chat.ChatEnded:
		go ChangeToState(in.UserID, in.BotSlug, SuccessState[string(in.Provider)], in.BotrunnnerClient, 0, 0, in.StudioNG, in.isStudioNGFlow)
		in.updateStatusRedis(string(Closed))
		in.runnigLongPolling = false
		in.Status = Closed
	default:
		logrus.WithFields(logFields).Infof("Event [%s]", event.Type)
	}
}

func (in *Interconnection) handleStatus() {
	for {
		if in.Status == Failed {
			// TODO :  Create new interconnection
			logrus.Info("Chat failed")
			in.finishChannel <- in
			return
		}

		if in.Status == Closed {
			logrus.Info("Chat Ended")
			in.finishChannel <- in
			return
		}
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
