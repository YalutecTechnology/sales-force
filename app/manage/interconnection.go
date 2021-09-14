package manage

import (
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"yalochat.com/salesforce-integration/app/services"
	"yalochat.com/salesforce-integration/base/clients/botrunner"
	"yalochat.com/salesforce-integration/base/clients/chat"
	"yalochat.com/salesforce-integration/base/clients/integrations"
	"yalochat.com/salesforce-integration/base/helpers"
)

// Status that an interconnection can have
type InterconnectionStatus string
type Provider string

const (
	Failed            InterconnectionStatus = "FAILED"
	OnHold            InterconnectionStatus = "ON_HOLD"
	Active            InterconnectionStatus = "ACTIVE"
	Closed            InterconnectionStatus = "CLOSED"
	WhatsappProvider  Provider              = "Whatsapp"
	MessengerProvider Provider              = "Facebook"
)

// Interconnection struct represents a connection between userBotYalo and salesforce agent
type Interconnection struct {
	UserID              string                              `json:"userId"`
	SessionID           string                              `json:"sessionId"`
	SessionKey          string                              `json:"sessionKey"`
	AffinityToken       string                              `json:"affinityToken"`
	Status              InterconnectionStatus               `json:"status"`
	Timestamp           time.Time                           `json:"timestamp"`
	Provider            Provider                            `json:"provider"`
	BotSlug             string                              `json:"botSlug"`
	BotID               string                              `json:"botId"`
	Name                string                              `json:"name"`
	Email               string                              `json:"email"`
	PhoneNumber         string                              `json:"phoneNumber"`
	CaseID              string                              `json:"caseId"`
	Context             string                              `json:"context"`
	ExtraData           map[string]interface{}              `json:"extraData"`
	salesforceChannel   chan *Message                       `json:"-"`
	integrationsChannel chan *Message                       `json:"-"`
	finishChannel       chan *Interconnection               `json:"-"`
	BotrunnnerClient    botrunner.BotRunnerInterface        `json:"-"`
	SalesforceService   services.SalesforceServiceInterface `json:"-"`
	IntegrationsClient  *integrations.IntegrationsClient    `json:"-"`
	runnigLongPolling   bool                                `json:"-"`
	// This field helps us reconnect the chat in Salesforce.
	offset int `json:"-"`
}

// Message represents the messages that will be sent through the chat
type Message struct {
	Text          string `json:"text"`
	ImageUrl      string `json:"imageUrl"`
	UserID        string `json:"userID"`
	SessionKey    string `json:"sessionKey"`
	AffinityToken string `json:"affinityToken"`
}

func NewInterconection(interconnection *Interconnection) *Interconnection {
	interconnection.Timestamp = time.Now()
	interconnection.Status = OnHold
	return interconnection
}

func NewIntegrationsMessage(userID, text string) *Message {
	return &Message{
		UserID: userID,
		Text:   text,
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
				_, err := in.BotrunnnerClient.SendTo(botrunner.GetRequestToSendTo(in.BotSlug, in.UserID, TimeoutState, ""))

				if err != nil {
					logrus.Infof(helpers.ErrorMessage("could not sent to state timeout", err))
				}

				in.Status = Closed
				in.runnigLongPolling = false
				logrus.Info("StatusForbidden")
			case http.StatusServiceUnavailable:
				// TODO: Reconnect Session
				in.Status = Closed
				in.runnigLongPolling = false
				logrus.Info("StatusServiceUnavailable")
			default:
				logrus.Errorf("Exists error in long polling : %s", errorResponse.Error.Error())
				_, err := in.BotrunnnerClient.SendTo(botrunner.GetRequestToSendTo(in.BotSlug, in.UserID, TimeoutState, ""))

				if err != nil {
					logrus.Infof(helpers.ErrorMessage("could not sent to state timeout", err))
				}
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
		_, err := in.BotrunnnerClient.SendTo(botrunner.GetRequestToSendTo(in.BotSlug, in.UserID, TimeoutState, ""))

		if err != nil {
			logrus.Errorf(helpers.ErrorMessage("could not sent to state timeout", err))
		}
		in.runnigLongPolling = false
		in.Status = Failed
	case chat.ChatRequestSuccess:
		logrus.Infof("Event [%s]", chat.ChatRequestSuccess)
		in.integrationsChannel <- NewIntegrationsMessage(in.UserID, "Esperando un agente")
		//in.integrationsChannel <- NewIntegrationsMessage(in.UserID, fmt.Sprintf("Posición en la cola: %v", event.Message.QueuePosition))
		in.integrationsChannel <- NewIntegrationsMessage(in.UserID, fmt.Sprintf("Tiempo de espera: %v seg", event.Message.EstimatedWaitTime))
	case chat.ChatEstablished:
		logrus.Infof("Event [%s]", event.Type)
		in.ActiveChat()
	case chat.ChatMessage:
		logrus.Infof("Message from salesforce : %s", event.Message.Text)
		in.integrationsChannel <- NewIntegrationsMessage(in.UserID, event.Message.Text)
	case chat.QueueUpdate:
		logrus.Infof("Event [%s]", chat.QueueUpdate)
		if event.Message.QueuePosition > 0 {
			//in.integrationsChannel <- NewIntegrationsMessage(in.UserID, fmt.Sprintf("Posición en la cola: %v", event.Message.QueuePosition))
			in.integrationsChannel <- NewIntegrationsMessage(in.UserID, fmt.Sprintf("Tiempo de espera: %v seg", event.Message.EstimatedWaitTime))
		}
	case chat.ChatEnded:
		_, err := in.BotrunnnerClient.SendTo(botrunner.GetRequestToSendTo(in.BotSlug, in.UserID, SuccessState, ""))

		if err != nil {
			logrus.Infof(helpers.ErrorMessage("could not sent to state timeout", err))
		}
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
	//TODO: Update interconnection in redis
	in.salesforceChannel <- NewSfMessage(in.AffinityToken, in.SessionKey, in.Context)
	in.salesforceChannel <- NewSfMessage(in.AffinityToken, in.SessionKey, fmt.Sprintf("Hola soy %s y necesito ayuda", in.Name))
}
