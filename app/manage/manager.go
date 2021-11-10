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
	"yalochat.com/salesforce-integration/app/cron"
	"yalochat.com/salesforce-integration/app/services"
	"yalochat.com/salesforce-integration/base/cache"
	"yalochat.com/salesforce-integration/base/clients/botrunner"
	"yalochat.com/salesforce-integration/base/clients/chat"
	"yalochat.com/salesforce-integration/base/clients/integrations"
	"yalochat.com/salesforce-integration/base/clients/login"
	"yalochat.com/salesforce-integration/base/clients/proxy"
	"yalochat.com/salesforce-integration/base/clients/salesforce"
	"yalochat.com/salesforce-integration/base/clients/studiong"
	"yalochat.com/salesforce-integration/base/constants"
	"yalochat.com/salesforce-integration/base/helpers"
	"yalochat.com/salesforce-integration/base/models"
)

var (
	SfcOrganizationID   string
	SfcDeploymentID     string
	SfcRecordTypeID     string
	BlockedUserState    map[string]string
	TimeoutState        map[string]string
	SuccessState        map[string]string
	SfcCustomFieldsCase map[string]string
	BotrunnerTimeout    int
	//TODO: move a integration clients constructor
	WAPhone         string
	FBPhone         string
	WebhookBaseUrl  string
	WebhookWhatsapp string
	WebhookFacebook string
	StudioNGTimeout int
	CodePhoneRemove []string
)

const (
	audioType           = "audio"
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
	client                string
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
	StudioNG              studiong.StudioNGInterface
	isStudioNGFlow        bool
	maxRetries            int
}

// ManagerOptions holds configurations for the interactions manager
type ManagerOptions struct {
	AppName                    string
	Client                     string
	BlockedUserState           map[string]string
	TimeoutState               map[string]string
	SuccessState               map[string]string
	RedisOptions               cache.RedisOptions
	BotrunnerUrl               string
	BotrunnerToken             string
	BotrunnerTimeout           int
	SfcClientID                string
	SfcClientSecret            string
	SfcUsername                string
	SfcPassword                string
	SfcSecurityToken           string
	SfcBaseUrl                 string
	SfcChatUrl                 string
	SfcLoginUrl                string
	SfcApiVersion              string
	SfcOrganizationID          string
	SfcDeploymentID            string
	SfcRecordTypeID            string
	SfcAccountRecordTypeID     string
	SfcDefaultBirthDateAccount string
	SfcCustomFieldsCase        map[string]string
	SfcCodePhoneRemove         []string
	IntegrationsUrl            string
	IntegrationsWAChannel      string
	IntegrationsFBChannel      string
	IntegrationsWAToken        string
	IntegrationsFBToken        string
	IntegrationsWABotID        string
	IntegrationsFBBotID        string
	IntegrationsSignature      string
	IntegrationsWABotPhone     string
	IntegrationsFBBotPhone     string
	WebhookBaseUrl             string
	WebhookWhatsapp            string
	WebhookFacebook            string
	Environment                string
	KeywordsRestart            []string
	SfcSourceFlowBot           envs.SfcSourceFlowBot
	SfcSourceFlowField         string
	SfcBlockedChatField        bool
	StudioNGUrl                string
	StudioNGToken              string
	StudioNGTimeout            int
	SpecSchedule               string
	MaxRetries                 int
	CleanContextSchedule       string
}

type ManagerI interface {
	SaveContext(integration *models.IntegrationsRequest) error
	CreateChat(interconnection *Interconnection) error
	GetContextByUserID(userID string) []cache.Context
	SaveContextFB(integration *models.IntegrationsFacebook) error
	FinishChat(userId string) error
	RegisterWebhookInIntegrations(provider string) error
	RemoveWebhookInIntegrations(provider string) error
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
	WAPhone = config.IntegrationsWABotPhone
	FBPhone = config.IntegrationsFBBotPhone
	WebhookBaseUrl = config.WebhookBaseUrl
	WebhookFacebook = config.WebhookFacebook
	WebhookWhatsapp = config.WebhookWhatsapp
	StudioNGTimeout = config.StudioNGTimeout
	CodePhoneRemove = config.SfcCodePhoneRemove
	isStudioNG := false

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

	var studioNG *studiong.StudioNG = nil
	if config.StudioNGUrl != "" {
		studioNG = studiong.NewStudioNGClient(
			config.StudioNGUrl,
			config.StudioNGToken,
		)
		isStudioNG = true
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

	salesforceService := services.NewSalesforceService(*sfcLoginClient,
		*sfcChatClient,
		*salesforceClient,
		tokenPayload,
		config.SfcCustomFieldsCase,
		SfcRecordTypeID)

	if config.SpecSchedule != "" {
		cronService := cron.NewCron(salesforceService, config.SpecSchedule, config.SfcUsername)

		if config.CleanContextSchedule != "" {
			cronService.Contextschedule = config.CleanContextSchedule
			cronService.Client = config.Client
			cronService.ContextCache = contextCache
		}
		cronService.Run()
	}

	salesforceService.AccountRecordTypeId = config.SfcAccountRecordTypeID
	if config.SfcDefaultBirthDateAccount != "" {
		salesforceService.DefaultBirthDateAccount = config.SfcDefaultBirthDateAccount
	}

	var botRunnerClient *botrunner.BotRunner = nil
	if config.BotrunnerUrl != "" {
		botRunnerClient = botrunner.NewBotrunnerClient(config.BotrunnerUrl, config.BotrunnerToken)
	}

	interconnections := interconnectionsCache.RetrieveAllInterconnections(config.Client)

	cacheLocal := cache.New()
	m := &Manager{
		clientName:            config.AppName,
		client:                config.Client,
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
		StudioNG:              studioNG,
		isStudioNGFlow:        isStudioNG,
		maxRetries:            config.MaxRetries,
	}

	for _, interconnection := range *interconnections {
		if InterconnectionStatus(interconnection.Status) == Active || interconnection.Status == string(OnHold) {
			in := convertInterconnectionCacheToInterconnection(interconnection)
			m.AddInterconnection(in)
		}
	}

	go m.handleInterconnection()
	go m.handleMessageToSalesforce()
	go m.handleMessageToUsers()
	return m
}

// handleInterconnection This function terminates the interconnections.
func (m *Manager) handleInterconnection() {
	for {
		select {
		case interconection := <-m.finishInterconnection:
			logrus.WithField("userID", interconection.UserID).Info("Finish interconnection")
			m.EndChat(interconection)
		default:

		}
	}
}

// handleMessageToSalesforce This function sends messages to salesforce agents.
func (m *Manager) handleMessageToSalesforce() {
	for {
		select {
		case messageSf := <-m.salesforceChannel:
			logrus.WithField("userID", messageSf.UserID).Info("Message to agent from user")
			m.sendMessageToSalesforce(messageSf)
		default:

		}
	}
}

// handleMessageToUsers This function sends messages to users.
func (m *Manager) handleMessageToUsers() {
	for {
		select {
		case messageInt := <-m.integrationsChannel:
			logrus.WithField("userID", messageInt.UserID).Info("Message to user from agent")
			m.sendMessageToUser(messageInt)
		default:
			
		}
	}
}

func (m *Manager) sendMessageToUser(message *Message) {
	retries := 0
	for {
		switch message.Provider {
		case WhatsappProvider:
			_, err := m.IntegrationsClient.SendMessage(integrations.SendTextPayload{
				Id:     helpers.RandomString(36),
				Type:   "text",
				UserID: message.UserID,
				Text:   integrations.TextMessage{Body: message.Text},
			}, string(message.Provider))
			if err != nil {
				logrus.WithField("userId", message.UserID).Error(helpers.ErrorMessage("Error sendMessage to user", err))
				if retries == m.maxRetries {
					logrus.WithField("userId", message.UserID).Error("Error sendMessage to user, max retries")
					return
				}
				retries++
				continue
			}
			logrus.Infof("Send message to UserID : %s", message.UserID)
			return
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
				logrus.WithField("userId", message.UserID).Error(helpers.ErrorMessage("Error sendMessage to user", err))
				if retries == m.maxRetries {
					logrus.WithField("userId", message.UserID).Error("Error sendMessage to user, max retries")
					return
				}
				retries++
				continue
			}
			logrus.Infof("Send message to UserID : %s", message.UserID)
			return
		}
	}
}

func (m *Manager) sendMessageToSalesforce(message *Message) {
	retries := 0
	for {
		_, err := m.SalesforceService.SendMessage(message.AffinityToken, message.SessionKey, chat.MessagePayload{Text: message.Text})
		if err != nil {
			logrus.WithField("userID", message.UserID).Error(helpers.ErrorMessage("Error sendMessage to salesforce", err))
			if retries == m.maxRetries {
				logrus.WithField("userID", message.UserID).Error("Error sendMessage to salesforce, max retries")
				return
			}
			retries++
			continue
		}
		logrus.Infof("Send message to agent from salesforce : %s", message.UserID)
		return
	}
}

// CreateChat Initialize a chat with salesforce
func (m *Manager) CreateChat(interconnection *Interconnection) error {
	titleMessage := "could not create chat in salesforce"
	interconnection.Client = m.client

	// Validate that user does not have an active session
	logrus.WithField("userID", interconnection.UserID).Info("Validate UserID")
	err := m.ValidateUserID(interconnection.UserID)
	if err != nil {
		logrus.WithField("interconnection", interconnection).WithError(err).Error("error ValidateUserID")
		return errors.New(helpers.ErrorMessage(titleMessage, err))
	}

	// Clean phoneNumber
	logrus.WithField("userID", interconnection.UserID).Info("cleanPrefixPhoneNumber")
	cleanPrefixPhoneNumber(interconnection)

	// We get the contact if it exists by your email or phone.
	logrus.WithField("userID", interconnection.UserID).Info("GetOrCreateContact")
	contact, err := m.SalesforceService.GetOrCreateContact(interconnection.Name, interconnection.Email, interconnection.PhoneNumber)
	if err != nil {
		logrus.WithField("interconnection", interconnection).WithError(err).Error("error GetOrCreateContact")
		return errors.New(helpers.ErrorMessage(titleMessage, err))
	}

	if contact.Blocked {
		logrus.WithField("userID", interconnection.UserID).Info("User Blocked")
		go ChangeToState(interconnection.UserID, interconnection.BotSlug, BlockedUserState[string(interconnection.Provider)], m.BotrunnnerClient, BotrunnerTimeout, StudioNGTimeout, m.StudioNG, m.isStudioNGFlow)
		return fmt.Errorf("%s: %s", "could not create chat in salesforce", "this contact is blocked")
	}

	buttonID, ownerID, subject := m.changeButtonIDAndOwnerID(interconnection.Provider, interconnection.ExtraData)

	logrus.WithField("userID", interconnection.UserID).Info("CreateCase")
	caseId, err := m.SalesforceService.CreatCase(contact.ID, descriptionDefualt, subject, string(interconnection.Provider), ownerID,
		interconnection.ExtraData)
	if err != nil {
		logrus.WithField("interconnection", interconnection).WithError(err).Error("error CreatCase")
		return errors.New(helpers.ErrorMessage(titleMessage, err))
	}
	interconnection.CaseID = caseId

	//Creating chat in Salesforce
	logrus.WithField("userID", interconnection.UserID).Info("CreateChat")
	session, err := m.SalesforceService.CreatChat(interconnection.Name, SfcOrganizationID, SfcDeploymentID, buttonID, caseId, contact.ID)
	if err != nil {
		logrus.WithField("interconnection", interconnection).WithError(err).Error("error CreatChat")
		return errors.New(helpers.ErrorMessage(titleMessage, err))
	}

	logrus.WithField("userID", interconnection.UserID).Info("GetContext Routine")
	go m.getContext(interconnection)
	interconnection.AffinityToken = session.AffinityToken
	interconnection.SessionID = session.Id
	interconnection.SessionKey = session.Key
	interconnection.Status = OnHold
	interconnection.Timestamp = time.Now()
	time.Sleep(time.Second * 1)

	//Add interconection to Redis and interconnectionMap
	logrus.WithField("userID", interconnection.UserID).Info("AddInterconnection")
	m.AddInterconnection(interconnection)
	return nil
}

func cleanPrefixPhoneNumber(interconnection *Interconnection) {
	if len(interconnection.PhoneNumber) > 10 {
		for _, code := range CodePhoneRemove {
			if strings.HasPrefix(interconnection.PhoneNumber, code) {
				newPhoneNumber := strings.TrimPrefix(interconnection.PhoneNumber, code)
				if len(newPhoneNumber) < 10 {
					continue
				}
				interconnection.PhoneNumber = newPhoneNumber
				break
			}
		}
	}
}

// ChangeToState Change to state with botrunner
func ChangeToState(userID, botSlug, state string, botRunnerClient botrunner.BotRunnerInterface, seconds, secondsNG int, studioNGClient studiong.StudioNGInterface, isStudio bool) {
	if !isStudio {
		time.Sleep(time.Second * time.Duration(seconds))
		_, err := botRunnerClient.SendTo(botrunner.GetRequestToSendTo(botSlug, userID, state, ""))
		if err != nil {
			logrus.Errorf(helpers.ErrorMessage(fmt.Sprintf("could not sent to state: %s", state), err))
		}
		return
	}

	time.Sleep(time.Second * time.Duration(secondsNG))
	err := studioNGClient.SendTo(state, userID)
	if err != nil {
		logrus.Errorf(helpers.ErrorMessage(fmt.Sprintf("could not sent to state: %s with studioNG client", state), err))
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

	in.runnigLongPolling = false

	in.updateStatusRedis(string(Closed))
	m.EndChat(in)
	return nil
}

func (m *Manager) ValidateUserID(userID string) error {
	sessionRedis, _ := m.interconnectionsCache.RetrieveInterconnection(cache.Interconnection{
		UserID: userID,
		Client: m.client,
	})

	if sessionRedis != nil && (sessionRedis.Status == string(Active) || sessionRedis.Status == string(OnHold)) {
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
	interconnection.StudioNG = m.StudioNG
	interconnection.isStudioNGFlow = m.isStudioNGFlow

	go m.storeInterconnectionInRedis(interconnection)

	m.interconnectionMap.Set(fmt.Sprintf(constants.UserKey, interconnection.UserID), interconnection, 0)

	go interconnection.handleLongPolling()
	go interconnection.handleStatus()
	logrus.Infof("Create interconnection successfully : %s", interconnection.UserID)
}

func (m *Manager) storeInterconnectionInRedis(interconnection *Interconnection) {
	err := m.interconnectionsCache.StoreInterconnection(NewInterconectionCache(interconnection))
	if err != nil {
		logrus.Errorf("Could not store interconnection with userID[%s] in redis[%s]", interconnection.UserID, interconnection.SessionID)
	}
}

func (m *Manager) getContext(interconnection *Interconnection) {
	interconnection.Context = "Contexto:\n" + m.getContextByUserID(interconnection.UserID)
	logrus.Infof("Get context of userID : %s", interconnection.UserID)
}

// SaveContext method will save context of integration message
func (m *Manager) SaveContext(integration *models.IntegrationsRequest) error {
	//logrus.Info("WEBHOOK WHATSAPP: ", integration)
	if m.cacheMessage.IsRepeatedMessage(integration.ID) {
		return nil
	}

	isSend := m.salesforceComunication(integration)
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

	ctx := &cache.Context{
		Client:    m.client,
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
	case integration.Type == audioType:
		ctx.URL = integration.Audio.URL
		ctx.MIMEType = integration.Audio.MIMEType
	default:
		return nil
	}

	if integration.To != "" {
		ctx.From = fromBot
		ctx.UserID = integration.To
	}

	go m.saveContextInRedis(ctx)
	return nil
}

func (m *Manager) salesforceComunication(integration *models.IntegrationsRequest) bool {
	interconnection, ok := m.validInterconnection(integration.From)
	isInterconnectionActive := ok && interconnection.Status == Active
	if isInterconnectionActive {
		logrus.WithField("userID", interconnection.UserID).Info("Send Message of user with chat")
		go m.sendMessageComunication(interconnection, integration)
	}
	return isInterconnectionActive
}

func (m *Manager) sendMessageComunication(interconnection *Interconnection, integration *models.IntegrationsRequest) {
	switch integration.Type {
	case textType:
		if strings.Contains(constants.DevEnvironments, m.environment) {
			for _, keyword := range m.keywordsRestart {
				if strings.ToLower(integration.Text.Body) == keyword {
					err := m.SalesforceService.EndChat(interconnection.AffinityToken, interconnection.SessionKey)
					if err != nil {
						logrus.WithFields(logrus.Fields{"interconnection": interconnection}).WithError(err).Error("EndChat error")
						return
					}
					interconnection.updateStatusRedis(string(Closed))
					interconnection.Status = Closed
					interconnection.runnigLongPolling = false
					return
				}
			}
		}
		interconnection.salesforceChannel <- NewSfMessage(interconnection.AffinityToken, interconnection.SessionKey, integration.Text.Body, interconnection.UserID)

	case imageType:
		err := m.SalesforceService.InsertImageInCase(
			integration.Image.URL,
			interconnection.SessionID,
			integration.Image.MIMEType,
			interconnection.CaseID)
		if err != nil {
			logrus.WithFields(logrus.Fields{"interconnection": interconnection}).WithError(err).Error("InsertImageInCase error")
			interconnection.integrationsChannel <- NewIntegrationsMessage(integration.From, messageError, WhatsappProvider)
			return
		}
		logrus.WithField("userID", interconnection.UserID).Info("Send Image to agent")
		interconnection.salesforceChannel <- NewSfMessage(interconnection.AffinityToken, interconnection.SessionKey, messageImageSuccess, interconnection.UserID)
	}
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
	allContext := m.contextcache.RetrieveContextFromSet(m.client, userID)

	sort.Slice(allContext, func(i, j int) bool { return allContext[j].Timestamp > allContext[i].Timestamp })
	builder := strings.Builder{}
	for _, ctx := range allContext {
		if ctx.Ttl.Before(time.Now()) {
			continue
		}

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
	allContext := m.contextcache.RetrieveContextFromSet(m.client, userID)

	sort.Slice(allContext, func(i, j int) bool { return allContext[j].Timestamp > allContext[i].Timestamp })

	for i, ctx := range allContext {
		if ctx.Ttl.Before(time.Now()) {
			allContext = append(allContext[:i], allContext[i+1:]...)
		}
	}

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
		Client:        interconnection.Client,
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

// SaveContextFB method will save context of integration message from facebook
func (m *Manager) SaveContextFB(integration *models.IntegrationsFacebook) error {
	//logrus.Info("WEBHOOK FACEBOOK: ", integration)
	isSend := false
	for _, entry := range integration.Message.Entry {
		for _, message := range entry.Messaging {
			userID := message.Recipient.ID
			from := fromBot
			if m.cacheMessage.IsRepeatedMessage(message.Message.Mid) {
				continue
			}
			if integration.AuthorRole == fromUser {
				isSend = m.salesforceComunicationFB(message)
				userID = message.Sender.ID
				from = fromUser
			}
			if isSend || message.Message.Text == "" {
				continue
			}

			ctx := &cache.Context{
				UserID:    userID,
				Timestamp: message.Timestamp,
				From:      from,
				Text:      message.Message.Text,
				Client:    m.client,
			}
			go m.saveContextInRedis(ctx)
		}
	}
	return nil
}

func (m *Manager) salesforceComunicationFB(message models.Messaging) bool {
	interconnection, ok := m.validInterconnection(message.Sender.ID)
	isInterconnectionActive := ok && interconnection.Status == Active
	if isInterconnectionActive {
		logrus.WithField("userID", interconnection.UserID).Info("Send Message of user with chat FB")
		go m.sendMessageComunicationFB(interconnection, &message)
	}

	return isInterconnectionActive
}

func (m *Manager) sendMessageComunicationFB(interconnection *Interconnection, message *models.Messaging) {
	switch {
	case message.Message.Text != "":
		if strings.Contains(constants.DevEnvironments, m.environment) {
			for _, keyword := range m.keywordsRestart {
				if strings.ToLower(message.Message.Text) == keyword {
					err := m.SalesforceService.EndChat(interconnection.AffinityToken, interconnection.SessionKey)
					if err != nil {
						logrus.WithFields(logrus.Fields{"interconnection": interconnection}).WithError(err).Error("InsertImageInCase error")
						return
					}
					interconnection.updateStatusRedis(string(Closed))
					interconnection.Status = Closed
					interconnection.runnigLongPolling = false
					return
				}
			}
		}
		interconnection.salesforceChannel <- NewSfMessage(interconnection.AffinityToken, interconnection.SessionKey, message.Message.Text, interconnection.UserID)

	case message.Message.Attachments != nil:
		for _, attachment := range message.Message.Attachments {
			if attachment.Type == imageType {
				err := m.SalesforceService.InsertImageInCase(
					attachment.Payload.URL,
					interconnection.SessionID,
					"",
					interconnection.CaseID)
				if err != nil {
					logrus.WithFields(logrus.Fields{"interconnection": interconnection}).WithError(err).Error("InsertImageInCase error")
					interconnection.integrationsChannel <- NewIntegrationsMessage(message.Sender.ID, messageError, FacebookProvider)
					return
				}
				logrus.WithField("userID", interconnection.UserID).Info("FB Send Image to agent")
				interconnection.salesforceChannel <- NewSfMessage(interconnection.AffinityToken, interconnection.SessionKey, messageImageSuccess, interconnection.UserID)
			}
		}
	}
}

func (m *Manager) RegisterWebhookInIntegrations(provider string) error {

	switch provider {
	case string(WhatsappProvider):
		_, err := m.IntegrationsClient.WebhookRegister(integrations.HealthcheckPayload{
			Phone:    WAPhone,
			Webhook:  WebhookBaseUrl + WebhookWhatsapp,
			Provider: string(WhatsappProvider),
		})
		if err != nil {
			errorMessage := fmt.Sprintf("could not set whatsapp webhook on integrations : %s", err.Error())
			logrus.Error(errorMessage)
			return errors.New(errorMessage)
		}
		return nil

	case string(FacebookProvider):
		_, err := m.IntegrationsClient.WebhookRegister(integrations.HealthcheckPayload{
			Phone:    FBPhone,
			Webhook:  WebhookBaseUrl + WebhookFacebook,
			Version:  3,
			Provider: string(FacebookProvider),
		})
		if err != nil {
			errorMessage := fmt.Sprintf("could not set facebook webhook on integrations : %s", err.Error())
			logrus.Error(errorMessage)
			return errors.New(errorMessage)
		}
		return nil
	default:
		errorMessage := fmt.Sprintf("Invalid provider webhook : %s", provider)
		logrus.Error(errorMessage)
		return errors.New(errorMessage)
	}
}

func (m *Manager) RemoveWebhookInIntegrations(provider string) error {

	switch provider {
	case string(WhatsappProvider):
		_, err := m.IntegrationsClient.WebhookRemove(integrations.RemoveWebhookPayload{
			Phone:    WAPhone,
			Provider: string(WhatsappProvider),
		})
		if err != nil {
			errorMessage := fmt.Sprintf("could not remove whatsapp webhook on integrations : %s", err.Error())
			logrus.Error(errorMessage)
			return errors.New(errorMessage)
		}
		return nil

	case string(FacebookProvider):
		_, err := m.IntegrationsClient.WebhookRemove(integrations.RemoveWebhookPayload{
			Phone:    FBPhone,
			Provider: string(FacebookProvider),
		})
		if err != nil {
			errorMessage := fmt.Sprintf("could not remove facebook webhook on integrations : %s", err.Error())
			logrus.Error(errorMessage)
			return errors.New(errorMessage)
		}
		return nil
	default:
		errorMessage := fmt.Sprintf("Invalid provider webhook : %s", provider)
		logrus.Error(errorMessage)
		return errors.New(errorMessage)
	}
}

func (m *Manager) saveContextInRedis(ctx *cache.Context) {
	ctx.Ttl = time.Now().Add(cache.Ttl)
	err := m.contextcache.StoreContextToSet(*ctx)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"context": ctx,
		}).WithError(err).Error("Error store context in set")
	}
}
