package integrations

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"yalochat.com/salesforce-integration/base/clients/proxy"
	"yalochat.com/salesforce-integration/base/constants"
	"yalochat.com/salesforce-integration/base/helpers"
)

func NewIntegrationsClient(url, token, channel, botId string) *IntegrationsClient {
	logrus.WithFields(logrus.Fields{
		"url": url,
	}).Info("Proxy setted")

	return &IntegrationsClient{
		Channel:     channel,
		BotId:       botId,
		Proxy:       proxy.NewProxy(url),
		AccessToken: token,
	}
}

type IntegrationsClient struct {
	Channel     string
	BotId       string
	AccessToken string
	Proxy       proxy.ProxyInterface
}

type HealthcheckResponse struct {
	BotId   string `json:"bot_id"`
	Channel string `json:"channel"`
	Webhook string `json:"webhook"`
}

type HealthcheckPayload struct {
	Phone   string `json:"phone" validate:"required"`
	Webhook string `json:"webhook" validate:"required"`
}

type RemoveWebhookPayload struct {
	Phone string `json:"phone" validate:"required"`
}

type TextMessage struct {
	Body string `json:"body"  validate:"required"`
}

type SendTextPayload struct {
	Id     string      `json:"id"`
	Type   string      `json:"type" validate:"required`
	UserId string      `json:"userId" validate:"required`
	Text   TextMessage `json:"text" validate:"required"`
}

type SendImagePayload struct {
	Id     string `json:"id"`
	Type   string `json:"type" validate:"required`
	UserId string `json:"userId" validate:"required`
	Image  Media  `json:"image" validate:"required"`
}

type SendVideoPayload struct {
	Id     string `json:"id"`
	Type   string `json:"type" validate:"required`
	UserId string `json:"userId" validate:"required`
	Video  Media  `json:"video" validate:"required"`
}

type SendDocumentPayload struct {
	Id       string `json:"id"`
	Type     string `json:"type" validate:"required`
	UserId   string `json:"userId" validate:"required`
	Document Media  `json:"document" validate:"required"`
}

type SendAudioPayload struct {
	Id     string `json:"id"`
	Type   string `json:"type" validate:"required`
	UserId string `json:"userId" validate:"required`
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

// Register Webhook listenner chat integration (healthcheck)
func (cc *IntegrationsClient) WebhookRegister(HealthcheckPayload HealthcheckPayload) (*HealthcheckResponse, error) {
	var errorMessage string

	logrus.WithFields(logrus.Fields{
		"payload": HealthcheckPayload,
	}).Info("HealthcheckPayload received")

	//validating ContentVersionPayload struct
	if err := helpers.Govalidator().Struct(HealthcheckPayload); err != nil {
		errorMessage = fmt.Sprintf("%s : %s", helpers.InvalidPayload, err.Error())
		logrus.Error(errorMessage)
		return nil, errors.New(errorMessage)
	}

	//building request to send through proxy
	requestBytes, _ := json.Marshal(HealthcheckPayload)

	header := make(map[string]string)
	header["Content-Type"] = "application/json"
	header["Authorization"] = fmt.Sprintf("Bearer %s", cc.AccessToken)

	newRequest := proxy.Request{
		Body:      requestBytes,
		Method:    http.MethodPost,
		URI:       fmt.Sprintf("/api/%s/bots/%s/healthcheck", cc.Channel, cc.BotId),
		HeaderMap: header,
	}

	proxiedResponse, proxyError := cc.Proxy.SendHTTPRequest(&newRequest)
	if proxyError != nil {
		errorMessage = fmt.Sprintf("%s : %s", constants.ForwardError, proxyError.Error())
		logrus.Error(errorMessage)
		return nil, errors.New(errorMessage)
	}

	if proxiedResponse.StatusCode != http.StatusOK {
		return nil, helpers.ErrorResponseMap(proxiedResponse.Body, constants.StatusError, proxiedResponse.StatusCode)
	}

	var response HealthcheckResponse
	readAndUnmarshalError := helpers.ReadAndUnmarshal(proxiedResponse.Body, &response)

	if readAndUnmarshalError != nil {
		errorMessage = fmt.Sprintf("%s : %s", constants.UnmarshallError, readAndUnmarshalError.Error())
		logrus.Error(errorMessage)
		return nil, errors.New(errorMessage)
	}

	logrus.WithFields(logrus.Fields{
		"response": response,
	}).Info("Healthcheck successfully")

	return &response, nil
}

// Remove webhook from bot
func (cc *IntegrationsClient) WebhookRemove(removeWebhookPayload RemoveWebhookPayload) (bool, error) {
	var errorMessage string

	logrus.WithFields(logrus.Fields{
		"payload": removeWebhookPayload,
	}).Info("Remove Webhook payload received")

	//validating ContentVersionPayload struct
	if err := helpers.Govalidator().Struct(removeWebhookPayload); err != nil {
		errorMessage = fmt.Sprintf("%s : %s", helpers.InvalidPayload, err.Error())
		logrus.Error(errorMessage)
		return false, errors.New(errorMessage)
	}

	//building request to send through proxy
	requestBytes, _ := json.Marshal(removeWebhookPayload)

	header := make(map[string]string)
	header["Content-Type"] = "application/json"
	header["Authorization"] = fmt.Sprintf("Bearer %s", cc.AccessToken)

	newRequest := proxy.Request{
		Body:      requestBytes,
		Method:    http.MethodPost,
		URI:       fmt.Sprintf("/api/%s/bots/%s/remove", cc.Channel, cc.BotId),
		HeaderMap: header,
	}

	proxiedResponse, proxyError := cc.Proxy.SendHTTPRequest(&newRequest)
	if proxyError != nil {
		errorMessage = fmt.Sprintf("%s : %s", constants.ForwardError, proxyError.Error())
		logrus.Error(errorMessage)
		return false, errors.New(errorMessage)
	}

	if proxiedResponse.StatusCode != http.StatusNoContent {
		return false, helpers.ErrorResponseMap(proxiedResponse.Body, constants.StatusError, proxiedResponse.StatusCode)
	}

	logrus.WithFields(logrus.Fields{
		"response": "Ok",
	}).Info("Webhook Remove successfully")

	return true, nil
}

// Send message to bot (Text, Image, Audio, Video, Document)
func (cc *IntegrationsClient) SendMessage(messagePayload interface{}) (*SendMessageResponse, error) {
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
		return nil, errors.New(errorMessage)
	}

	//building request to send through proxy
	requestBytes, _ := json.Marshal(messagePayload)
	header := make(map[string]string)
	header["Content-Type"] = "application/json"
	header["Authorization"] = fmt.Sprintf("Bearer %s", cc.AccessToken)

	newRequest := proxy.Request{
		Body:      requestBytes,
		Method:    http.MethodPost,
		URI:       fmt.Sprintf("/api/%s/bots/%s/messages", cc.Channel, cc.BotId),
		HeaderMap: header,
	}

	proxiedResponse, proxyError := cc.Proxy.SendHTTPRequest(&newRequest)
	if proxyError != nil {
		errorMessage = fmt.Sprintf("%s : %s", constants.ForwardError, proxyError.Error())
		logrus.Error(errorMessage)
		return nil, errors.New(errorMessage)
	}

	if proxiedResponse.StatusCode != http.StatusCreated {
		return nil, helpers.ErrorResponseMap(proxiedResponse.Body, constants.StatusError, proxiedResponse.StatusCode)
	}

	var response SendMessageResponse
	readAndUnmarshalError := helpers.ReadAndUnmarshal(proxiedResponse.Body, &response)

	if readAndUnmarshalError != nil {
		errorMessage = fmt.Sprintf("%s : %s", constants.UnmarshallError, readAndUnmarshalError.Error())
		logrus.Error(errorMessage)
		return nil, errors.New(errorMessage)
	}

	logrus.WithFields(logrus.Fields{
		"response": response,
	}).Info("Send Message successfully")

	return &response, nil
}
