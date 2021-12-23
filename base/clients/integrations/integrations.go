package integrations

import (
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"net/http"
	"yalochat.com/salesforce-integration/base/events"

	"github.com/sirupsen/logrus"
	"yalochat.com/salesforce-integration/base/clients/proxy"
	"yalochat.com/salesforce-integration/base/constants"
	"yalochat.com/salesforce-integration/base/helpers"
)

func NewIntegrationsClient(url, tokenWA, tokenFB, channelWA, channelFB, botWAID, botFBID string) *IntegrationsClient {
	logrus.WithFields(logrus.Fields{
		"url": url,
	}).Info("Proxy setted")

	return &IntegrationsClient{
		ChannelWA:     channelWA,
		ChannelFB:     channelFB,
		BotWAID:       botWAID,
		BotFBID:       botFBID,
		Proxy:         proxy.NewProxy(url, 30),
		AccessTokenWA: tokenWA,
		AccessTokenFB: tokenFB,
	}
}

type IntegrationsClient struct {
	ChannelWA     string
	ChannelFB     string
	BotWAID       string
	BotFBID       string
	AccessTokenWA string
	AccessTokenFB string

	Proxy proxy.ProxyInterface
}

type IntegrationInterface interface {
	WebhookRegister(HealthcheckPayload HealthcheckPayload) (*HealthcheckResponse, error)
	WebhookRemove(removeWebhookPayload RemoveWebhookPayload) (bool, error)
	SendMessage(messagePayload interface{}, provider string) (*SendMessageResponse, error)
}

type HealthcheckResponse struct {
	BotId   string `json:"bot_id"`
	Channel string `json:"channel"`
	Webhook string `json:"webhook"`
}

type HealthcheckPayload struct {
	Phone    string `json:"phone" validate:"required"`
	Webhook  string `json:"webhook" validate:"required"`
	Provider string `json:"provider" validate:"required"`
	Version  int    `json:"version"`
}

type RemoveWebhookPayload struct {
	Phone    string `json:"phone" validate:"required"`
	Provider string `json:"provider" validate:"required"`
}

type TextMessage struct {
	Body string `json:"body"  validate:"required"`
}

type SendTextPayload struct {
	Id     string      `json:"id"`
	Type   string      `json:"type" validate:"required"`
	UserID string      `json:"userId" validate:"required"`
	Text   TextMessage `json:"text" validate:"required"`
}

type SendTextPayloadFB struct {
	MessagingType string    `json:"messaging_type"`
	Recipient     Recipient `json:"recipient" validate:"required"`
	Message       Message   `json:"message" validate:"required"`
	Metadata      string    `json:"metadata" validate:"required"`
}

type Recipient struct {
	ID string `json:"id"`
}
type Message struct {
	Text string `json:"text"`
}

type SendImagePayload struct {
	ID     string `json:"id"`
	Type   string `json:"type" validate:"required"`
	UserID string `json:"userId" validate:"required"`
	Image  Media  `json:"image" validate:"required"`
}

type SendVideoPayload struct {
	Id     string `json:"id"`
	Type   string `json:"type" validate:"required"`
	UserID string `json:"userId" validate:"required"`
	Video  Media  `json:"video" validate:"required"`
}

type SendDocumentPayload struct {
	Id       string `json:"id"`
	Type     string `json:"type" validate:"required"`
	UserID   string `json:"userId" validate:"required"`
	Document Media  `json:"document" validate:"required"`
}

type SendAudioPayload struct {
	Id     string `json:"id"`
	Type   string `json:"type" validate:"required"`
	UserID string `json:"userId" validate:"required"`
	Audio  Media  `json:"audio" validate:"required"`
}

type MessageId struct {
	Id string `json:"id"`
}

type SendMessageResponse struct {
	Messages []MessageId `json:"messages"`
}

type Media struct {
	Url     string `json:"url" validate:"required"`
	Caption string `json:"caption"`
}

// WebhookRegister Register Webhook listenner chat integration (healthcheck)
func (cc *IntegrationsClient) WebhookRegister(HealthcheckPayload HealthcheckPayload) (*HealthcheckResponse, error) {
	// datadog tracing
	span := tracer.StartSpan("register_webhook")
	span.SetTag(ext.AnalyticsEvent, true)
	span.SetTag(events.Payload, fmt.Sprintf("%#v", HealthcheckPayload))
	defer span.Finish()
	botID, channel, token := cc.getDataFromProvider(HealthcheckPayload.Provider)
	uri := fmt.Sprintf("/api/%s/bots/%s/healthcheck", channel, botID)
	span.SetTag(ext.ResourceName, fmt.Sprintf("%s %s", http.MethodPost, uri))

	var errorMessage string

	logrus.WithFields(logrus.Fields{
		"payload": HealthcheckPayload,
	}).Info("HealthcheckPayload received")

	//validating ContentVersionPayload struct
	if err := helpers.Govalidator().Struct(HealthcheckPayload); err != nil {
		errorMessage = fmt.Sprintf("%s : %s", helpers.InvalidPayload, err.Error())
		logrus.Error(errorMessage)
		span.SetTag(ext.Error, err)
		return nil, errors.New(errorMessage)
	}

	//building request to send through proxy
	requestBytes, _ := json.Marshal(HealthcheckPayload)

	header := make(map[string]string)
	header["Content-Type"] = "application/json"
	header["Authorization"] = fmt.Sprintf("Bearer %s", token)

	newRequest := proxy.Request{
		Body:      requestBytes,
		Method:    http.MethodPost,
		URI:       uri,
		HeaderMap: header,
	}

	proxiedResponse, proxyError := cc.Proxy.SendHTTPRequest(span, &newRequest)
	if proxyError != nil {
		errorMessage = fmt.Sprintf("%s : %s", constants.ForwardError, proxyError.Error())
		logrus.Error(errorMessage)
		span.SetTag(ext.Error, proxyError)
		return nil, errors.New(errorMessage)
	}

	if proxiedResponse.StatusCode != http.StatusCreated {
		err := helpers.ErrorResponseMap(proxiedResponse.Body, constants.StatusError, proxiedResponse.StatusCode)
		span.SetTag(ext.Error, err)
		return nil, err
	}

	var response HealthcheckResponse
	readAndUnmarshalError := helpers.ReadAndUnmarshal(proxiedResponse.Body, &response)

	if readAndUnmarshalError != nil {
		errorMessage = fmt.Sprintf("%s : %s", constants.UnmarshallError, readAndUnmarshalError.Error())
		logrus.Error(errorMessage)
		span.SetTag(ext.Error, readAndUnmarshalError)
		return nil, errors.New(errorMessage)
	}

	logrus.WithFields(logrus.Fields{
		"response": response,
	}).Info("Healthcheck successfully")

	return &response, nil
}

func (cc *IntegrationsClient) getDataFromProvider(provider string) (string, string, string) {
	botID := cc.BotWAID
	channel := cc.ChannelWA
	token := cc.AccessTokenWA
	if provider == constants.FacebookProvider {
		botID = cc.BotFBID
		channel = cc.ChannelFB
		token = cc.AccessTokenFB
	}
	return botID, channel, token
}

// WebhookRemove Remove webhook from bot
func (cc *IntegrationsClient) WebhookRemove(removeWebhookPayload RemoveWebhookPayload) (bool, error) {
	// datadog tracing
	span := tracer.StartSpan("remove_webhook")
	span.SetTag(ext.AnalyticsEvent, true)
	span.SetTag(events.Payload, fmt.Sprintf("%#v", removeWebhookPayload))
	defer span.Finish()
	botID, channel, token := cc.getDataFromProvider(removeWebhookPayload.Provider)
	uri := fmt.Sprintf("/api/%s/bots/%s/remove", channel, botID)
	span.SetTag(ext.ResourceName, fmt.Sprintf("%s %s", http.MethodPost, uri))

	var errorMessage string

	logrus.WithFields(logrus.Fields{
		"payload": removeWebhookPayload,
	}).Info("Remove Webhook payload received")

	//validating ContentVersionPayload struct
	if err := helpers.Govalidator().Struct(removeWebhookPayload); err != nil {
		errorMessage = fmt.Sprintf("%s : %s", helpers.InvalidPayload, err.Error())
		logrus.Error(errorMessage)
		span.SetTag(ext.Error, err)
		return false, errors.New(errorMessage)
	}

	//building request to send through proxy
	requestBytes, _ := json.Marshal(removeWebhookPayload)

	header := make(map[string]string)
	header["Content-Type"] = "application/json"
	header["Authorization"] = fmt.Sprintf("Bearer %s", token)

	newRequest := proxy.Request{
		Body:      requestBytes,
		Method:    http.MethodPost,
		URI:       uri,
		HeaderMap: header,
	}

	proxiedResponse, proxyError := cc.Proxy.SendHTTPRequest(span, &newRequest)
	if proxyError != nil {
		errorMessage = fmt.Sprintf("%s : %s", constants.ForwardError, proxyError.Error())
		logrus.Error(errorMessage)
		span.SetTag(ext.Error, proxyError)
		return false, errors.New(errorMessage)
	}

	if proxiedResponse.StatusCode != http.StatusNoContent {
		err := helpers.ErrorResponseMap(proxiedResponse.Body, constants.StatusError, proxiedResponse.StatusCode)
		span.SetTag(ext.Error, err)
		return false, err
	}

	logrus.WithFields(logrus.Fields{
		"response": "Ok",
	}).Info("Webhook Remove successfully")

	return true, nil
}

// SendMessage Send message to bot (Text, Image, Audio, Video, Document)
func (cc *IntegrationsClient) SendMessage(messagePayload interface{}, provider string) (*SendMessageResponse, error) {
	// datadog tracing
	span := tracer.StartSpan("send_message")
	span.SetTag(ext.AnalyticsEvent, true)
	span.SetTag(events.Payload, fmt.Sprintf("%#v", messagePayload))
	defer span.Finish()
	botID, channel, token := cc.getDataFromProvider(provider)
	uri := fmt.Sprintf("/api/%s/bots/%s/messages", channel, botID)
	span.SetTag(ext.ResourceName, fmt.Sprintf("%s %s", http.MethodPost, uri))

	var errorMessage string

	// If not have Id set a random id
	/*fieldId := reflect.ValueOf(messagePayload).Elem().FieldByName("id")
	if fieldId.IsValid() && fieldId.String() == "" {
		id := helpers.RandomString(24)
		fieldId.SetString(id)
	}*/

	logrus.WithFields(logrus.Fields{
		"payload": messagePayload,
	}).Info("SendMessage payload received")

	//validating ContentVersionPayload struct
	if err := helpers.Govalidator().Struct(messagePayload); err != nil {
		errorMessage = fmt.Sprintf("%s : %s", helpers.InvalidPayload, err.Error())
		logrus.Error(errorMessage)
		span.SetTag(ext.Error, err)
		return nil, errors.New(errorMessage)
	}

	//building request to send through proxy
	requestBytes, _ := json.Marshal(messagePayload)
	header := make(map[string]string)
	header["Content-Type"] = "application/json"
	header["Authorization"] = fmt.Sprintf("Bearer %s", token)

	newRequest := proxy.Request{
		Body:      requestBytes,
		Method:    http.MethodPost,
		URI:       uri,
		HeaderMap: header,
	}

	proxiedResponse, proxyError := cc.Proxy.SendHTTPRequest(span, &newRequest)
	if proxyError != nil {
		errorMessage = fmt.Sprintf("%s : %s", constants.ForwardError, proxyError.Error())
		logrus.Error(errorMessage)
		span.SetTag(ext.Error, proxyError)
		return nil, errors.New(errorMessage)
	}

	if proxiedResponse.StatusCode != http.StatusCreated && proxiedResponse.StatusCode != http.StatusOK {
		err := helpers.ErrorResponseMap(proxiedResponse.Body, constants.StatusError, proxiedResponse.StatusCode)
		span.SetTag(ext.Error, err)
		return nil, err
	}

	var response SendMessageResponse
	readAndUnmarshalError := helpers.ReadAndUnmarshal(proxiedResponse.Body, &response)

	if readAndUnmarshalError != nil {
		errorMessage = fmt.Sprintf("%s : %s", constants.UnmarshallError, readAndUnmarshalError.Error())
		logrus.Error(errorMessage)
		span.SetTag(ext.Error, readAndUnmarshalError)
		return nil, errors.New(errorMessage)
	}

	logrus.WithFields(logrus.Fields{
		"response": response,
	}).Info("Send Message successfully")

	return &response, nil
}
