package manage

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/sirupsen/logrus"

	"yalochat.com/salesforce-integration/app/services"
	"yalochat.com/salesforce-integration/base/cache"
	"yalochat.com/salesforce-integration/base/clients/chat"
	"yalochat.com/salesforce-integration/base/clients/integrations"
	"yalochat.com/salesforce-integration/base/clients/login"
	"yalochat.com/salesforce-integration/base/clients/proxy"
	"yalochat.com/salesforce-integration/base/clients/salesforce"
	"yalochat.com/salesforce-integration/base/helpers"
	"yalochat.com/salesforce-integration/base/models"
)

var (
	SfcOrganizationId string
	SfcDeploymentId   string
	SfcButtonId       string
)

const (
	voiceType    = "voice"
	documentType = "document"
	imageType    = "image"
	textType     = "text"
	fromUser     = "user"
	fromBot      = "bot"
)

// Manager controls the process of the app
type Manager struct {
	clientName            string
	interconnectionMap    map[string]*Interconnection
	sessionMap            map[string]string
	sfcContactMap         map[string]*models.SfcContact
	SalesforceService     *services.SalesforceService
	IntegrationsClient    *integrations.IntegrationsClient
	salesforceChannel     chan *Message
	integrationsChannel   chan *Message
	finishInterconnection chan *Interconnection
	contextCache          cache.ContextCache
}

// ManagerOptions holds configurations for the interactions manager
type ManagerOptions struct {
	AppName             string
	RedisOptions        cache.RedisOptions
	SfcClientId         string
	SfcClientSecret     string
	SfcUsername         string
	SfcPassword         string
	SfcSecurityToken    string
	SfcBaseUrl          string
	SfcChatUrl          string
	SfcLoginUrl         string
	SfcApiVersion       string
	SfcOrganizationId   string
	SfcDeploymentId     string
	SfcButtonId         string
	SfcOwnerId          string
	IntegrationsUrl     string
	IntegrationsChannel string
	IntegrationsToken   string
	IntegrationsBotId   string
}

type ManagerI interface {
	SaveContext(integration *models.IntegrationsRequest) error
}

// CreateManager retrieves an agents manager
func CreateManager(config *ManagerOptions) *Manager {
	SfcOrganizationId = config.SfcOrganizationId
	SfcDeploymentId = config.SfcDeploymentId
	SfcButtonId = config.SfcButtonId
	contextCache, err := cache.NewRedisCache(&config.RedisOptions)

	if err != nil {
		logrus.WithError(err).Error("Error initializing Redis Manager")
	}

	sfcLoginClient := &login.SfcLoginClient{
		Proxy: proxy.NewProxy(config.SfcLoginUrl),
	}

	// Get token for salesforce
	tokenPayload := login.TokenPayload{
		ClientId:     config.SfcClientId,
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

	salesforceService := services.NewSalesforceService(*sfcLoginClient, *sfcChatClient, *salesforceClient)

	integrationsClient := integrations.NewIntegrationsClient(config.IntegrationsUrl, config.IntegrationsToken, config.IntegrationsChannel, config.IntegrationsBotId)

	m := &Manager{
		clientName:            config.AppName,
		SalesforceService:     salesforceService,
		interconnectionMap:    make(map[string]*Interconnection),
		sessionMap:            make(map[string]string),
		sfcContactMap:         make(map[string]*models.SfcContact),
		IntegrationsClient:    integrationsClient,
		salesforceChannel:     make(chan *Message),
		integrationsChannel:   make(chan *Message),
		finishInterconnection: make(chan *Interconnection),
		contextCache:          contextCache,
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
		UserId: message.UserId,
		Text:   integrations.TextMessage{Body: message.Text},
	})
	if err != nil {
		logrus.Error(helpers.ErrorMessage("Error sendMessage", err))
	}
	logrus.Infof("Send message to userId : %s", message.UserId)
}

func (m *Manager) sendMessageToSalesforce(message *Message) {
	_, err := m.SalesforceService.SfcChatClient.SendMessage(message.AffinityToken, message.SessionKey, chat.MessagePayload{Text: message.Text})
	if err != nil {
		logrus.Error(helpers.ErrorMessage("Error sendMessage", err))
	}
	logrus.Infof("Send message to agent from salesforce : %s", message.UserId)
}

// Initialize a chat with salesforce
func (m *Manager) CreateChat(interconnection *Interconnection) error {
	titleMessage := "could not create chat in salesforce"

	// Validate that user does not have an active session
	err := m.ValidateUserId(interconnection.UserId)
	if err != nil {
		return errors.New(helpers.ErrorMessage(titleMessage, err))
	}

	contact, err := m.GetOrCreateContact(interconnection.UserId, interconnection.Name, interconnection.Email, interconnection.PhoneNumber)
	if err != nil {
		return errors.New(helpers.ErrorMessage(titleMessage, err))
	}

	//TODO: Validate if status of contact in Salesforce isn't locked
	if contact.Blocked {
		return errors.New(helpers.ErrorMessage(titleMessage, errors.New("This contact is blocked")))
	}

	//Creating chat in Salesforce
	session, err := m.SalesforceService.CreatChat(interconnection.Name, SfcOrganizationId, SfcDeploymentId, SfcButtonId)
	if err != nil {
		return errors.New(helpers.ErrorMessage(titleMessage, err))
	}
	interconnection.AffinityToken = session.AffinityToken
	interconnection.SessionId = session.Id
	interconnection.SessionKey = session.Key

	//Add interconection to Redis and interconnectionMap
	m.AddInterconnection(interconnection)
	return nil
}

// We get the contact if it exists by your email or phone.
func (m *Manager) GetOrCreateContact(userId, name, email, phoneNumber string) (*models.SfcContact, error) {
	contact, ok := m.sfcContactMap[userId]
	if ok {
		return contact, nil
	}
	contact, err := m.SalesforceService.GetOrCreateContact(name, email, phoneNumber)

	if contact != nil {
		logrus.Infof("Contact added in contactMap to user %s", userId)
		m.sfcContactMap[userId] = contact
	}

	return contact, err
}

func (m *Manager) ValidateUserId(userId string) error {
	//TODO: validate that there is an active user session in redis

	// validate that it exists on the map
	if _, ok := m.sessionMap[userId]; ok {
		return errors.New("Session exists with this userID")
	}
	return nil
}

func (m *Manager) AddInterconnection(interconnection *Interconnection) {
	interconnection = NewInterconection(interconnection)
	interconnection.SfcChatClient = m.SalesforceService.SfcChatClient
	interconnection.IntegrationsClient = m.IntegrationsClient
	interconnection.finishChannel = m.finishInterconnection
	interconnection.integrationsChannel = m.integrationsChannel
	interconnection.salesforceChannel = m.salesforceChannel
	//TODO: Store interconnection in Redis
	m.interconnectionMap[interconnection.Id] = interconnection
	m.sessionMap[interconnection.UserId] = interconnection.SessionId
	go interconnection.handleLongPolling()
	go interconnection.handleStatus()
	logrus.WithFields(logrus.Fields{
		"interconnection":        interconnection.Id,
		"sessionMap":             m.sessionMap,
		"InterconectionMapCount": len(m.interconnectionMap),
	}).Info("Create interconnection successfully")
}

// SaveContext method will save context of integration message
func (m *Manager) SaveContext(integration *models.IntegrationsRequest) error {
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
	case integration.Type == imageType ||
		integration.Type == voiceType ||
		integration.Type == documentType:
		ctx.URL = integration.Image.URL
		ctx.Caption = integration.Image.Caption
		ctx.MIMEType = integration.Image.MIMEType
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

	err = m.contextCache.StoreContext(ctx)
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
	delete(m.sessionMap, interconnection.UserId)
	delete(m.interconnectionMap, interconnection.Id)
	logrus.Infof("Ending Interconnection : %s", interconnection.UserId)
}
