package manage

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"yalochat.com/salesforce-integration/app/services"
	"yalochat.com/salesforce-integration/base/cache"
	"yalochat.com/salesforce-integration/base/clients/botrunner"
	"yalochat.com/salesforce-integration/base/clients/chat"
	"yalochat.com/salesforce-integration/base/clients/integrations"
	"yalochat.com/salesforce-integration/base/clients/login"
	"yalochat.com/salesforce-integration/base/clients/proxy"
	"yalochat.com/salesforce-integration/base/clients/salesforce"
	"yalochat.com/salesforce-integration/base/constants"
	"yalochat.com/salesforce-integration/base/helpers"
	"yalochat.com/salesforce-integration/base/models"
)

var (
	SfcOrganizationID   string
	SfcDeploymentID     string
	SfcButtonID         string
	SfcRecordTypeID     string
	BlockedUserState    string
	TimeoutState        string
	SuccessState        string
	SfcCustomFieldsCase []string
)

const (
	voiceType          = "voice"
	documentType       = "document"
	imageType          = "image"
	textType           = "text"
	fromUser           = "user"
	fromBot            = "bot"
	descriptionDefualt = "Caso levantado por el Bot : "
)

type (
	interconnectionMap   map[string]*Interconnection
	interconnectionCache struct {
		interconnections interconnectionMap
		sync.RWMutex
	}
)

// Manager controls the process of the app
type Manager struct {
	clientName            string
	interconnectionMap    interconnectionCache
	SalesforceService     services.SalesforceServiceInterface
	IntegrationsClient    *integrations.IntegrationsClient
	BotrunnnerClient      botrunner.BotRunnerInterface
	salesforceChannel     chan *Message
	integrationsChannel   chan *Message
	finishInterconnection chan *Interconnection
	cache                 cache.ContextCache
}

// ManagerOptions holds configurations for the interactions manager
type ManagerOptions struct {
	AppName               string
	BlockedUserState      string
	TimeoutState          string
	SuccessState          string
	RedisOptions          cache.RedisOptions
	BotrunnerUrl          string
	BotrunnerToken        string
	SfcClientID           string
	SfcClientSecret       string
	SfcUsername           string
	SfcPassword           string
	SfcSecurityToken      string
	SfcBaseUrl            string
	SfcChatUrl            string
	SfcLoginUrl           string
	SfcApiVersion         string
	SfcOrganizationID     string
	SfcDeploymentID       string
	SfcButtonID           string
	SfcOwnerId            string
	SfcRecordTypeID       string
	SfcCustomFieldsCase   []string
	IntegrationsUrl       string
	IntegrationsChannel   string
	IntegrationsToken     string
	IntegrationsBotID     string
	IntegrationsSignature string
	IntegrationsBotPhone  string
	WebhookBaseUrl        string
}

type ManagerI interface {
	SaveContext(integration *models.IntegrationsRequest) error
	CreateChat(interconnection *Interconnection) error
	GetContextByUserID(userID string) string
}

// CreateManager retrieves an agents manager
func CreateManager(config *ManagerOptions) *Manager {
	SfcOrganizationID = config.SfcOrganizationID
	SfcDeploymentID = config.SfcDeploymentID
	SfcButtonID = config.SfcButtonID
	SfcRecordTypeID = config.SfcRecordTypeID
	BlockedUserState = config.BlockedUserState
	TimeoutState = config.TimeoutState
	SuccessState = config.SuccessState
	SfcCustomFieldsCase = config.SfcCustomFieldsCase
	cache, err := cache.NewRedisCache(&config.RedisOptions)

	if err != nil {
		logrus.WithError(err).Error("Error initializing Redis Manager")
	}

	sfcLoginClient := &login.SfcLoginClient{
		Proxy: proxy.NewProxy(config.SfcLoginUrl),
	}

	// Get token for salesforce
	tokenPayload := login.TokenPayload{
		ClientId:     config.SfcClientID,
		ClientSecret: config.SfcClientSecret,
		Username:     config.SfcUsername,
		Password:     config.SfcPassword + config.SfcSecurityToken,
	}

	token, err := sfcLoginClient.GetToken(tokenPayload)
	if err != nil {
		logrus.Errorf("Could not get access token from salesforce Server : %s", err.Error())
	}

	sfcChatClient := &chat.SfcChatClient{
		Proxy:       proxy.NewProxy(config.SfcChatUrl),
		ApiVersion:  config.SfcApiVersion,
		AccessToken: token,
	}

	salesforceClient := &salesforce.SalesforceClient{
		Proxy:       proxy.NewProxy(config.SfcBaseUrl),
		APIVersion:  config.SfcApiVersion,
		AccessToken: token,
	}

	integrationsClient := integrations.NewIntegrationsClient(
		config.IntegrationsUrl,
		config.IntegrationsToken,
		config.IntegrationsChannel,
		config.IntegrationsBotID,
	)

	_, err = integrationsClient.WebhookRegister(integrations.HealthcheckPayload{
		Phone:   config.IntegrationsBotPhone,
		Webhook: fmt.Sprintf("%s/v1/integrations/webhook", config.WebhookBaseUrl),
	})
	if err != nil {
		logrus.Errorf("could not set webhook on integrations : %s", err.Error())
	}

	salesforceService := services.NewSalesforceService(*sfcLoginClient, *sfcChatClient, *salesforceClient)
	botRunnerClient := botrunner.NewBotrunnerClient(config.BotrunnerUrl, config.BotrunnerToken)

	interconnections := cache.RetrieveAllInterconnections()

	interconnectionsMap := make(interconnectionMap)
	for _, interconnection := range *interconnections {
		if InterconnectionStatus(interconnection.Status) == Active || interconnection.Status == string(OnHold) {
			interconnectionID := fmt.Sprintf("%s-%s", interconnection.UserID, interconnection.SessionID)
			interconnectionsMap[interconnectionID] = &Interconnection{
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

	}

	m := &Manager{
		clientName:            config.AppName,
		SalesforceService:     salesforceService,
		interconnectionMap:    interconnectionCache{interconnections: make(interconnectionMap)},
		IntegrationsClient:    integrationsClient,
		salesforceChannel:     make(chan *Message),
		integrationsChannel:   make(chan *Message),
		finishInterconnection: make(chan *Interconnection),
		cache:                 cache,
		BotrunnnerClient:      botRunnerClient,
	}
	go m.handleInterconnection()
	return m
}

// Esta funcion finaliza las interconecciones y envia los mensajes a salesforce o yalo
func (m *Manager) handleInterconnection() {
	for {
		select {
		case interconection := <-m.finishInterconnection:
			m.EndChat(interconection)
		case messageSf := <-m.salesforceChannel:
			m.sendMessageToSalesforce(messageSf)
		case messageInt := <-m.integrationsChannel:
			m.sendMessageToUser(messageInt)
		}
	}
}

func (m *Manager) sendMessageToUser(message *Message) {
	_, err := m.IntegrationsClient.SendMessage(integrations.SendTextPayload{
		Id:     helpers.RandomString(36),
		Type:   "text",
		UserId: message.UserID,
		Text:   integrations.TextMessage{Body: message.Text},
	})
	if err != nil {
		logrus.Error(helpers.ErrorMessage("Error sendMessage", err))
	}
	logrus.Infof("Send message to UserID : %s", message.UserID)
}

func (m *Manager) sendMessageToSalesforce(message *Message) {
	_, err := m.SalesforceService.SendMessage(message.AffinityToken, message.SessionKey, chat.MessagePayload{Text: message.Text})
	if err != nil {
		logrus.Error(helpers.ErrorMessage("Error sendMessage", err))
	}
	logrus.Infof("Send message to agent from salesforce : %s", message.UserID)
}

// Initialize a chat with salesforce
func (m *Manager) CreateChat(interconnection *Interconnection) error {
	titleMessage := "could not create chat in salesforce"

	// Validate that user does not have an active session
	err := m.ValidateUserID(interconnection.UserID)
	if err != nil {
		return errors.New(helpers.ErrorMessage(titleMessage, err))
	}

	// We get the contact if it exists by your email or phone.
	contact, err := m.SalesforceService.GetOrCreateContact(interconnection.Name, interconnection.Email, interconnection.PhoneNumber)
	if err != nil {
		return errors.New(helpers.ErrorMessage(titleMessage, err))
	}

	if contact.Blocked {
		return m.SentToBlockedState(interconnection.UserID, interconnection.BotSlug)
	}

	caseId, err := m.SalesforceService.CreatCase(SfcRecordTypeID, contact.Id, descriptionDefualt, string(interconnection.Provider),
		interconnection.ExtraData, SfcCustomFieldsCase)
	if err != nil {
		return errors.New(helpers.ErrorMessage(titleMessage, err))
	}
	interconnection.CaseID = caseId

	//Creating chat in Salesforce
	session, err := m.SalesforceService.CreatChat(interconnection.Name, SfcOrganizationID, SfcDeploymentID, SfcButtonID, caseId, contact.Id)
	if err != nil {
		return errors.New(helpers.ErrorMessage(titleMessage, err))
	}
	interconnection.AffinityToken = session.AffinityToken
	interconnection.SessionID = session.Id
	interconnection.SessionKey = session.Key
	interconnection.Context = "Contexto:\n" + m.GetContextByUserID(interconnection.UserID)

	//Add interconection to Redis and interconnectionMap
	m.AddInterconnection(interconnection)
	return nil
}

// Change to from-sf-blocked state
func (m *Manager) SentToBlockedState(userID, botSlug string) error {
	ok, err := m.BotrunnnerClient.SendTo(botrunner.GetRequestToSendTo(botSlug, userID, BlockedUserState, ""))

	if ok {
		return errors.New(helpers.ErrorMessage("could not create chat in salesforce", err))
	}
	return err
}

func (m *Manager) ValidateUserID(userID string) error {
	//TODO: validate that there is an active user session in redis

	// validate that it exists on the map
	if _, ok := m.interconnectionMap.interconnections[userID]; ok {
		return errors.New("session exists with this userID")
	}
	return nil
}

func (m *Manager) AddInterconnection(interconnection *Interconnection) {
	interconnection = NewInterconection(interconnection)
	interconnection.SalesforceService = m.SalesforceService
	interconnection.IntegrationsClient = m.IntegrationsClient
	interconnection.BotrunnnerClient = m.BotrunnnerClient
	interconnection.finishChannel = m.finishInterconnection
	interconnection.integrationsChannel = m.integrationsChannel
	interconnection.salesforceChannel = m.salesforceChannel
	//TODO: Store interconnection in Redis

	m.interconnectionMap.Lock()
	defer m.interconnectionMap.Unlock()
	m.interconnectionMap.interconnections[interconnection.UserID] = interconnection

	go interconnection.handleLongPolling()
	go interconnection.handleStatus()
	logrus.WithFields(logrus.Fields{
		"InterconectionMapCount": len(m.interconnectionMap.interconnections),
	}).Info("Create interconnection successfully")
}

// SaveContext method will save context of integration message
func (m *Manager) SaveContext(integration *models.IntegrationsRequest) error {
	m.interconnectionMap.RWMutex.RLock()
	defer m.interconnectionMap.RUnlock()

	if interconnection, ok := m.interconnectionMap.interconnections[integration.From]; ok {
		//TODO: add image
		if interconnection.Status == Active && integration.Type == textType {
			m.sendMessageToSalesforce(&Message{
				Text:          integration.Text.Body,
				UserID:        integration.From,
				SessionKey:    interconnection.SessionKey,
				AffinityToken: interconnection.AffinityToken,
			})
		}

		return nil
	}

	timestamp, err := strconv.Atoi(integration.Timestamp)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"integration": integration,
		}).WithError(err).Error("Error format timestamp")
		return err
	}

	ctx := cache.Context{
		UserID:    integration.From,
		Timestamp: timestamp,
		From:      fromUser,
	}

	switch {
	case integration.Type == imageType:
		ctx.URL = integration.Image.URL
		ctx.Caption = integration.Image.Caption
		ctx.MIMEType = integration.Image.MIMEType
	case integration.Type == voiceType:
		ctx.URL = integration.Voice.URL
		ctx.Caption = integration.Voice.Caption
		ctx.MIMEType = integration.Voice.MIMEType
	case integration.Type == documentType:
		ctx.URL = integration.Document.URL
		ctx.Caption = integration.Document.Caption
		ctx.MIMEType = integration.Document.MIMEType
	case integration.Type == textType:
		ctx.Text = integration.Text.Body
	default:
		logrus.WithFields(logrus.Fields{
			"integration": integration,
		}).WithError(err).Error("invalid type message")
		return fmt.Errorf("invalid type message")
	}

	if integration.To != "" {
		ctx.From = fromBot
		ctx.UserID = integration.To
	}

	err = m.cache.StoreContext(ctx)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"context": ctx,
		}).WithError(err).Error("Error store context")
		return err
	}

	return nil
}

func (m *Manager) EndChat(interconnection *Interconnection) {
	// TODO: EndChat
	m.interconnectionMap.Lock()
	defer m.interconnectionMap.Unlock()
	delete(m.interconnectionMap.interconnections, interconnection.UserID)
	logrus.Infof("Ending Interconnection : %s", interconnection.UserID)
}

func (m *Manager) GetContextByUserID(userID string) string {
	allContext := m.cache.RetrieveContext(userID)

	sort.Slice(allContext, func(i, j int) bool { return allContext[j].Timestamp > allContext[i].Timestamp })
	builder := strings.Builder{}
	for _, ctx := range allContext {
		loc, _ := time.LoadLocation("America/Mexico_City")
		date := time.Unix(0, int64(ctx.Timestamp)*int64(time.Millisecond)).
			In(loc).
			Format(constants.DateFormat)

		if ctx.From == fromUser {
			fmt.Fprintf(&builder, "Cliente [%s]:%s\n", date, ctx.Text)
		} else {
			fmt.Fprintf(&builder, "Bot [%s]:%s\n", date, ctx.Text)
		}
	}

	return builder.String()
}
