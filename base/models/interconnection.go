package models

import (
	"fmt"
	"time"
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

// Interconnection struct representa una conexion entre el userBotYalo y agente de salesforce
type Interconnection struct {
	// El Id tiene la siguiente estructura {userId}-{sessionID}
	Id                  string                 `json:"id"`
	UserId              string                 `json:"userId"`
	SessionId           string                 `json:"sessionId"`
	SessionKey          string                 `json:"sessionKey"`
	AffinityToken       string                 `json:"affinityToken"`
	Status              InterconnectionStatus  `json:"status"`
	Timestamp           time.Time              `json:"timestamp"`
	Provider            Provider               `json:"provider"`
	BotSlug             string                 `json:"botSlug"`
	BotId               string                 `json:"botId"`
	Name                string                 `json:"name"`
	Email               string                 `json:"email"`
	PhoneNumber         string                 `json:"phoneNumber"`
	CaseId              string                 `json:"caseId"`
	ExtraData           map[string]interface{} `json:"extraData"`
	SalesforceChannel   chan *Message          `json:"-"`
	IntegrationsChannel chan *Message          `json:"-"`
}

// Message representa los mensajes que se enviaran a través del chat
type Message struct{}

func NewInterconection(interconnection *Interconnection) *Interconnection {
	interconnection.Id = GetKey(*interconnection)
	interconnection.Timestamp = time.Now()
	interconnection.Status = OnHold
	interconnection.IntegrationsChannel = make(chan *Message)
	interconnection.SalesforceChannel = make(chan *Message)
	return interconnection
}

func GetKey(interonection Interconnection) string {
	return fmt.Sprintf("%s-%s", interonection.UserId, interonection.SessionId)
}
