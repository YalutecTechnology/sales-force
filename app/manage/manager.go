package manage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"yalochat.com/salesforce-integration/base/events"
	"yalochat.com/salesforce-integration/base/subscribers"
	"yalochat.com/salesforce-integration/base/subscribers/kafka"

	"golang.org/x/time/rate"

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
	WAPhone                string
	FBPhone                string
	WebhookBaseUrl         string
	WebhookWhatsapp        string
	WebhookFacebook        string
	StudioNGTimeout        int
	CodePhoneRemove        []string
	Messages               models.MessageTemplate
	Timezone               string
	SendImageNameInMessage bool
)

const (
	fromUser           = "user"
	fromBot            = "bot"
	defaultFieldCustom = "default"
)

// Manager controls the process of the app
type Manager struct {
	clientName                   string
	client                       string
	interconnectionMap           cache.ICache
	SalesforceService            services.SalesforceServiceInterface
	IntegrationsClient           integrations.IntegrationInterface
	BotrunnnerClient             botrunner.BotRunnerInterface
	finishInterconnection        chan *Interconnection
	contextcache                 cache.IContextCache
	interconnectionsCache        cache.IInterconnectionCache
	environment                  string
	keywordsRestart              []string
	cacheMessage                 cache.IMessageCache
	SfcSourceFlowBot             envs.SfcSourceFlowBot
	SfcSourceFlowField           string
	StudioNG                     studiong.StudioNGInterface
	isStudioNGFlow               bool
	maxRetries                   int
	IntegrationChanRateLimit     int
	IntegrationChanRateLimiter   *rate.Limiter
	SalesforceChanRequestLimit   int
	SalesforceChanRequestLimiter *rate.Limiter
	kafkaProducer                subscribers.Producer
	KafkaTopic                   string
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
	SfcCustomFieldsContact     map[string]string
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
	IntegrationsRateLimit      float64
	SalesforceRateLimit        float64
	Messages                   models.MessageTemplate
	Timezone                   string
	SendImageNameInMessage     bool
	KafkaHost                  string
	KafkaPort                  string
	KafkaUser                  string
	KafkaPassword              string
	KafkaTopic                 string
}

type ManagerI interface {
	SaveContext(ctx context.Context, integration *models.IntegrationsRequest) error
	CreateChat(ctx context.Context, interconnection *Interconnection) error
	GetContextByUserID(userID string) []cache.Context
	SaveContextFB(ctx context.Context, integration *models.IntegrationsFacebook) error
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
	Messages = config.Messages
	Timezone = config.Timezone
	SendImageNameInMessage = config.SendImageNameInMessage

	salesforceRateLimit := rate.Limit(config.SalesforceRateLimit)
	salesforceRateLimiter := rate.NewLimiter(salesforceRateLimit, int(salesforceRateLimit)+1)

	integrationsRateLimit := rate.Limit(config.IntegrationsRateLimit)
	integrationsRateLimiter := rate.NewLimiter(integrationsRateLimit, int(integrationsRateLimit)+1)

	redisCache, err := cache.NewRedisCache(&config.RedisOptions)
	if err != nil {
		logrus.WithError(err).Error("Error initializing Context Redis Manager")
	}

	var contextCache *cache.ContextCache
	var interconnectionsCache *cache.InterconnectionCache

	if redisCache != nil {
		contextCache = cache.NewContextCache(redisCache)
		interconnectionsCache = cache.NewInterconnectionCache(redisCache)
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
		SfcRecordTypeID,
		Messages.FirstNameContact,
		config.SfcCustomFieldsContact,
	)

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

	cacheLocal := cache.New()
	m := &Manager{
		clientName:                   config.AppName,
		client:                       config.Client,
		SalesforceService:            salesforceService,
		interconnectionMap:           cacheLocal,
		IntegrationsClient:           integrationsClient,
		finishInterconnection:        make(chan *Interconnection),
		contextcache:                 contextCache,
		interconnectionsCache:        interconnectionsCache,
		BotrunnnerClient:             botRunnerClient,
		environment:                  config.Environment,
		keywordsRestart:              config.KeywordsRestart,
		SfcSourceFlowBot:             config.SfcSourceFlowBot,
		SfcSourceFlowField:           config.SfcSourceFlowField,
		cacheMessage:                 cache.NewMessageCache(cacheLocal),
		StudioNG:                     studioNG,
		isStudioNGFlow:               isStudioNG,
		maxRetries:                   config.MaxRetries,
		IntegrationChanRateLimiter:   integrationsRateLimiter,
		SalesforceChanRequestLimiter: salesforceRateLimiter,
		KafkaTopic:                   config.KafkaTopic,
	}

	// TODO: Add function restore interconnections
	if !reflect.ValueOf(m.interconnectionsCache).IsNil() {
		interconnections := interconnectionsCache.RetrieveAllInterconnections(config.Client)
		ctx := context.Background()
		for _, interconnection := range *interconnections {
			if InterconnectionStatus(interconnection.Status) == Active || interconnection.Status == string(OnHold) {
				in := convertInterconnectionCacheToInterconnection(interconnection)
				m.AddInterconnection(ctx, in)
			}
		}
	}

	go m.handleInterconnection()

	if config.KafkaUser != "" {
		producer := kafka.NewProducer(kafka.KafkaSettings{
			Host:     config.KafkaHost,
			Port:     config.KafkaPort,
			User:     config.KafkaUser,
			Password: config.KafkaPassword,
		})

		m.kafkaProducer = producer

		consumer := kafka.NewConsumer(m, m.KafkaTopic, config.AppName, constants.Latest, kafka.KafkaSettings{
			Host:     config.KafkaHost,
			Port:     config.KafkaPort,
			User:     config.KafkaUser,
			Password: config.KafkaPassword,
		})
		go consumer.Start()
	}

	return m
}

// handleInterconnection This function terminates the interconnections.
func (m *Manager) handleInterconnection() {
	for interconection := range m.finishInterconnection {
		logrus.WithField("userID", interconection.UserID).Info("Finish interconnection")
		m.EndChat(interconection)
	}

}

func (m *Manager) sendMessageToUser(message *Message) {
	// datadog tracing
	spanContext := events.GetSpanContextFromSpan(message.MainSpan)
	span := tracer.StartSpan("sendMessageToUser", tracer.ChildOf(spanContext))
	span.SetTag(ext.AnalyticsEvent, true)
	span.SetTag(events.MessageIntegrations, fmt.Sprintf("%#v", message))
	span.SetTag(events.UserID, message.UserID)
	span.SetTag(events.Provider, message.Provider)
	span.SetTag(events.SendMessage, false)
	span.SetTag(events.RetryMessage, false)
	defer span.Finish()

	retries := 0
	for {
		switch message.Provider {
		case WhatsappProvider:
			_, err := m.IntegrationsClient.SendMessage(integrations.SendTextPayload{
				Id:     message.ID,
				Type:   "text",
				UserID: message.UserID,
				Text:   integrations.TextMessage{Body: message.Text},
			}, string(message.Provider))
			if err != nil {
				span.SetTag(ext.Error, err)
				logrus.WithField(events.UserID, message.UserID).Error(helpers.ErrorMessage("Error sendMessage to user", err))
				if retries == m.maxRetries {
					logrus.WithField(events.UserID, message.UserID).Error("Error sendMessage to user, max retries")
					return
				}
				retries++
				span.SetTag(events.RetryMessage, true)
				continue
			}
			logrus.Infof("Send message to UserID : %s", message.UserID)
			span.SetTag(events.SendMessage, true)
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
				span.SetTag(ext.Error, err)
				logrus.WithField(events.UserID, message.UserID).Error(helpers.ErrorMessage("Error sendMessage to user", err))
				if retries == m.maxRetries {
					logrus.WithField(events.UserID, message.UserID).Error("Error sendMessage to user, max retries")
					return
				}
				retries++
				span.SetTag(events.RetryMessage, true)
				continue
			}
			logrus.Infof("Send message to UserID : %s", message.UserID)
			span.SetTag(events.SendMessage, true)
			return
		}
	}
}

func (m *Manager) sendMessageToSalesforce(message *Message) {
	// datadog tracing
	spanContext := events.GetSpanContextFromSpan(message.MainSpan)
	span := tracer.StartSpan("sendMessageToSalesforce", tracer.ChildOf(spanContext))
	span.SetTag(ext.AnalyticsEvent, true)
	span.SetTag(events.MessageSalesforce, fmt.Sprintf("%#v", message))
	span.SetTag(events.UserID, message.UserID)
	span.SetTag(events.SendMessage, false)
	span.SetTag(events.RetryMessage, false)
	defer span.Finish()
	retries := 0
	for {
		_, err := m.SalesforceService.SendMessage(span, message.AffinityToken, message.SessionKey, chat.MessagePayload{Text: message.Text})
		if err != nil {
			span.SetTag(ext.Error, err)
			logrus.WithField("userID", message.UserID).Error(helpers.ErrorMessage("Error sendMessage to salesforce", err))
			if retries == m.maxRetries {
				logrus.WithField("userID", message.UserID).Error("Error sendMessage to salesforce, max retries")
				return
			}
			span.SetTag(events.RetryMessage, true)
			retries++
			continue
		}
		logrus.Infof("Send message to agent from salesforce : %s", message.UserID)
		span.SetTag(events.SendMessage, true)
		return
	}
}

// CreateChat Initialize a chat with salesforce
func (m *Manager) CreateChat(ctx context.Context, interconnection *Interconnection) error {
	// datadog tracing
	span, _ := tracer.StartSpanFromContext(ctx, "manager.CreateChat")
	span.SetTag(ext.AnalyticsEvent, true)
	span.SetTag(events.UserID, interconnection.UserID)
	span.SetTag(events.Provider, interconnection.Provider)
	span.SetTag(events.Interconnection, fmt.Sprintf("%#v", interconnection))
	defer span.Finish()

	titleMessage := "could not create chat in salesforce"
	interconnection.Client = m.client
	span.SetTag(events.Client, interconnection.Client)

	logFields := logrus.Fields{
		constants.TraceIdKey: span.Context().TraceID(),
		constants.SpanIdKey:  span.Context().SpanID(),
		events.UserID:        interconnection.UserID,
	}

	// Validate that user does not have an active session
	logrus.WithFields(logFields).Info("Validate UserID")
	err := m.ValidateUserID(ctx, interconnection.UserID)
	if err != nil {
		logrus.WithFields(logFields).WithError(err).Error("error ValidateUserID")
		span.SetTag(ext.Error, err)
		span.SetTag(ext.ErrorDetails, helpers.ErrorMessage("error ValidateUserID", err))
		return errors.New(helpers.ErrorMessage(titleMessage, err))
	}

	// Clean phoneNumber
	logrus.WithFields(logFields).Info("cleanPrefixPhoneNumber")
	cleanPrefixPhoneNumber(interconnection)

	// We get the contact if it exists by your email or phone.
	logrus.WithFields(logFields).Info("GetOrCreateContact")
	contact, err := m.SalesforceService.GetOrCreateContact(ctx, interconnection.Name, interconnection.Email, interconnection.PhoneNumber, interconnection.ExtraData)
	if err != nil {
		logrus.WithFields(logFields).WithError(err).Error("error GetOrCreateContact")
		span.SetTag(ext.Error, err)
		go ChangeToState(interconnection.UserID, interconnection.BotSlug, TimeoutState[string(interconnection.Provider)], m.BotrunnnerClient, BotrunnerTimeout, StudioNGTimeout, m.StudioNG, m.isStudioNGFlow)
		return errors.New(helpers.ErrorMessage(titleMessage, err))
	}

	span.SetTag("Contact", contact)
	logFields["contactId"] = contact.ID
	if contact.Blocked {
		logrus.WithFields(logFields).Info("User Blocked")
		span.SetTag(events.UserBlocked, true)
		go ChangeToState(interconnection.UserID, interconnection.BotSlug, BlockedUserState[string(interconnection.Provider)], m.BotrunnnerClient, BotrunnerTimeout, StudioNGTimeout, m.StudioNG, m.isStudioNGFlow)
		return fmt.Errorf("%s: %s", "could not create chat in salesforce", "this contact is blocked")
	}
	buttonID, ownerID, subject := m.changeButtonIDAndOwnerID(interconnection.Provider, interconnection.ExtraData)

	logrus.WithFields(logFields).Info("CreateCase")
	caseId, err := m.SalesforceService.CreatCase(ctx, contact.ID, Messages.DescriptionCase, subject, string(interconnection.Provider), ownerID,
		interconnection.ExtraData)
	if err != nil {
		span.SetTag(ext.Error, err)
		logrus.WithFields(logFields).WithError(err).Error("error CreatCase")
		go ChangeToState(interconnection.UserID, interconnection.BotSlug, TimeoutState[string(interconnection.Provider)], m.BotrunnnerClient, BotrunnerTimeout, StudioNGTimeout, m.StudioNG, m.isStudioNGFlow)
		return errors.New(helpers.ErrorMessage(titleMessage, err))
	}
	interconnection.CaseID = caseId
	logFields["caseId"] = caseId
	span.SetTag(events.Interconnection, fmt.Sprintf("%#v", interconnection))

	//Creating chat in Salesforce
	logrus.WithFields(logFields).Info("CreateChat")
	session, err := m.SalesforceService.CreatChat(ctx, interconnection.Name, SfcOrganizationID, SfcDeploymentID, buttonID, caseId, contact.ID)
	if err != nil {
		logrus.WithFields(logFields).WithError(err).Error("error CreatChat")
		span.SetTag(ext.Error, err)
		go ChangeToState(interconnection.UserID, interconnection.BotSlug, TimeoutState[string(interconnection.Provider)], m.BotrunnnerClient, BotrunnerTimeout, StudioNGTimeout, m.StudioNG, m.isStudioNGFlow)
		return errors.New(helpers.ErrorMessage(titleMessage, err))
	}

	logFields["sessionId"] = session.Id
	logrus.WithFields(logFields).Info("GetContext Routine")
	go m.getContext(interconnection)
	interconnection.AffinityToken = session.AffinityToken
	interconnection.SessionID = session.Id
	interconnection.SessionKey = session.Key
	interconnection.Status = OnHold
	interconnection.Timestamp = time.Now()

	//Add interconection to Redis and interconnectionMap
	logrus.WithFields(logFields).Info("AddInterconnection")
	span.SetTag(events.Interconnection, fmt.Sprintf("%#v", interconnection))
	m.AddInterconnection(ctx, interconnection)
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

func (m *Manager) ValidateUserID(ctx context.Context, userID string) error {
	// datadog tracing
	span, _ := tracer.StartSpanFromContext(ctx, "manager.ValidateUserID")
	span.SetTag(ext.AnalyticsEvent, true)
	span.SetTag("userID", userID)
	defer span.Finish()

	sessionRedis, _ := m.interconnectionsCache.RetrieveInterconnection(cache.Interconnection{
		UserID: userID,
		Client: m.client,
	})

	if sessionRedis != nil && (sessionRedis.Status == string(Active) || sessionRedis.Status == string(OnHold)) {
		err := errors.New("session exists in redis with this userID")
		span.SetTag(ext.Error, err)
		return err
	}

	// validate that it exists on the map
	if _, ok := m.interconnectionMap.Get(fmt.Sprintf(constants.UserKey, userID)); ok {
		err := errors.New("session exists with this userID")
		span.SetTag(ext.Error, err)
		return err
	}
	return nil
}

func (m *Manager) AddInterconnection(ctx context.Context, interconnection *Interconnection) {
	// datadog tracing
	span, _ := tracer.StartSpanFromContext(ctx, "manager.AddInterconnection")
	span.SetTag(ext.AnalyticsEvent, true)
	span.SetTag(events.UserID, interconnection.UserID)
	span.SetTag(events.Client, interconnection.Client)
	defer span.Finish()

	interconnection.SalesforceService = m.SalesforceService
	interconnection.IntegrationsClient = m.IntegrationsClient
	interconnection.BotrunnnerClient = m.BotrunnnerClient
	interconnection.finishChannel = m.finishInterconnection
	interconnection.interconnectionCache = m.interconnectionsCache
	interconnection.StudioNG = m.StudioNG
	interconnection.isStudioNGFlow = m.isStudioNGFlow
	interconnection.kafkaProducer = m.kafkaProducer
	interconnection.KafkaTopic = m.KafkaTopic

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
	interconnection.Context = fmt.Sprintf("%s:\n%s", Messages.Context, m.getContextByUserID(interconnection.UserID))
	logrus.Infof("Get context of userID : %s", interconnection.UserID)
}

// SaveContext method will save context of integration message
func (m *Manager) SaveContext(context context.Context, integration *models.IntegrationsRequest) error {
	// datadog tracing
	mainSpan, _ := tracer.StartSpanFromContext(context, "manager.SaveContext")
	mainSpan.SetTag(ext.AnalyticsEvent, true)
	mainSpan.SetTag("integrationsMessageWhatsapp", fmt.Sprintf("%#v", integration))
	mainSpan.SetTag(events.Client, m.client)
	defer mainSpan.Finish()

	logFields := logrus.Fields{
		constants.TraceIdKey: mainSpan.Context().TraceID(),
		constants.SpanIdKey:  mainSpan.Context().SpanID(),
		events.Payload:       integration,
	}

	if m.cacheMessage.IsRepeatedMessage(integration.ID) {
		mainSpan.SetTag(events.MessageRepeated, true)
		return nil
	}

	isSend := m.salesforceComunication(mainSpan, integration)
	if isSend {
		mainSpan.SetTag(events.MessageSentAgent, true)
		return nil
	}

	timestamp, err := strconv.ParseInt(integration.Timestamp, 10, 64)
	if err != nil {
		mainSpan.SetTag(ext.Error, err)
		mainSpan.SetTag(events.ContextSaved, false)
		logrus.WithFields(logFields).WithError(err).Error("Error format timestamp")
		return err
	}

	ctx := &cache.Context{
		Client:    m.client,
		UserID:    integration.From,
		Timestamp: timestamp,
		From:      fromUser,
	}

	switch {
	case integration.Type == constants.TextType:
		ctx.Text = integration.Text.Body
	case integration.Type == constants.ImageType:
		ctx.URL = integration.Image.URL
		ctx.Caption = integration.Image.Caption
		ctx.MIMEType = integration.Image.MIMEType
	case integration.Type == constants.VoiceType:
		ctx.URL = integration.Voice.URL
		ctx.Caption = integration.Voice.Caption
		ctx.MIMEType = integration.Voice.MIMEType
	case integration.Type == constants.DocumentType:
		ctx.URL = integration.Document.URL
		ctx.Caption = integration.Document.Caption
		ctx.MIMEType = integration.Document.MIMEType
	case integration.Type == constants.AudioType:
		ctx.URL = integration.Audio.URL
		ctx.MIMEType = integration.Audio.MIMEType
	default:
		return nil
	}

	if integration.To != "" {
		ctx.From = fromBot
		ctx.UserID = integration.To
	}

	mainSpan.SetTag(events.UserContext, fmt.Sprintf("%#v", ctx))
	go m.saveContextInRedis(mainSpan, ctx)
	return nil
}

func (m *Manager) salesforceComunication(mainSpan tracer.Span, integration *models.IntegrationsRequest) bool {
	interconnection, ok := m.validInterconnection(integration.From)
	isInterconnectionActive := ok && interconnection.Status == Active
	mainSpan.SetTag(events.ChatActive, isInterconnectionActive)
	if isInterconnectionActive {
		logrus.WithFields(logrus.Fields{
			constants.TraceIdKey: mainSpan.Context().TraceID(),
			constants.SpanIdKey:  mainSpan.Context().SpanID(),
			events.UserID:        interconnection.UserID,
			events.Message:       integration,
		}).Info("Send Message of user with chat")
		mainSpan.SetTag(events.UserID, interconnection.UserID)
		mainSpan.SetTag(events.Client, interconnection.Client)
		go m.sendMessageComunication(mainSpan, interconnection, integration)
	}
	return isInterconnectionActive
}

func (m *Manager) sendMessageComunication(mainSpan tracer.Span, interconnection *Interconnection, integration *models.IntegrationsRequest) {
	logFields := logrus.Fields{
		constants.TraceIdKey:   mainSpan.Context().TraceID(),
		constants.SpanIdKey:    mainSpan.Context().SpanID(),
		events.UserID:          interconnection.UserID,
		events.Interconnection: interconnection,
		"messageWhatsapp":      integration,
	}
	switch integration.Type {
	case constants.TextType:
		if strings.Contains(constants.DevEnvironments, m.environment) {
			for _, keyword := range m.keywordsRestart {
				if strings.ToLower(integration.Text.Body) == keyword {
					err := m.SalesforceService.EndChat(interconnection.AffinityToken, interconnection.SessionKey)
					mainSpan.SetTag("finishChat", true)
					if err != nil {
						mainSpan.SetTag(ext.Error, err)
						mainSpan.SetTag("finishChat", false)
						logrus.WithFields(logFields).WithError(err).Error("EndChat error")
						return
					}
					interconnection.updateStatusRedis(string(Closed))
					interconnection.Status = Closed
					interconnection.runnigLongPolling = false
					return
				}
			}
		}

		interconnection.sendMessageToQueue(mainSpan,
			integration.ID,
			integration.Text.Body,
			constants.SendMessageToSalesforce)

	case constants.ImageType:
		imageName := defineImageName(interconnection, integration)

		err := m.SalesforceService.InsertImageInCase(
			integration.Image.URL,
			imageName,
			integration.Image.MIMEType,
			interconnection.CaseID)
		if err != nil {
			mainSpan.SetTag(ext.Error, err)
			mainSpan.SetTag(events.SendImage, false)
			logrus.WithFields(logFields).WithError(err).Error("InsertImageInCase error")
			interconnection.sendMessageToQueue(mainSpan,
				integration.ID,
				Messages.UploadImageError,
				constants.SendMessageToUser)

			return
		}
		logrus.WithFields(logFields).Info("Send Image to agent")
		mainSpan.SetTag(events.SendImage, true)

		textMessage := Messages.UploadImageSuccess
		if SendImageNameInMessage {
			textMessage += imageName
		}

		interconnection.sendMessageToQueue(mainSpan,
			integration.ID,
			textMessage,
			constants.SendMessageToSalesforce)
	}
}

func defineImageName(interconnection *Interconnection, integration *models.IntegrationsRequest) string {
	maxLength := 255
	if integration.Image.Caption != "" && len(integration.Image.Caption) <= maxLength {
		return integration.Image.Caption
	}

	regexToFindImageId := regexp.MustCompile("\\/.+\\/(.+-.+-.+-.+)")
	imageId := regexToFindImageId.FindStringSubmatch(integration.Image.URL)
	if len(imageId) > 0 && imageId[1] != "" {
		return imageId[1]
	}

	return interconnection.SessionID
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

		loc, _ := time.LoadLocation(Timezone)

		date := time.Unix(0, int64(ctx.Timestamp)*int64(time.Millisecond)).
			In(loc).
			Format(constants.DateFormat)

		ctx.Text = strings.TrimRight(ctx.Text, "\n")
		if ctx.From == fromUser {
			fmt.Fprintf(&builder, "%s [%s]:%s\n\n", Messages.ClientLabel, date, ctx.Text)
		} else {
			fmt.Fprintf(&builder, "%s [%s]:%s\n\n", Messages.BotLabel, date, ctx.Text)
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
func (m *Manager) SaveContextFB(context context.Context, integration *models.IntegrationsFacebook) error {
	// datadog tracing
	mainSpan, _ := tracer.StartSpanFromContext(context, "manager.SaveContextFB")
	mainSpan.SetTag(ext.AnalyticsEvent, true)
	mainSpan.SetTag("integrationsMessageFacebook", fmt.Sprintf("%#v", integration))
	mainSpan.SetTag(events.Client, m.client)
	defer mainSpan.Finish()

	isSend := false
	for _, entry := range integration.Message.Entry {
		for _, message := range entry.Messaging {
			userID := message.Recipient.ID
			from := fromBot
			if m.cacheMessage.IsRepeatedMessage(message.Message.Mid) {
				mainSpan.SetTag(events.MessageRepeated, true)
				continue
			}
			if integration.AuthorRole == fromUser {
				isSend = m.salesforceComunicationFB(mainSpan, message)
				userID = message.Sender.ID
				from = fromUser
			}

			if isSend {
				mainSpan.SetTag(events.MessageSentAgent, true)
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

			mainSpan.SetTag(events.UserContext, fmt.Sprintf("%#v", ctx))
			go m.saveContextInRedis(mainSpan, ctx)
		}
	}
	return nil
}

func (m *Manager) salesforceComunicationFB(mainSpan tracer.Span, message models.Messaging) bool {
	mainSpan.SetTag(events.Message, fmt.Sprintf("%#v", message))
	interconnection, ok := m.validInterconnection(message.Sender.ID)
	isInterconnectionActive := ok && interconnection.Status == Active
	mainSpan.SetTag(events.ChatActive, isInterconnectionActive)
	if isInterconnectionActive {
		logrus.WithFields(logrus.Fields{
			constants.TraceIdKey: mainSpan.Context().TraceID(),
			constants.SpanIdKey:  mainSpan.Context().SpanID(),
			events.UserID:        interconnection.UserID,
			events.Message:       message,
		}).Info("Send Message of user with chat FB")
		mainSpan.SetTag(events.UserID, interconnection.UserID)
		mainSpan.SetTag(events.Client, interconnection.Client)
		go m.sendMessageComunicationFB(mainSpan, interconnection, &message)
	}

	return isInterconnectionActive
}

func (m *Manager) sendMessageComunicationFB(mainSpan tracer.Span, interconnection *Interconnection, message *models.Messaging) {
	logFields := logrus.Fields{
		constants.TraceIdKey:   mainSpan.Context().TraceID(),
		constants.SpanIdKey:    mainSpan.Context().SpanID(),
		events.UserID:          interconnection.UserID,
		events.Interconnection: interconnection,
		"messageFacebook":      message,
	}
	switch {
	case message.Message.Text != "":
		if strings.Contains(constants.DevEnvironments, m.environment) {
			for _, keyword := range m.keywordsRestart {
				if strings.ToLower(message.Message.Text) == keyword {
					err := m.SalesforceService.EndChat(interconnection.AffinityToken, interconnection.SessionKey)
					mainSpan.SetTag("finishChat", true)
					if err != nil {
						mainSpan.SetTag(ext.Error, err)
						mainSpan.SetTag("finishChat", false)
						logrus.WithFields(logFields).WithError(err).Error("End Chat for restart key error")
						return
					}
					interconnection.updateStatusRedis(string(Closed))
					interconnection.Status = Closed
					interconnection.runnigLongPolling = false
					return
				}
			}
		}
		interconnection.sendMessageToQueue(mainSpan,
			message.Sender.ID,
			message.Message.Text,
			constants.SendMessageToSalesforce)

	case message.Message.Attachments != nil:
		for _, attachment := range message.Message.Attachments {
			if attachment.Type == constants.ImageType {
				err := m.SalesforceService.InsertImageInCase(
					attachment.Payload.URL,
					interconnection.SessionID,
					"",
					interconnection.CaseID)
				if err != nil {
					mainSpan.SetTag(ext.Error, err)
					mainSpan.SetTag(events.SendImage, false)
					logrus.WithFields(logFields).WithError(err).Error("InsertImageInCase error")
					interconnection.sendMessageToQueue(mainSpan,
						message.Sender.ID,
						Messages.UploadImageError,
						constants.SendMessageToUser)
					return
				}
				logrus.WithFields(logFields).Info("FB Send Image to agent")
				mainSpan.SetTag(events.SendImage, true)
				interconnection.sendMessageToQueue(mainSpan,
					message.Sender.ID,
					Messages.UploadImageSuccess,
					constants.SendMessageToSalesforce)
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

func (m *Manager) saveContextInRedis(mainSpan tracer.Span, ctx *cache.Context) {
	// datadog tracing
	spanContext := events.GetSpanContextFromSpan(mainSpan)
	span := tracer.StartSpan(fmt.Sprintf("%s.saveContextInRedis", events.UserID), tracer.ChildOf(spanContext))
	span.SetTag(ext.AnalyticsEvent, true)
	span.SetTag(events.UserID, ctx.UserID)
	span.SetTag(events.Client, ctx.Client)
	mainSpan.SetTag(events.ContextSaved, true)
	defer span.Finish()

	ctx.Ttl = time.Now().Add(cache.Ttl)
	mainSpan.SetTag(events.UserContext, fmt.Sprintf("%#v", ctx))
	err := m.contextcache.StoreContextToSet(*ctx)
	if err != nil {
		mainSpan.SetTag(events.ContextSaved, false)
		span.SetTag(ext.Error, err)
		logrus.WithFields(logrus.Fields{
			constants.TraceIdKey: span.Context().TraceID(),
			constants.SpanIdKey:  span.Context().SpanID(),
			events.Context:       ctx,
		}).WithError(err).Error("Error store context in set")
	}
}

func (m *Manager) Process(ctx context.Context, msg []byte) error {
	readSpan, _ := tracer.StartSpanFromContext(ctx, "read_kafka")
	readSpan.SetTag(ext.AnalyticsEvent, true)
	defer readSpan.Finish()

	message := InterconnectionMessageQueue{}
	if err := json.Unmarshal(msg, &message); err != nil {
		readSpan.SetTag(ext.Error, err)
		errorMessage := fmt.Sprintf("could not marshal message: %s", err.Error())
		logrus.Error(errorMessage)
		return errors.New(errorMessage)
	}
	spanContext := events.GetSpanContextFromtraceId(message.TraceID)
	span := tracer.StartSpan("process_kafka", tracer.ChildOf(spanContext))
	span.SetTag(ext.AnalyticsEvent, true)
	defer span.Finish()

	span.SetTag(events.Interconnection, fmt.Sprintf("%#v", message.Params))
	span.SetTag(events.Topic, m.KafkaTopic)
	span.SetTag(events.EventType, message.EventType)
	span.SetTag(events.UserID, message.Params.UserID)
	span.SetTag(events.Client, message.Params.Client)

	switch message.EventType {
	case constants.SendMessageToSalesforce:
		m.SalesforceChanRequestLimiter.Wait(ctx)

		go m.sendMessageToSalesforce(NewSfMessage(span,
			message.Params.AffinityToken,
			message.Params.SessionKey,
			message.Params.Text,
			message.Params.UserID))

	case constants.SendMessageToUser:
		m.IntegrationChanRateLimiter.Wait(ctx)

		go m.sendMessageToUser(NewIntegrationsMessage(span,
			message.ID,
			message.Params.UserID,
			message.Params.Text,
			message.Params.Provider))
	}
	return nil
}
