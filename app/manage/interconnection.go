package manage

import (
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"yalochat.com/salesforce-integration/app/services"
	"yalochat.com/salesforce-integration/base/cache"
	"yalochat.com/salesforce-integration/base/clients/botrunner"
	"yalochat.com/salesforce-integration/base/clients/chat"
	"yalochat.com/salesforce-integration/base/clients/integrations"
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
	Context              string                              `json:"context"`
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
	offset        int    `json:"-"`
	lastMessageId string `json:"-"`
}

// Message represents the messages that will be sent through the chat
type Message struct {
	Text          string   `json:"text"`
	ImageUrl      string   `json:"imageUrl"`
	UserID        string   `json:"userID"`
	SessionKey    string   `json:"sessionKey"`
	AffinityToken string   `json:"affinityToken"`
	Provider      Provider `json:"provider"`
}

func NewInterconection(interconnection *Interconnection) *Interconnection {
	interconnection.Timestamp = time.Now()
	interconnection.Status = OnHold
	return interconnection
}

func NewIntegrationsMessage(userID, text string, provider Provider) *Message {
	return &Message{
		UserID:   userID,
		Text:     text,
		Provider: provider,
	}
}

func NewSfMessage(affinityToken, key, text string) *Message {
	return &Message{
		AffinityToken: affinityToken,
		SessionKey:    key,
		Text:          text,
	}
}

func (in *Interconnection) handleLongPolling() {
	logrus.Info("Starting long polling service from salesforce ")
	in.runnigLongPolling = true
	for in.runnigLongPolling {
		response, errorResponse := in.SalesforceService.GetMessages(in.AffinityToken, in.SessionKey)

		if errorResponse != nil {
			switch errorResponse.StatusCode {
			case http.StatusNoContent:
				logrus.Info("Not content events")
			case http.StatusForbidden:
				go ChangeToState(in.UserID, in.BotSlug, TimeoutState, in.BotrunnnerClient, BotrunnerTimeout)
				in.updateStatusRedis(string(Closed))
				in.Status = Closed
				in.runnigLongPolling = false
				logrus.Info("StatusForbidden")
			case http.StatusServiceUnavailable:
				// TODO: Reconnect Session
				in.updateStatusRedis(string(Closed))
				in.Status = Closed
				in.runnigLongPolling = false
				logrus.Info("StatusServiceUnavailable")
			default:
				logrus.Errorf("Exists error in long polling : %s", errorResponse.Error.Error())
				go ChangeToState(in.UserID, in.BotSlug, TimeoutState, in.BotrunnnerClient, BotrunnerTimeout)
				in.updateStatusRedis(string(Closed))
				in.Status = Closed
				in.runnigLongPolling = false
			}
			continue
		}

		in.offset = response.Offset
		for _, event := range response.Messages {
			in.checkEvent(&event)
		}
		time.Sleep(time.Second * 5)
	}
}

func (in *Interconnection) checkEvent(event *chat.MessageObject) {
	switch event.Type {
	case chat.ChatRequestFail:
		logrus.Infof("Event [%s]", chat.ChatRequestFail)
		go ChangeToState(in.UserID, in.BotSlug, TimeoutState, in.BotrunnnerClient, BotrunnerTimeout)
		in.updateStatusRedis(string(Failed))
		in.runnigLongPolling = false
		in.Status = Failed
	case chat.ChatRequestSuccess:
		logrus.Infof("Event [%s]", chat.ChatRequestSuccess)
		in.integrationsChannel <- NewIntegrationsMessage(in.UserID, "Esperando un agente", in.Provider)
		//in.integrationsChannel <- NewIntegrationsMessage(in.UserID, fmt.Sprintf("Posición en la cola: %v", event.Message.QueuePosition))
		in.integrationsChannel <- NewIntegrationsMessage(in.UserID, fmt.Sprintf("Tiempo de espera: %v seg", event.Message.EstimatedWaitTime), in.Provider)
	case chat.ChatEstablished:
		logrus.Infof("Event [%s]", event.Type)
		in.ActiveChat()
	case chat.ChatMessage:
		logrus.Infof("Message from salesforce : %s", event.Message.Text)
		in.integrationsChannel <- NewIntegrationsMessage(in.UserID, event.Message.Text, in.Provider)
	case chat.QueueUpdate:
		logrus.Infof("Event [%s]", chat.QueueUpdate)
		if event.Message.QueuePosition > 0 {
			//in.integrationsChannel <- NewIntegrationsMessage(in.UserID, fmt.Sprintf("Posición en la cola: %v", event.Message.QueuePosition))
			in.integrationsChannel <- NewIntegrationsMessage(in.UserID, fmt.Sprintf("Tiempo de espera: %v seg", event.Message.EstimatedWaitTime), in.Provider)
		}
	case chat.ChatEnded:
		go ChangeToState(in.UserID, in.BotSlug, SuccessState, in.BotrunnnerClient, 0)
		in.updateStatusRedis(string(Closed))
		in.runnigLongPolling = false
		in.Status = Closed
	default:
		logrus.Infof("Event [%s]", event.Type)
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

func (in *Interconnection) ActiveChat() {
	in.Status = Active
	in.updateStatusRedis(string(in.Status))
	in.salesforceChannel <- NewSfMessage(in.AffinityToken, in.SessionKey, in.Context)
	time.Sleep(500 * time.Millisecond)
	in.salesforceChannel <- NewSfMessage(in.AffinityToken, in.SessionKey, fmt.Sprintf("Hola soy %s y necesito ayuda", in.Name))
}

func convertInterconnectionCacheToInterconnection(interconnection cache.Interconnection) *Interconnection {
	return &Interconnection{
		UserID:        interconnection.UserID,
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
	interconnectionCache, err := in.interconnectionCache.RetrieveInterconnection(cache.Interconnection{UserID: in.UserID, SessionID: in.SessionID})

	if err != nil {
		logrus.Errorf("Could not update status in interconnection userID[%s]-sessionId[%s] from redis : [%s]", in.UserID, in.SessionID, err.Error())
	}

	interconnectionCache.Status = status
	err = in.interconnectionCache.StoreInterconnection(*interconnectionCache)
	if err != nil {
		logrus.Errorf("Could not update status in interconnection userID[%s]-sessionId[%s] from redis : [%s]", in.UserID, in.SessionID, err.Error())
	}
}
