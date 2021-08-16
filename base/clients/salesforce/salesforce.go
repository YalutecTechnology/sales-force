package salesforce

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"

	"github.com/sirupsen/logrus"
	"yalochat.com/salesforce-integration/base/clients/proxy"
	"yalochat.com/salesforce-integration/base/helpers"
)

const (
	post             = "POST"
	get              = "GET"
	granType         = "password"
	apiVersionHeader = "X-LIVEAGENT-API-VERSION"
	affinityHeader   = "X-LIVEAGENT-AFFINITY"
	sessionKeyHeader = "X-LIVEAGENT-SESSION-KEY"
	sequenceHeader   = "X-LIVEAGENT-SEQUENCE"
	sequenceValue    = "1"
	forwardError     = "Error forwarding the request through the Proxy"
	unmarshallError  = "Error unmarshalling the response from salesForce"
	statusError      = "Error call with status"
)

type SalesforceClient struct {
	LoginURL      string
	SalesforceURL string
	CaseURL       string
	ApiVersion    int
	accessToken   string
	Proxy         proxy.ProxyInterface
}

type TokenPayload struct {
	GrantType    string `json:"grant_type" validate:"required"`
	ClientId     string `json:"client_id" validate:"required"`
	ClientSecret string `json:"client_secret" validate:"required"`
	Username     string `json:"username" validate:"required"`
	Password     string `json:"password" validate:"required"`
}

type SessionResponse struct {
	ClientPollTimeout int    `json:"clientPollTimeout"`
	Key               string `json:"key"`
	AffinityToken     string `json:"affinityToken"`
	Id                string `json:"id"`
}

type ChatRequest struct {
	OrganizationId      string   `json:"organizationId" validate:"required"`
	DeploymentId        string   `json:"deploymentId" validate:"required"`
	ButtonId            string   `json:"buttonId" validate:"required"`
	SessionId           string   `json:"sessionId" validate:"required"`
	UserAgent           string   `json:"userAgent" validate:"required"`
	Language            string   `json:"language" validate:"required"`
	ScreenResolution    string   `json:"screenResolution" validate:"required"`
	VisitorName         string   `json:"visitorName" validate:"required"`
	PrechatDetails      []string `json:"prechatDetails" validate:"required"`
	PrechatEntities     []string `json:"prechatEntities" validate:"required"`
	ReceiveQueueUpdates bool     `json:"receiveQueueUpdates" validate:"required"`
	IsPost              bool     `json:"isPost" validate:"required"`
}

type MessagePayload struct {
	Text string `json:"text" validate:"required"`
}

type Message struct{}

type SaleforceInterface interface {
	GetToken(TokenPayload) (bool, error)
	CreateSession() (*SessionResponse, error)
	CreateChat(string, string, ChatRequest) (bool, error)
	CreateCase() error
	GetMessages(string, string) ([]Message, error)
	CreateMediaObject()
	SendMessage(string, string, MessagePayload) (bool, error)
	SendMessageWithMedia()
}

// Get Access Token for Salesforce Requests
func (c *SalesforceClient) GetToken(tokenPayload TokenPayload) (bool, error) {
	var errorMessage string
	if tokenPayload.GrantType == "" {
		tokenPayload.GrantType = granType
	}

	// This log should hide the secrets before sending to production
	logrus.WithFields(logrus.Fields{
		"payload": tokenPayload,
	}).Info("Payload received")

	//validating token Payload struct
	if err := helpers.Govalidator().Struct(tokenPayload); err != nil {
		errorMessage = fmt.Sprintf("%s : %s", helpers.InvalidPayload, err.Error())
		logrus.Error(errorMessage)
		return false, errors.New(errorMessage)
	}

	//building request to send through proxy
	dataEncode := url.Values{}
	dataEncode.Set("grant_type", tokenPayload.GrantType)
	dataEncode.Set("client_id", tokenPayload.ClientId)
	dataEncode.Set("client_secret", tokenPayload.ClientSecret)
	dataEncode.Set("username", tokenPayload.Username)
	dataEncode.Set("password", tokenPayload.Password)

	header := make(map[string]string)
	header["Content-Type"] = proxy.FormUrlencodeHeader

	c.Proxy.SetBaseURL(c.LoginURL)
	newRequest := proxy.Request{
		DataEncode: dataEncode,
		Method:     post,
		URI:        "/services/oauth2/token",
		HeaderMap:  header,
	}

	proxiedResponse, proxyError := c.Proxy.SendHTTPRequest(&newRequest)
	if proxyError != nil {
		errorMessage = fmt.Sprintf("%s : %s", forwardError, proxyError.Error())
		logrus.Error(errorMessage)
		return false, errors.New(errorMessage)
	}

	responseMap := map[string]interface{}{}
	readAndUnmarshalError := helpers.ReadAndUnmarshal(proxiedResponse.Body, &responseMap)

	if readAndUnmarshalError != nil {
		errorMessage = fmt.Sprintf("%s : %s", unmarshallError, readAndUnmarshalError.Error())
		logrus.Error(errorMessage)
		return false, errors.New(errorMessage)
	}

	if proxiedResponse.StatusCode != 200 {
		errorMessage = fmt.Sprintf("%s : %d", statusError, proxiedResponse.StatusCode)
		logrus.WithFields(logrus.Fields{
			"response": responseMap,
		}).Error(errorMessage)
		return false, errors.New(errorMessage)
	}

	//check this one if this is a response success
	logrus.WithFields(logrus.Fields{
		"response": responseMap,
	}).Info("Get accessToken sucessfully")

	if _, ok := responseMap["access_token"]; ok {
		c.accessToken = responseMap["access_token"].(string)
	}

	return true, nil
}

//To create a new Live Agent session, you must call the SessionId request.
//SessionId : Establishes a new Live Agent session. The SessionId request is required as the first request to create every new Live Agent session.
func (c *SalesforceClient) CreateSession() (*SessionResponse, error) {
	var errorMessage string

	//building request to send through proxy
	header := make(map[string]string)
	header[apiVersionHeader] = strconv.Itoa(c.ApiVersion)
	header[affinityHeader] = "null"

	c.Proxy.SetBaseURL(c.SalesforceURL)
	newRequest := proxy.Request{
		Method:    get,
		URI:       "/chat/rest/System/SessionId",
		HeaderMap: header,
	}

	proxiedResponse, proxyError := c.Proxy.SendHTTPRequest(&newRequest)
	if proxyError != nil {
		errorMessage = fmt.Sprintf("%s : %s", forwardError, proxyError.Error())
		logrus.Error(errorMessage)
		return nil, errors.New(errorMessage)
	}

	if proxiedResponse.StatusCode != 200 {
		responseMap := map[string]interface{}{}
		readAndUnmarshalError := helpers.ReadAndUnmarshal(proxiedResponse.Body, &responseMap)

		if readAndUnmarshalError != nil {
			errorMessage = fmt.Sprintf("%s : %s", unmarshallError, readAndUnmarshalError.Error())
			logrus.Error(errorMessage)
			return nil, errors.New(errorMessage)
		}
		errorMessage = fmt.Sprintf("%s : %d", statusError, proxiedResponse.StatusCode)
		logrus.WithFields(logrus.Fields{
			"response": responseMap,
		}).Error(errorMessage)
		return nil, errors.New(errorMessage)
	}

	var session SessionResponse
	readAndUnmarshalError := helpers.ReadAndUnmarshal(proxiedResponse.Body, &session)

	if readAndUnmarshalError != nil {
		errorMessage = fmt.Sprintf("%s : %s", unmarshallError, readAndUnmarshalError.Error())
		logrus.Error(errorMessage)
		return nil, errors.New(errorMessage)
	}

	//check this one if this is a response success
	logrus.WithFields(logrus.Fields{
		"response": session,
	}).Info("Get Session sucessfully")
	return &session, nil
}

// We’ll send a chat request. To make sure this works, you should log in as a Live Agent user and make yourself available.
func (c *SalesforceClient) CreateChat(affinityToken, sessionKey string, request ChatRequest) (bool, error) {
	var errorMessage string
	// This log should hide the secrets before sending to production
	logrus.WithFields(logrus.Fields{
		"payload": request,
	}).Info("Payload received")

	//validating token Payload struct
	if err := helpers.Govalidator().Struct(request); err != nil {
		errorMessage = fmt.Sprintf("%s : %s", helpers.InvalidPayload, err.Error())
		logrus.Error(errorMessage)
		return false, errors.New(errorMessage)
	}

	//building request to send through proxy
	requestBytes, _ := json.Marshal(request)

	header := make(map[string]string)
	header[apiVersionHeader] = strconv.Itoa(c.ApiVersion)
	header[affinityHeader] = affinityToken
	header[sessionKeyHeader] = sessionKey
	header[sequenceHeader] = sequenceValue

	c.Proxy.SetBaseURL(c.SalesforceURL)
	newRequest := proxy.Request{
		Body:      requestBytes,
		Method:    post,
		URI:       "/chat/rest/Chasitor/ChasitorInit",
		HeaderMap: header,
	}

	proxiedResponse, proxyError := c.Proxy.SendHTTPRequest(&newRequest)
	if proxyError != nil {
		errorMessage = fmt.Sprintf("%s : %s", forwardError, proxyError.Error())
		logrus.Error(errorMessage)
		return false, errors.New(errorMessage)
	}

	responseMap := map[string]interface{}{}
	readAndUnmarshalError := helpers.ReadAndUnmarshal(proxiedResponse.Body, &responseMap)

	if readAndUnmarshalError != nil {
		errorMessage = fmt.Sprintf("%s : %s", unmarshallError, readAndUnmarshalError.Error())
		logrus.Error(errorMessage)
		return false, errors.New(errorMessage)
	}

	if proxiedResponse.StatusCode != 200 {
		errorMessage = fmt.Sprintf("%s : %d", statusError, proxiedResponse.StatusCode)
		logrus.WithFields(logrus.Fields{
			"response": responseMap,
		}).Error(errorMessage)
		return false, errors.New(errorMessage)
	}

	//check this one if this is a response success
	logrus.WithFields(logrus.Fields{
		"response": responseMap,
	}).Info("Create chat sucessfully")
	return true, nil
}

/*
// Get messages from chat of salesforce.
func (c *SalesforceClient) GetMessages(affinityToken, sessionKey string) ([]Message, error) {
	var errorMessage string
	//building request to send through proxy
	header := make(map[string]string)
	header[apiVersionHeader] = apiversionValue
	header[affinityHeader] = affinityToken
	header[sessionKeyHeader] = sessionKey

	c.Proxy.SetBaseURL(c.SalesforceURL)
	newRequest := proxy.Request{
		Method:    get,
		URI:       "chat/rest/System/Messages",
		HeaderMap: header,
	}

	proxiedResponse, proxyError := c.Proxy.SendHTTPRequest(&newRequest)
	if proxyError != nil {
		errorMessage = fmt.Sprintf("%s : %s", forwardError, proxyError.Error())
		logrus.Error(errorMessage)
		return false, errors.New(errorMessage)
	}

	responseMap := map[string]interface{}{}
	readAndUnmarshalError := helpers.ReadAndUnmarshal(proxiedResponse.Body, &responseMap)

	if readAndUnmarshalError != nil {
		errorMessage = fmt.Sprintf("%s : %s", unmarshallError, readAndUnmarshalError.Error())
		logrus.Error(errorMessage)
		return false, errors.New(errorMessage)
	}

	if proxiedResponse.StatusCode != 200 {
		errorMessage = fmt.Sprintf("%s : %d", statusError, proxiedResponse.StatusCode)
		logrus.WithFields(logrus.Fields{
			"response": responseMap,
		}).Error(errorMessage)
		return false, errors.New(errorMessage)
	}

	//check this one if this is a response success
	logrus.WithFields(logrus.Fields{
		"response": responseMap,
	}).Info("Send Message sucessfully")
	return true, nil
}*/

// To send messages to the live agent user.
func (c *SalesforceClient) SendMessage(affinityToken, sessionKey string, payload MessagePayload) (bool, error) {
	var errorMessage string
	// This log should hide the secrets before sending to production
	logrus.WithFields(logrus.Fields{
		"payload": payload,
	}).Info("Payload received")

	//validating token Payload struct
	if err := helpers.Govalidator().Struct(payload); err != nil {
		errorMessage = fmt.Sprintf("%s : %s", helpers.InvalidPayload, err.Error())
		logrus.Error(errorMessage)
		return false, errors.New(errorMessage)
	}

	//building request to send through proxy
	requestBytes, _ := json.Marshal(payload)

	header := make(map[string]string)
	header[apiVersionHeader] = strconv.Itoa(c.ApiVersion)
	header[affinityHeader] = affinityToken
	header[sessionKeyHeader] = sessionKey

	c.Proxy.SetBaseURL(c.SalesforceURL)
	newRequest := proxy.Request{
		Body:      requestBytes,
		Method:    post,
		URI:       "/chat/rest/Chasitor/ChatMessage",
		HeaderMap: header,
	}

	proxiedResponse, proxyError := c.Proxy.SendHTTPRequest(&newRequest)
	if proxyError != nil {
		errorMessage = fmt.Sprintf("%s : %s", forwardError, proxyError.Error())
		logrus.Error(errorMessage)
		return false, errors.New(errorMessage)
	}

	responseMap := map[string]interface{}{}
	readAndUnmarshalError := helpers.ReadAndUnmarshal(proxiedResponse.Body, &responseMap)

	if readAndUnmarshalError != nil {
		errorMessage = fmt.Sprintf("%s : %s", unmarshallError, readAndUnmarshalError.Error())
		logrus.Error(errorMessage)
		return false, errors.New(errorMessage)
	}

	if proxiedResponse.StatusCode != 200 {
		errorMessage = fmt.Sprintf("%s : %d", statusError, proxiedResponse.StatusCode)
		logrus.WithFields(logrus.Fields{
			"response": responseMap,
		}).Error(errorMessage)
		return false, errors.New(errorMessage)
	}

	//check this one if this is a response success
	logrus.WithFields(logrus.Fields{
		"response": responseMap,
	}).Info("Send Message sucessfully")
	return true, nil
}
