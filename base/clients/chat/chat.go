package chat

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"yalochat.com/salesforce-integration/base/clients/proxy"
	"yalochat.com/salesforce-integration/base/constants"
	"yalochat.com/salesforce-integration/base/helpers"
)

const (
	granType         = "password"
	apiVersionHeader = "X-LIVEAGENT-API-VERSION"
	affinityHeader   = "X-LIVEAGENT-AFFINITY"
	sessionKeyHeader = "X-LIVEAGENT-SESSION-KEY"
	sequenceHeader   = "X-LIVEAGENT-SEQUENCE"
)

type SfcChatClient struct {
	ApiVersion  string
	AccessToken string
	Proxy       proxy.ProxyInterface
}

type SessionResponse struct {
	ClientPollTimeout int    `json:"clientPollTimeout"`
	Key               string `json:"key"`
	AffinityToken     string `json:"affinityToken"`
	Id                string `json:"id"`
}

type ChatRequest struct {
	OrganizationId      string        `json:"organizationId" validate:"required"`
	DeploymentId        string        `json:"deploymentId" validate:"required"`
	ButtonId            string        `json:"buttonId" validate:"required"`
	SessionId           string        `json:"sessionId" validate:"required"`
	UserAgent           string        `json:"userAgent" validate:"required"`
	Language            string        `json:"language" validate:"required"`
	ScreenResolution    string        `json:"screenResolution" validate:"required"`
	VisitorName         string        `json:"visitorName" validate:"required"`
	PrechatDetails      []interface{} `json:"prechatDetails" validate:"required"`
	PrechatEntities     []interface{} `json:"prechatEntities" validate:"required"`
	ReceiveQueueUpdates bool          `json:"receiveQueueUpdates" validate:"required"`
	IsPost              bool          `json:"isPost" validate:"required"`
}

type MessagePayload struct {
	Text string `json:"text" validate:"required"`
}

type MessagesResponse struct {
	Messages []MessageObject `json:"messages,omitempty"`
	Sequence int             `json:"sequence,omitempty"`
	Offset   int             `json:"offset,omitempty"`
}

type MessageObject struct {
	Type    string  `json:"type,omitempty"`
	Message Message `json:"message,omitempty"`
}

type Message struct {
	Name                  string                 `json:"name,omitempty"`
	UserId                string                 `json:"userId,omitempty"`
	AgentId               string                 `json:"agentId,omitempty"`
	Text                  string                 `json:"text,omitempty"`
	Schedule              map[string]interface{} `json:"schedule,omitempty"`
	Items                 []string               `json:"items,omitempty"`
	SneakPeekEnabled      bool                   `json:"sneakPeekEnabled,omitempty"`
	ChasitorIdleTimeout   map[string]interface{} `json:"chasitorIdleTimeout,omitempty"`
	ConnectionTimeout     int                    `json:"connectionTimeout,omitempty"`
	Position              int                    `json:"position,omitempty"`
	EstimatedWaitTime     int                    `json:"estimatedWaitTime,omitempty"`
	SensitiveDataRules    []string               `json:"sensitiveDataRules,omitempty"`
	TranscriptSaveEnabled bool                   `json:"transcriptSaveEnabled,omitempty"`
	Url                   string                 `json:"url,omitempty"`
	QueuePosition         int                    `json:"queuePosition,omitempty"`
	CustomDetails         []string               `json:"customDetails,omitempty"`
	VisitorId             string                 `json:"visitorId,omitempty"`
	Type                  string                 `json:"type,omitempty"`
	GeoLocation           GeoLocation            `json:"geoLocation,omitempty"`
	ResetSequence         bool                   `json:"resetSequence,omitempty"`
	AffinityToken         string                 `json:"affinityToken,omitempty"`
}

type GeoLocation struct {
	Organization string  `json:"organization,omitempty"`
	CountryName  string  `json:"countryName,omitempty"`
	CountryCode  string  `json:"countryCode,omitempty"`
	Latitude     float32 `json:"latitude,omitempty"`
	Longitude    float32 `json:"longitude,omitempty"`
}

type SfcChatInterface interface {
	CreateSession() (*SessionResponse, error)
	CreateChat(string, string, ChatRequest) (bool, error)
	GetMessages(string, string) ([]Message, *helpers.ErrorResponse)
	SendMessage(string, string, MessagePayload) (bool, error)
	EndChat(string, string) (bool, error)
	ReconnectSession(affinityToken, sessionKey, offset string) (*MessagesResponse, error)
}

func NewChatRequest(organizationID, deployementID, seassionID, ButtonID, userName string) ChatRequest {
	return ChatRequest{
		OrganizationId:      organizationID,
		DeploymentId:        deployementID,
		ButtonId:            ButtonID,
		SessionId:           seassionID,
		VisitorName:         userName,
		UserAgent:           "WhatsApp",
		Language:            "en-US",
		ScreenResolution:    "1900x1080",
		PrechatDetails:      []interface{}{},
		PrechatEntities:     []interface{}{},
		ReceiveQueueUpdates: true,
		IsPost:              true,
	}
}

//To create a new Live Agent session, you must call the SessionId request.
//SessionId : Establishes a new Live Agent session. The SessionId request is required as the first request to create every new Live Agent session.
func (c *SfcChatClient) CreateSession() (*SessionResponse, error) {
	var errorMessage string

	//building request to send through proxy
	header := make(map[string]string)
	header[apiVersionHeader] = c.ApiVersion
	header[affinityHeader] = "null"

	newRequest := proxy.Request{
		Method:    http.MethodGet,
		URI:       "/chat/rest/System/SessionId",
		HeaderMap: header,
	}

	proxiedResponse, proxyError := c.Proxy.SendHTTPRequest(&newRequest)
	if proxyError != nil {
		errorMessage = fmt.Sprintf("%s : %s", constants.ForwardError, proxyError.Error())
		logrus.Error(errorMessage)
		return nil, errors.New(errorMessage)
	}

	if proxiedResponse.StatusCode != http.StatusOK {
		buf := new(bytes.Buffer)
		buf.ReadFrom(proxiedResponse.Body)
		bodyResponse := buf.String()
		errorMessage = fmt.Sprintf("[%d] - %s : %s", proxiedResponse.StatusCode, constants.StatusError, bodyResponse)
		logrus.Error(errorMessage)
		return nil, errors.New(errorMessage)
	}

	var session SessionResponse
	readAndUnmarshalError := helpers.ReadAndUnmarshal(proxiedResponse.Body, &session)

	if readAndUnmarshalError != nil {
		errorMessage = fmt.Sprintf("%s : %s", constants.UnmarshallError, readAndUnmarshalError.Error())
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
//Initiates a new chat visitor session. The ChasitorInit request is always required as the first POST request in a new chat session.
func (c *SfcChatClient) CreateChat(affinityToken, sessionKey string, request ChatRequest) (bool, error) {
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
	header[apiVersionHeader] = c.ApiVersion
	header[affinityHeader] = affinityToken
	header[sessionKeyHeader] = sessionKey
	header[sequenceHeader] = "1"

	newRequest := proxy.Request{
		Body:      requestBytes,
		Method:    http.MethodPost,
		URI:       "/chat/rest/Chasitor/ChasitorInit",
		HeaderMap: header,
	}

	proxiedResponse, proxyError := c.Proxy.SendHTTPRequest(&newRequest)
	if proxyError != nil {
		errorMessage = fmt.Sprintf("%s : %s", constants.ForwardError, proxyError.Error())
		logrus.Error(errorMessage)
		return false, errors.New(errorMessage)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(proxiedResponse.Body)
	bodyResponse := buf.String()

	if proxiedResponse.StatusCode != http.StatusOK {
		errorMessage = fmt.Sprintf("[%d] - %s : %s", proxiedResponse.StatusCode, constants.StatusError, bodyResponse)
		logrus.Error(errorMessage)
		return false, errors.New(errorMessage)
	}

	//check this one if this is a response success
	logrus.WithFields(logrus.Fields{
		"response": bodyResponse,
	}).Info("Create chat sucessfully")
	return true, nil
}

// Get messages from chat of salesforce.
func (c *SfcChatClient) GetMessages(affinityToken, sessionKey string) (*MessagesResponse, *helpers.ErrorResponse) {
	var errorMessage string
	//building request to send through proxy
	header := make(map[string]string)
	header[apiVersionHeader] = c.ApiVersion
	header[affinityHeader] = affinityToken
	header[sessionKeyHeader] = sessionKey

	newRequest := proxy.Request{
		Method:    http.MethodGet,
		URI:       "/chat/rest/System/Messages",
		HeaderMap: header,
	}

	proxiedResponse, proxyError := c.Proxy.SendHTTPRequest(&newRequest)
	if proxyError != nil {
		errorMessage = fmt.Sprintf("%s : %s", constants.ForwardError, proxyError.Error())
		logrus.Error(errorMessage)
		return nil, &helpers.ErrorResponse{Error: errors.New(errorMessage)}
	}

	if proxiedResponse.StatusCode != http.StatusOK {
		responseMap := map[string]interface{}{}
		readAndUnmarshalError := helpers.ReadAndUnmarshal(proxiedResponse.Body, &responseMap)

		if readAndUnmarshalError != nil {
			errorMessage = fmt.Sprintf("%s : %s", constants.UnmarshallError, readAndUnmarshalError.Error())
			logrus.Error(errorMessage)
			return nil, &helpers.ErrorResponse{Error: errors.New(errorMessage)}
		}
		errorMessage = fmt.Sprintf("%s : %d", constants.StatusError, proxiedResponse.StatusCode)
		logrus.WithFields(logrus.Fields{
			"response": responseMap,
		}).Error(errorMessage)
		return nil, &helpers.ErrorResponse{
			StatusCode: proxiedResponse.StatusCode,
			Error:      errors.New(errorMessage),
		}
	}

	var messages MessagesResponse
	readAndUnmarshalError := helpers.ReadAndUnmarshal(proxiedResponse.Body, &messages)

	if readAndUnmarshalError != nil {
		errorMessage = fmt.Sprintf("%s : %s", constants.UnmarshallError, readAndUnmarshalError.Error())
		logrus.Error(errorMessage)
		return nil, &helpers.ErrorResponse{Error: errors.New(errorMessage)}
	}

	//check this one if this is a response success
	logrus.WithFields(logrus.Fields{
		"response": messages,
	}).Info("Get Messages sucessfully")
	return &messages, nil
}

// SendMessage to send messages to the live agent user.
func (c *SfcChatClient) SendMessage(affinityToken, sessionKey string, payload MessagePayload) (bool, error) {
	var errorMessage string
	// This log should hide the secrets before sending to production
	logrus.WithFields(logrus.Fields{
		"payload": payload,
	}).Info("Payload received")

	//validating Payload struct
	if err := helpers.Govalidator().Struct(payload); err != nil {
		errorMessage = fmt.Sprintf("%s : %s", helpers.InvalidPayload, err.Error())
		logrus.Error(errorMessage)
		return false, errors.New(errorMessage)
	}

	//building request to send through proxy
	requestBytes, _ := json.Marshal(payload)

	header := make(map[string]string)
	header[apiVersionHeader] = c.ApiVersion
	header[affinityHeader] = affinityToken
	header[sessionKeyHeader] = sessionKey

	newRequest := proxy.Request{
		Body:      requestBytes,
		Method:    http.MethodPost,
		URI:       "/chat/rest/Chasitor/ChatMessage",
		HeaderMap: header,
	}

	proxiedResponse, proxyError := c.Proxy.SendHTTPRequest(&newRequest)
	if proxyError != nil {
		errorMessage = fmt.Sprintf("%s : %s", constants.ForwardError, proxyError.Error())
		logrus.Error(errorMessage)
		return false, errors.New(errorMessage)
	}

	buf := new(bytes.Buffer)
	_, err := buf.ReadFrom(proxiedResponse.Body)
	if err != nil {
		errorMessage = fmt.Sprintf("%s : %s", constants.ResponseError, err.Error())
		logrus.Error(errorMessage)
		return false, errors.New(errorMessage)
	}
	response := buf.String()

	if proxiedResponse.StatusCode != http.StatusOK {
		errorMessage = fmt.Sprintf("%s : %d", constants.StatusError, proxiedResponse.StatusCode)
		logrus.WithFields(logrus.Fields{
			"response": response,
		}).Error(errorMessage)
		return false, errors.New(errorMessage)
	}

	//check this one if this is a response success
	logrus.WithFields(logrus.Fields{
		"response": response,
	}).Info("Send Message sucessfully")
	return true, nil
}

//ReconnectSession Reconnet session to the live agent user.
func (c *SfcChatClient) ReconnectSession(affinityToken, sessionKey, offset string) (*MessagesResponse, error) {
	var errorMessage string
	logrus.WithFields(logrus.Fields{
		"offset": offset,
	}).Info("query params received")

	//validating query params struct
	if offset == "" {
		errorMessage = fmt.Sprintf("%s ", constants.QueryParamError)
		logrus.Error(errorMessage)
		return nil, errors.New(errorMessage)
	}

	header := make(map[string]string)
	header[apiVersionHeader] = c.ApiVersion
	header[affinityHeader] = affinityToken
	header[sessionKeyHeader] = sessionKey

	newRequest := proxy.Request{
		Body:      []byte{},
		Method:    http.MethodPost,
		URI:       fmt.Sprintf("/chat/rest/System/ReconnectSession?ReconnectSession.offset=%s", offset),
		HeaderMap: header,
	}

	proxiedResponse, proxyError := c.Proxy.SendHTTPRequest(&newRequest)
	if proxyError != nil {
		errorMessage = fmt.Sprintf("%s : %s", constants.ForwardError, proxyError.Error())
		logrus.Error(errorMessage)
		return nil, errors.New(errorMessage)
	}

	if proxiedResponse.StatusCode != 200 {
		return nil, helpers.ErrorResponseMap(proxiedResponse.Body, constants.StatusError, proxiedResponse.StatusCode)
	}

	var session MessagesResponse
	readAndUnmarshalError := helpers.ReadAndUnmarshal(proxiedResponse.Body, &session)

	if readAndUnmarshalError != nil {
		errorMessage = fmt.Sprintf("%s : %s", constants.UnmarshallError, readAndUnmarshalError.Error())
		logrus.Error(errorMessage)
		return nil, errors.New(errorMessage)
	}

	//check this one if this is a response success
	logrus.WithFields(logrus.Fields{
		"response": session,
	}).Info("Reconnet session sucessfully")
	return &session, nil
}
