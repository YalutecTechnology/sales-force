package manage

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"yalochat.com/salesforce-integration/app/config/envs"
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
	SfcRecordTypeID     string
	BlockedUserState    string
	TimeoutState        string
	SuccessState        string
	SfcCustomFieldsCase map[string]string
	BotrunnerTimeout    int
)

const (
	voiceType           = "voice"
	documentType        = "document"
	imageType           = "image"
	textType            = "text"
	fromUser            = "user"
	fromBot             = "bot"
	descriptionDefualt  = "Caso levantado por el Bot : "
	messageError        = "Imagen no enviada"
	messageImageSuccess = "**El usuario adjunto una imagen al caso**"
	defaultFieldCustom  = "default"
)

// Manager controls the process of the app
type Manager struct {
	clientName            string
	interconnectionMap    cache.ICache
	SalesforceService     services.SalesforceServiceInterface
	IntegrationsClient    integrations.IntegrationInterface
	BotrunnnerClient      botrunner.BotRunnerInterface
	salesforceChannel     chan *Message
	integrationsChannel   chan *Message
	finishInterconnection chan *Interconnection
	contextcache          cache.ContextCache
	interconnectionsCache cache.InterconnectionCache
	environment           string
	keywordsRestart       []string
	cacheMessage          cache.IMessageCache
	SfcSourceFlowBot      envs.SfcSourceFlowBot
	SfcSourceFlowField    string
}

// ManagerOptions holds configurations for the interactions manager
type ManagerOptions struct {
	AppName                string
	BlockedUserState       string
	TimeoutState           string
	SuccessState           string
	RedisOptions           cache.RedisOptions
	BotrunnerUrl           string
	BotrunnerToken         string
	BotrunnerTimeout       int
	SfcClientID            string
	SfcClientSecret        string
	SfcUsername            string
	SfcPassword            string
	SfcSecurityToken       string
	SfcBaseUrl             string
	SfcChatUrl             string
	SfcLoginUrl            string
	SfcApiVersion          string
	SfcOrganizationID      string
	SfcDeploymentID        string
	SfcRecordTypeID        string
	SfcAccountRecordTypeID string
	SfcCustomFieldsCase    map[string]string
	IntegrationsUrl        string
	IntegrationsWAChannel  string
	IntegrationsFBChannel  string
	IntegrationsWAToken    string
	IntegrationsFBToken    string
	IntegrationsWABotID    string
	IntegrationsFBBotID    string
	IntegrationsSignature  string
	IntegrationsWABotPhone string
	IntegrationsFBBotPhone string
	WebhookBaseUrl         string
	Environment            string
	KeywordsRestart        []string
	SfcSourceFlowBot       envs.SfcSourceFlowBot
	SfcSourceFlowField     string
	SfcBlockedChatField    bool
}

type ManagerI interface {
	SaveContext(integration *models.IntegrationsRequest) error
	CreateChat(interconnection *Interconnection) error
	GetContextByUserID(userID string) []cache.Context
	SaveContextFB(integration *models.IntegrationsFacebook) error
	FinishChat(userId string) error
}

// CreateManager retrieves an agents manager
func CreateManager(config *ManagerOptions) *Manager {
	SfcOrganizationID = config.SfcOrganizationID
	SfcDeploymentID = config.SfcDeploymentID
	SfcRecordTypeID = config.SfcRecordTypeID
	BlockedUserState = config.BlockedUserState
	TimeoutState = config.TimeoutState
	SuccessState = config.SuccessState
	SfcCustomFieldsCase = config.SfcCustomFieldsCase
	BotrunnerTimeout = config.BotrunnerTimeout

	contextCache, err := cache.NewRedisCache(&config.RedisOptions)

	if err != nil {
		logrus.WithError(err).Error("Error initializing Context Redis Manager")
	}

	interconnectionsCache, err := cache.NewRedisCache(&config.RedisOptions)

	if err != nil {
		logrus.WithError(err).Error("Error initializing Interconnection Redis Manager")
	}

	sfcLoginClient := &login.SfcLoginClient{
		Proxy: proxy.NewProxy(config.SfcLoginUrl, 30),
	}

	// Get token for salesforce
	tokenPayload := login.TokenPayload{
		ClientId:     config.SfcClientID,
		ClientSecret: config.SfcClientSecret,
		Username:     config.SfcUsername,
		Password:     config.SfcPassword + config.SfcSecurityToken,
	}

	sfcChatClient := &chat.SfcChatClient{
		Proxy:      proxy.NewProxy(config.SfcChatUrl, 60),
		ApiVersion: config.SfcApiVersion,
	}

	salesforceClient := &salesforce.SalesforceClient{
		Proxy:               proxy.NewProxy(config.SfcBaseUrl, 30),
		APIVersion:          config.SfcApiVersion,
		SfcBlockedChatField: config.SfcBlockedChatField,
	}

	integrationsClient := integrations.NewIntegrationsClient(
		config.IntegrationsUrl,
		config.IntegrationsWAToken,
		config.IntegrationsFBToken,
		config.IntegrationsWAChannel,
		config.IntegrationsFBChannel,
		config.IntegrationsWABotID,
		config.IntegrationsFBBotID,
	)

	_, err = integrationsClient.WebhookRegister(integrations.HealthcheckPayload{
		Phone:    config.IntegrationsWABotPhone,
		Webhook:  fmt.Sprintf("%s/v1/integrations/whatsapp/webhook", config.WebhookBaseUrl),
		Provider: string(WhatsappProvider),
	})
	if err != nil {
		logrus.Errorf("could not set whatsapp webhook on integrations : %s", err.Error())
	}

	_, err = integrationsClient.WebhookRegister(integrations.HealthcheckPayload{
		Phone:    config.IntegrationsFBBotPhone,
		Webhook:  fmt.Sprintf("%s/v1/integrations/facebook/webhook", config.WebhookBaseUrl),
		Provider: string(FacebookProvider),
	})
	if err != nil {
		logrus.Errorf("could not set facebook webhook on integrations : %s", err.Error())
	}

	salesforceService := services.NewSalesforceService(*sfcLoginClient,
		*sfcChatClient,
		*salesforceClient,
		tokenPayload,
		config.SfcCustomFieldsCase,
		SfcRecordTypeID)

	salesforceService.AccountRecordTypeId = config.SfcAccountRecordTypeID
	botRunnerClient := botrunner.NewBotrunnerClient(config.BotrunnerUrl, config.BotrunnerToken)

	interconnections := interconnectionsCache.RetrieveAllInterconnections()

	cacheLocal := cache.New()
	m := &Manager{
		clientName:            config.AppName,
		SalesforceService:     salesforceService,
		interconnectionMap:    cacheLocal,
		IntegrationsClient:    integrationsClient,
		salesforceChannel:     make(chan *Message),
		integrationsChannel:   make(chan *Message),
		finishInterconnection: make(chan *Interconnection),
		contextcache:          contextCache,
		interconnectionsCache: interconnectionsCache,
		BotrunnnerClient:      botRunnerClient,
		environment:           config.Environment,
		keywordsRestart:       config.KeywordsRestart,
		SfcSourceFlowBot:      config.SfcSourceFlowBot,
		SfcSourceFlowField:    config.SfcSourceFlowField,
		cacheMessage:          cache.NewMessageCache(cacheLocal),
	}

	for _, interconnection := range *interconnections {
		if InterconnectionStatus(interconnection.Status) == Active || interconnection.Status == string(OnHold) {
			in := convertInterconnectionCacheToInterconnection(interconnection)
			m.AddInterconnection(in)
		}

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
	switch message.Provider {
	case WhatsappProvider:
		_, err := m.IntegrationsClient.SendMessage(integrations.SendTextPayload{
			Id:     helpers.RandomString(36),
			Type:   "text",
			UserID: message.UserID,
			Text:   integrations.TextMessage{Body: message.Text},
		}, string(message.Provider))
		if err != nil {
			logrus.Error(helpers.ErrorMessage("Error sendMessage", err))
		}
	case FacebookProvider:
		_, err := m.IntegrationsClient.SendMessage(integrations.SendTextPayloadFB{
			MessagingType: "RESPONSE",
			Recipient: integrations.Recipient{
				ID: message.UserID,
			},
			Message: integrations.Message{
				Text: message.Text,
			},
			Metadata: "YALOSOURCE:FIREHOSE",
		}, string(message.Provider))
		if err != nil {
			logrus.Error(helpers.ErrorMessage("Error sendMessage", err))
		}
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
		logrus.WithFields(logrus.Fields{
			"interconnection": interconnection,
		}).WithError(err).Error("error ValidateUserID")
		return errors.New(helpers.ErrorMessage(titleMessage, err))
	}

	// We get the contact if it exists by your email or phone.
	contact, err := m.SalesforceService.GetOrCreateContact(interconnection.Name, interconnection.Email, interconnection.PhoneNumber)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"interconnection": interconnection,
		}).WithError(err).Error("error GetOrCreateContact")
		return errors.New(helpers.ErrorMessage(titleMessage, err))
	}

	if contact.Blocked {
		go ChangeToState(interconnection.UserID, interconnection.BotSlug, BlockedUserState, m.BotrunnnerClient, BotrunnerTimeout)
		return fmt.Errorf("%s: %s", "could not create chat in salesforce", "this contact is blocked")
	}

	buttonID, ownerID, subject := m.changeButtonIDAndOwnerID(interconnection.Provider, interconnection.ExtraData)

	caseId, err := m.SalesforceService.CreatCase(contact.ID, descriptionDefualt, subject, string(interconnection.Provider), ownerID,
		interconnection.ExtraData)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"interconnection": interconnection,
		}).WithError(err).Error("error CreatCase")
		return errors.New(helpers.ErrorMessage(titleMessage, err))
	}
	interconnection.CaseID = caseId

	//Creating chat in Salesforce
	session, err := m.SalesforceService.CreatChat(interconnection.Name, SfcOrganizationID, SfcDeploymentID, buttonID, caseId, contact.ID)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"interconnection": interconnection,
		}).WithError(err).Error("error CreatChat")
		return errors.New(helpers.ErrorMessage(titleMessage, err))
	}
	interconnection.AffinityToken = session.AffinityToken
	interconnection.SessionID = session.Id
	interconnection.SessionKey = session.Key
	interconnection.Context = "Contexto:\n" + m.getContextByUserID(interconnection.UserID)
	interconnection.Status = OnHold
	interconnection.Timestamp = time.Now()
	time.Sleep(time.Second * 1)

	//Add interconection to Redis and interconnectionMap
	m.AddInterconnection(interconnection)
	return nil
}

// Change to state with botrunner
func ChangeToState(userID, botSlug, state string, botRunnerClient botrunner.BotRunnerInterface, seconds int) {
	time.Sleep(time.Second * time.Duration(seconds))
	_, err := botRunnerClient.SendTo(botrunner.GetRequestToSendTo(botSlug, userID, state, ""))
	if err != nil {
		logrus.Errorf(helpers.ErrorMessage("could not sent to state timeout", err))
	}
}

func (m *Manager) FinishChat(userId string) error {
	titleMessage := "could not finish chat in salesforce"

	// Get session interconnection
	interconnection, exist := m.interconnectionMap.Get(fmt.Sprintf(constants.UserKey, userId))
	if !exist {
		return errors.New(helpers.ErrorMessage(titleMessage, errors.New("this contact does not have an interconnection")))
	}

	in := interconnection.(*Interconnection)

	// End chat Salesforce
	err := m.SalesforceService.EndChat(in.AffinityToken, in.SessionKey)
	if err != nil {
		return errors.New(helpers.ErrorMessage(titleMessage, err))
	}

	m.EndChat(in)
	return nil
}

func (m *Manager) ValidateUserID(userID string) error {
	sessionRedis := m.interconnectionsCache.RetrieveInterconnectionActiveByUserId(userID)

	if sessionRedis != nil {
		return errors.New("session exists in redis with this userID")
	}

	// validate that it exists on the map
	if _, ok := m.interconnectionMap.Get(fmt.Sprintf(constants.UserKey, userID)); ok {
		return errors.New("session exists with this userID")
	}
	return nil
}

func (m *Manager) AddInterconnection(interconnection *Interconnection) {
	interconnection.SalesforceService = m.SalesforceService
	interconnection.IntegrationsClient = m.IntegrationsClient
	interconnection.BotrunnnerClient = m.BotrunnnerClient
	interconnection.finishChannel = m.finishInterconnection
	interconnection.integrationsChannel = m.integrationsChannel
	interconnection.salesforceChannel = m.salesforceChannel
	interconnection.interconnectionCache = m.interconnectionsCache

	err := m.interconnectionsCache.StoreInterconnection(NewInterconectionCache(interconnection))
	if err != nil {
		logrus.Errorf("Could not store interconnection with userID[%s] in redis[%s]", interconnection.UserID, interconnection.SessionID)
	}

	m.interconnectionMap.Set(fmt.Sprintf(constants.UserKey, interconnection.UserID), interconnection, 0)

	go interconnection.handleLongPolling()
	go interconnection.handleStatus()
	logrus.Info("Create interconnection successfully")
}

// SaveContext method will save context of integration message
func (m *Manager) SaveContext(integration *models.IntegrationsRequest) error {
	logrus.Info("WEBHOOK WHATSAPP: ", integration)
	if m.cacheMessage.IsRepeatedMessage(integration.ID) {
		return nil
	}

	isSend, err := m.salesforceComunication(integration)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"integration": integration,
		}).WithError(err).Error("Error salesforce comunication")
		return err
	}
	if isSend {
		return nil
	}

	timestamp, err := strconv.ParseInt(integration.Timestamp, 10, 64)
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
	case integration.Type == textType:
		ctx.Text = integration.Text.Body
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

	err = m.contextcache.StoreContext(ctx)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"context": ctx,
		}).WithError(err).Error("Error store context")
		return err
	}

	return nil
}

func (m *Manager) salesforceComunication(integration *models.IntegrationsRequest) (bool, error) {
	if interconnection, ok := m.validInterconnection(integration.From); ok && interconnection.Status == Active {
		switch integration.Type {
		case textType:
			if strings.Contains(constants.DevEnvironments, m.environment) {
				for _, keyword := range m.keywordsRestart {
					if strings.ToLower(integration.Text.Body) == keyword {
						err := m.SalesforceService.EndChat(interconnection.AffinityToken, interconnection.SessionKey)
						if err != nil {
							return false, err
						}
						interconnection.updateStatusRedis(string(Closed))
						interconnection.Status = Closed
						interconnection.runnigLongPolling = false
						return true, nil
					}
				}
			}

			interconnection.salesforceChannel <- NewSfMessage(interconnection.AffinityToken, interconnection.SessionKey, integration.Text.Body)

		case imageType:
			err := m.SalesforceService.InsertImageInCase(
				integration.Image.URL,
				interconnection.SessionID,
				integration.Image.MIMEType,
				interconnection.CaseID)
			if err != nil {
				interconnection.integrationsChannel <- NewIntegrationsMessage(integration.From, messageError, WhatsappProvider)
				return false, err
			}
			interconnection.salesforceChannel <- NewSfMessage(interconnection.AffinityToken, interconnection.SessionKey, messageImageSuccess)
		}

		return true, nil
	}
	return false, nil
}

func (m *Manager) validInterconnection(from string) (*Interconnection, bool) {
	if interconnection, ok := m.interconnectionMap.Get(fmt.Sprintf(constants.UserKey, from)); ok {
		if in, ok := interconnection.(*Interconnection); ok {
			return in, true
		}
	}
	return nil, false
}

func (m *Manager) EndChat(interconnection *Interconnection) {
	m.interconnectionMap.Delete(fmt.Sprintf(constants.UserKey, interconnection.UserID))
	logrus.Infof("Ending Interconnection : %s", interconnection.UserID)
}

func (m *Manager) getContextByUserID(userID string) string {
	allContext := m.contextcache.RetrieveContext(userID)

	sort.Slice(allContext, func(i, j int) bool { return allContext[j].Timestamp > allContext[i].Timestamp })
	builder := strings.Builder{}
	for _, ctx := range allContext {
		loc, _ := time.LoadLocation("America/Mexico_City")

		date := time.Unix(0, int64(ctx.Timestamp)*int64(time.Millisecond)).
			In(loc).
			Format(constants.DateFormat)

		ctx.Text = strings.TrimRight(ctx.Text, "\n")
		if ctx.From == fromUser {
			fmt.Fprintf(&builder, "Cliente [%s]:%s\n\n", date, ctx.Text)
		} else {
			fmt.Fprintf(&builder, "Bot [%s]:%s\n\n", date, ctx.Text)
		}
	}

	return builder.String()
}

func (m *Manager) GetContextByUserID(userID string) []cache.Context {
	allContext := m.contextcache.RetrieveContext(userID)

	sort.Slice(allContext, func(i, j int) bool { return allContext[j].Timestamp > allContext[i].Timestamp })

	return allContext
}

// Change button or owner according to the provider or by custom fields
func (m *Manager) changeButtonIDAndOwnerID(provider Provider, extraData map[string]interface{}) (buttonID, ownerID, subject string) {
	var sourceFlow envs.SourceFlowBot
	if SourceFlowBotOption, ok := extraData[m.SfcSourceFlowField]; ok {
		if sourceFlow, ok = m.SfcSourceFlowBot[SourceFlowBotOption.(string)]; !ok {
			if sourceFlow, ok = m.SfcSourceFlowBot[defaultFieldCustom]; !ok {
				return
			}
		}

	} else {
		if sourceFlow, ok = m.SfcSourceFlowBot[defaultFieldCustom]; !ok {
			return
		}

	}

	providerConf := sourceFlow.Providers[string(provider)]
	buttonID = providerConf.ButtonID
	ownerID = providerConf.OwnerID
	subject = sourceFlow.Subject

	return
}

func NewInterconectionCache(interconnection *Interconnection) cache.Interconnection {
	return cache.Interconnection{
		UserID:        interconnection.UserID,
		SessionID:     interconnection.SessionID,
		SessionKey:    interconnection.SessionKey,
		AffinityToken: interconnection.AffinityToken,
		Status:        string(interconnection.Status),
		Timestamp:     interconnection.Timestamp,
		Provider:      string(interconnection.Provider),
		BotSlug:       interconnection.BotSlug,
		BotID:         interconnection.BotID,
		Name:          interconnection.Name,
		Email:         interconnection.Email,
		PhoneNumber:   interconnection.PhoneNumber,
		CaseID:        interconnection.CaseID,
		ExtraData:     interconnection.ExtraData,
	}
}

// SaveContext method will save context of integration message from facebook
func (m *Manager) SaveContextFB(integration *models.IntegrationsFacebook) error {
	logrus.Info("WEBHOOK FACEBOOK: ", integration)
	errorsMessage := []string{}
	var err error
	isSend := false
	for _, entry := range integration.Message.Entry {
		for _, message := range entry.Messaging {
			userID := message.Recipient.ID
			from := fromBot
			if m.cacheMessage.IsRepeatedMessage(message.Message.Mid) {
				continue
			}
			if integration.AuthorRole == fromUser {
				isSend, err = m.salesforceComunicationFB(message)
				if err != nil {
					errorsMessage = append(errorsMessage, err.Error())
				}
				userID = message.Sender.ID
				from = fromUser
			}
			if isSend || message.Message.Text == "" {
				continue
			}

			ctx := cache.Context{
				UserID:    userID,
				Timestamp: message.Timestamp,
				From:      from,
				Text:      message.Message.Text,
			}
			err = m.contextcache.StoreContext(ctx)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"context": ctx,
				}).WithError(err).Error("Error store context")
				errorsMessage = append(errorsMessage, err.Error())
			}
		}

	}

	if len(errorsMessage) > 0 {
		return errors.New(strings.Join(errorsMessage, "|"))
	}

	return nil
}

func (m *Manager) salesforceComunicationFB(message models.Messaging) (bool, error) {
	isSend := false
	if interconnection, ok := m.validInterconnection(message.Sender.ID); ok && interconnection.Status == Active {
		switch {
		case message.Message.Text != "":
			if strings.Contains(constants.DevEnvironments, m.environment) {
				for _, keyword := range m.keywordsRestart {
					if strings.ToLower(message.Message.Text) == keyword {
						err := m.SalesforceService.EndChat(interconnection.AffinityToken, interconnection.SessionKey)
						if err != nil {
							return false, err
						}
						interconnection.updateStatusRedis(string(Closed))
						interconnection.Status = Closed
						interconnection.runnigLongPolling = false
						return true, nil
					}
				}
			}

			interconnection.salesforceChannel <- NewSfMessage(interconnection.AffinityToken, interconnection.SessionKey, message.Message.Text)

			isSend = true
		case message.Message.Attachments != nil:
			for _, attachment := range message.Message.Attachments {
				if attachment.Type == imageType {
					err := m.SalesforceService.InsertImageInCase(
						attachment.Payload.URL,
						interconnection.SessionID,
						"",
						interconnection.CaseID)
					if err != nil {
						interconnection.integrationsChannel <- NewIntegrationsMessage(message.Sender.ID, messageError, FacebookProvider)

						return false, err
					}
					interconnection.salesforceChannel <- NewSfMessage(interconnection.AffinityToken, interconnection.SessionKey, messageImageSuccess)
					isSend = true
				}
			}
		}
	}

	return isSend, nil
}
