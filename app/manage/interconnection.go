package manage

import (
	"fmt"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"yalochat.com/salesforce-integration/base/clients/chat"
	"yalochat.com/salesforce-integration/base/clients/integrations"
)

// Status that an interconnection can have
type InterconnectionStatus string
type Provider string

const (
	Failed            InterconnectionStatus = "FAILED"
	OnHold            InterconnectionStatus = "ON_HOLD"
	Active            InterconnectionStatus = "ACTIVE"
	Closed            InterconnectionStatus = "CLOSED"
	WhatsappProvider  Provider              = "whatsapp"
	MessengerProvider Provider              = "messenger"
)

// Interconnection struct represents a connection between userBotYalo and salesforce agent
type Interconnection struct {
	// The Id has the following structure {userId}-{sessionID}
	Id                  string                           `json:"id"`
	UserId              string                           `json:"userId"`
	SessionId           string                           `json:"sessionId"`
	SessionKey          string                           `json:"sessionKey"`
	AffinityToken       string                           `json:"affinityToken"`
	Status              InterconnectionStatus            `json:"status"`
	Timestamp           time.Time                        `json:"timestamp"`
	Provider            Provider                         `json:"provider"`
	BotSlug             string                           `json:"botSlug"`
	BotId               string                           `json:"botId"`
	Name                string                           `json:"name"`
	Email               string                           `json:"email"`
	PhoneNumber         string                           `json:"phoneNumber"`
	CaseId              string                           `json:"caseId"`
	ExtraData           map[string]interface{}           `json:"extraData"`
	salesforceChannel   chan *Message                    `json:"-"`
	integrationsChannel chan *Message                    `json:"-"`
	finishChannel       chan *Interconnection            `json:"-"`
	SfcChatClient       *chat.SfcChatClient              `json:"-"`
	IntegrationsClient  *integrations.IntegrationsClient `json:"-"`
	runnigLongPolling   bool                             `json:"-"`
	// This field helps us reconnect the chat in Salesforce.
	offset int `json:"-"`
}

// Message represents the messages that will be sent through the chat
type Message struct {
	Text          string `json:"text"`
	ImageUrl      string `json:"imageUrl"`
	UserId        string `json:"userId"`
	SessionKey    string `json:"sessionKey"`
	AffinityToken string `json:"affinityToken"`
}

func NewInterconection(interconnection *Interconnection) *Interconnection {
	interconnection.Id = GetKey(interconnection)
	interconnection.Timestamp = time.Now()
	interconnection.Status = OnHold
	return interconnection
}

func NewIntegrationsMessage(userId, text string) *Message {
	return &Message{
		UserId: userId,
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

func GetKey(interonection *Interconnection) string {
	return fmt.Sprintf("%s-%s", interonection.UserId, interonection.SessionId)
}

func (in *Interconnection) handleLongPolling() {
	logrus.Info("Starting long polling service from salesforce ")
	in.runnigLongPolling = true
	for in.runnigLongPolling {
		response, errorResponse := in.SfcChatClient.GetMessages(in.AffinityToken, in.SessionKey)

		if errorResponse != nil {
			switch errorResponse.StatusCode {
			case http.StatusNoContent:
				logrus.Info("Not content events")
			case http.StatusForbidden:
				// TODO: Send to state `from-sf-timeout` in the bot
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
		// TODO: Send to state `from-sf-timeout` in the bot
		logrus.Infof("Event [%s]", chat.ChatRequestFail)
		in.integrationsChannel <- NewIntegrationsMessage(in.UserId, "No hay agentes disponibles")
		in.runnigLongPolling = false
		in.Status = Failed
	case chat.ChatRequestSuccess:
		// TODO : queuePosition and send message to user
		logrus.Infof("Event [%s]", chat.ChatRequestSuccess)
		in.integrationsChannel <- NewIntegrationsMessage(in.UserId, "Esperando un agente")
		in.integrationsChannel <- NewIntegrationsMessage(in.UserId, fmt.Sprintf("Posición en la cola: %v", event.Message.QueuePosition))
		in.integrationsChannel <- NewIntegrationsMessage(in.UserId, fmt.Sprintf("Tiempo de espera: %v seg", event.Message.EstimatedWaitTime))
	case chat.ChatEstablished:
		// TODO : Activate Chat function
		in.Status = Active
		in.salesforceChannel <- NewSfMessage(in.AffinityToken, in.SessionKey, "Aqui el contexto")
		in.salesforceChannel <- NewSfMessage(in.AffinityToken, in.SessionKey, fmt.Sprintf("Hola soy %s y necesito ayuda", in.Name))
	case chat.ChatMessage:
		logrus.Infof("Message from salesforce : %s", event.Message.Text)
		//TODO: Send Message to user
		in.integrationsChannel <- NewIntegrationsMessage(in.UserId, event.Message.Text)
	case chat.QueueUpdate:
		//TODO: update queuePosition and send message to user
		logrus.Infof("Event [%s]", chat.QueueUpdate)
	case chat.ChatEnded:
		in.integrationsChannel <- NewIntegrationsMessage(in.UserId, "Terminó el chat")
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
