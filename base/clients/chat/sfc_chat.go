package chat

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"io"
	"net/http"
	"yalochat.com/salesforce-integration/base/events"

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
	bodyReason       = `{"reason": "client"}`

	ChatRequestFail    = "ChatRequestFail"
	ChatRequestSuccess = "ChatRequestSuccess"
	QueueUpdate        = "QueueUpdate"
	ChatEstablished    = "ChatEstablished"
	ChatMessage        = "ChatMessage"
	AgentTyping        = "AgentTyping"
	AgentNotTyping     = "AgentNotTyping"
	ChatEnded          = "ChatEnded"
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
	OrganizationId      string                  `json:"organizationId" validate:"required"`
	DeploymentId        string                  `json:"deploymentId" validate:"required"`
	ButtonId            string                  `json:"buttonId" validate:"required"`
	SessionId           string                  `json:"sessionId" validate:"required"`
	UserAgent           string                  `json:"userAgent" validate:"required"`
	Language            string                  `json:"language" validate:"required"`
	ScreenResolution    string                  `json:"screenResolution" validate:"required"`
	VisitorName         string                  `json:"visitorName" validate:"required"`
	PrechatDetails      []PreChatDetailsObject  `json:"prechatDetails" validate:"required,dive"`
	PrechatEntities     []PrechatEntitiesObject `json:"prechatEntities" validate:"required,dive"`
	ReceiveQueueUpdates bool                    `json:"receiveQueueUpdates"`
	IsPost              bool                    `json:"isPost"`
}

type PreChatDetailsObject struct {
	Label            string   `json:"label" validate:"required"`
	Value            string   `json:"value" validate:"required"`
	DisplayToAgent   bool     `json:"displayToAgent"`
	TranscriptFields []string `json:"transcriptFields" validate:"required"`
}

type PrechatEntitiesObject struct {
	EntityName        string        `json:"entityName" validate:"required"`
	LinkToEntityName  string        `json:"linkToEntityName" validate:"required"`
	LinkToEntityField string        `json:"linkToEntityField" validate:"required"`
	SaveToTranscript  string        `json:"saveToTranscript" validate:"required"`
	ShowOnCreate      bool          `json:"showOnCreate"`
	EntityFieldsMaps  []EntityField `json:"entityFieldsMaps" validate:"required,dive"`
}

type EntityField struct {
	FieldName    string `json:"fieldName" validate:"required"`
	Label        string `json:"label" validate:"required"`
	DoFind       bool   `json:"doFind"`
	IsExactMatch bool   `json:"isExactMatch"`
	DoCreate     bool   `json:"doCreate"`
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
	Reason                string                 `json:"reason,omitempty"`
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
	CreateSession(mainSpan tracer.Span) (*SessionResponse, error)
	CreateChat(tracer.Span, string, string, ChatRequest) (bool, error)
	GetMessages(mainSpan tracer.Span, affinityToken, sessionKey string) (*MessagesResponse, *helpers.ErrorResponse)
	SendMessage(tracer.Span, string, string, MessagePayload) (bool, error)
	ChatEnd(affinityToken, sessionKey string) error
	ReconnectSession(affinityToken, sessionKey, offset string) (*MessagesResponse, error)
	UpdateToken(accessToken string)
}

func NewChatRequest(organizationID, deployementID, sessionID, ButtonID, userName string) ChatRequest {
	return ChatRequest{
		OrganizationId:      organizationID,
		DeploymentId:        deployementID,
		ButtonId:            ButtonID,
		SessionId:           sessionID,
		VisitorName:         userName,
		UserAgent:           "Yalo Bot",
		Language:            "es-MX",
		ScreenResolution:    "1900x1080",
		PrechatDetails:      []PreChatDetailsObject{},
		PrechatEntities:     []PrechatEntitiesObject{},
		ReceiveQueueUpdates: true,
		IsPost:              true,
	}
}

//CreateSession To create a new Live Agent session, you must call the SessionId request.
//SessionId : Establishes a new Live Agent session. The SessionId request is required as the first request to create every new Live Agent session.
func (c *SfcChatClient) CreateSession(mainSpan tracer.Span) (*SessionResponse, error) {
	// datadog tracing
	spanContext := events.GetSpanContextFromSpan(mainSpan)
	span := tracer.StartSpan("create_session", tracer.ChildOf(spanContext))
	span.SetTag(ext.AnalyticsEvent, true)
	defer span.Finish()
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
	span.SetTag(ext.ResourceName, fmt.Sprintf("%s %s", newRequest.Method, newRequest.URI))

	proxiedResponse, proxyError := c.Proxy.SendHTTPRequest(span, &newRequest)
	if proxyError != nil {
		errorMessage = fmt.Sprintf("%s : %s", constants.ForwardError, proxyError.Error())
		logrus.Error(errorMessage)
		span.SetTag(ext.Error, proxyError)
		return nil, errors.New(errorMessage)
	}

	if proxiedResponse.StatusCode != http.StatusOK {
		buf := new(bytes.Buffer)
		buf.ReadFrom(proxiedResponse.Body)
		bodyResponse := buf.String()
		errorMessage = fmt.Sprintf("[%d] - %s : %s", proxiedResponse.StatusCode, constants.StatusError, bodyResponse)
		logrus.Error(errorMessage)
		err := errors.New(errorMessage)
		span.SetTag(ext.Error, err)
		return nil, err
	}

	var session SessionResponse
	readAndUnmarshalError := helpers.ReadAndUnmarshal(proxiedResponse.Body, &session)

	if readAndUnmarshalError != nil {
		errorMessage = fmt.Sprintf("%s : %s", constants.UnmarshallError, readAndUnmarshalError.Error())
		logrus.Error(errorMessage)
		span.SetTag(ext.Error, readAndUnmarshalError)
		return nil, errors.New(errorMessage)
	}

	//check this one if this is a response success
	logrus.WithFields(logrus.Fields{
		"response": session,
	}).Info("Get Session sucessfully")
	return &session, nil
}

//CreateChat We’ll send a chat request. To make sure this works, you should log in as a Live Agent user and make yourself available.
//Initiates a new chat visitor session. The ChasitorInit request is always required as the first POST request in a new chat session.
func (c *SfcChatClient) CreateChat(mainSpan tracer.Span, affinityToken, sessionKey string, request ChatRequest) (bool, error) {
	// datadog tracing
	spanContext := events.GetSpanContextFromSpan(mainSpan)
	span := tracer.StartSpan("create_chat", tracer.ChildOf(spanContext))
	span.SetTag(ext.AnalyticsEvent, true)
	span.SetTag(events.Payload, fmt.Sprintf("%#v", request))
	defer span.Finish()

	var errorMessage string
	// This log should hide the secrets before sending to production
	logrus.WithFields(logrus.Fields{
		"payload": request,
	}).Info("Payload received")

	//validating token Payload struct
	if err := helpers.Govalidator().Struct(request); err != nil {
		errorMessage = fmt.Sprintf("%s : %s", helpers.InvalidPayload, err.Error())
		logrus.Error(errorMessage)
		span.SetTag(ext.Error, err)
		return false, errors.New(errorMessage)
	}

	//building request to send through proxy
	requestBytes, _ := json.Marshal(request)

	newRequest := c.getRequest(
		affinityToken,
		sessionKey,
		http.MethodPost,
		"/chat/rest/Chasitor/ChasitorInit",
		requestBytes)
	span.SetTag(ext.ResourceName, fmt.Sprintf("%s %s", newRequest.Method, newRequest.URI))

	proxiedResponse, proxyError := c.Proxy.SendHTTPRequest(span, &newRequest)
	if proxyError != nil {
		errorMessage = fmt.Sprintf("%s : %s", constants.ForwardError, proxyError.Error())
		logrus.Error(errorMessage)
		span.SetTag(ext.Error, proxyError)
		return false, errors.New(errorMessage)
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(proxiedResponse.Body)
	bodyResponse := buf.String()

	if proxiedResponse.StatusCode != http.StatusOK {
		errorMessage = fmt.Sprintf("[%d] - %s : %s", proxiedResponse.StatusCode, constants.StatusError, bodyResponse)
		logrus.Error(errorMessage)
		err := errors.New(errorMessage)
		span.SetTag(ext.Error, err)
		return false, err
	}

	//check this one if this is a response success
	logrus.WithFields(logrus.Fields{
		"response": bodyResponse,
	}).Info("Create chat sucessfully")
	return true, nil
}

// GetMessages Get messages from chat of salesforce.
func (c *SfcChatClient) GetMessages(mainSpan tracer.Span, affinityToken, sessionKey string) (*MessagesResponse, *helpers.ErrorResponse) {
	// datadog tracing
	spanContext := events.GetSpanContextFromSpan(mainSpan)
	span := tracer.StartSpan("get_messages", tracer.ChildOf(spanContext))
	span.SetTag(ext.AnalyticsEvent, true)
	defer span.Finish()
	var errorMessage string

	newRequest := c.getRequest(
		affinityToken,
		sessionKey,
		http.MethodGet,
		"/chat/rest/System/Messages",
		nil)
	span.SetTag(ext.ResourceName, fmt.Sprintf("%s %s", newRequest.Method, newRequest.URI))

	proxiedResponse, proxyError := c.Proxy.SendHTTPRequest(span, &newRequest)
	if proxyError != nil {
		errorMessage = fmt.Sprintf("%s : %s", constants.ForwardError, proxyError.Error())
		logrus.Error(errorMessage)
		span.SetTag(ext.Error, proxyError)
		return nil, &helpers.ErrorResponse{Error: errors.New(errorMessage)}
	}

	if proxiedResponse.StatusCode != http.StatusOK {
		buf := new(bytes.Buffer)
		buf.ReadFrom(proxiedResponse.Body)
		bodyResponse := buf.String()
		errorMessage = fmt.Sprintf("[%d] - %s : %s", proxiedResponse.StatusCode, constants.StatusError, bodyResponse)
		if proxiedResponse.StatusCode != http.StatusNoContent {
			span.SetTag(ext.Error, errors.New(errorMessage))
			logrus.Error(errorMessage)
		}

		return nil, &helpers.ErrorResponse{StatusCode: proxiedResponse.StatusCode, Error: errors.New(errorMessage)}
	}

	var messages MessagesResponse
	readAndUnmarshalError := helpers.ReadAndUnmarshal(proxiedResponse.Body, &messages)

	if readAndUnmarshalError != nil {
		errorMessage = fmt.Sprintf("%s : %s", constants.UnmarshallError, readAndUnmarshalError.Error())
		logrus.Error(errorMessage)
		span.SetTag(ext.Error, readAndUnmarshalError)
		return nil, &helpers.ErrorResponse{Error: errors.New(errorMessage)}
	}

	//check this one if this is a response success
	logrus.WithFields(logrus.Fields{
		"response": messages,
	}).Info("Get Messages sucessfully")
	return &messages, nil
}

// SendMessage to send messages to the live agent user.
func (c *SfcChatClient) SendMessage(mainSpan tracer.Span, affinityToken, sessionKey string, payload MessagePayload) (bool, error) {
	// datadog tracing
	spanContext := events.GetSpanContextFromSpan(mainSpan)
	span := tracer.StartSpan("send_message", tracer.ChildOf(spanContext))
	span.SetTag(ext.AnalyticsEvent, true)
	span.SetTag(events.Payload, fmt.Sprintf("%#v", payload))
	defer span.Finish()

	uri := "/chat/rest/Chasitor/ChatMessage"
	span.SetTag(ext.ResourceName, fmt.Sprintf("%s %s", http.MethodPost, uri))

	var errorMessage string
	// This log should hide the secrets before sending to production
	logrus.WithFields(logrus.Fields{
		"payload": payload,
	}).Info("Payload received")

	//validating Payload struct
	if err := helpers.Govalidator().Struct(payload); err != nil {
		errorMessage = fmt.Sprintf("%s : %s", helpers.InvalidPayload, err.Error())
		logrus.Error(errorMessage)
		span.SetTag(ext.Error, err)
		return false, errors.New(errorMessage)
	}

	//building request to send through proxy
	requestBytes, _ := json.Marshal(payload)

	newRequest := c.getRequest(
		affinityToken,
		sessionKey,
		http.MethodPost,
		uri,
		requestBytes)

	proxiedResponse, proxyError := c.Proxy.SendHTTPRequest(span, &newRequest)
	if proxyError != nil {
		errorMessage = fmt.Sprintf("%s : %s", constants.ForwardError, proxyError.Error())
		logrus.Error(errorMessage)
		span.SetTag(ext.Error, proxyError)
		return false, errors.New(errorMessage)
	}

	response, err := stringResponse(proxiedResponse.Body)
	if err != nil {
		span.SetTag(ext.Error, err)
		return false, err
	}

	if proxiedResponse.StatusCode != http.StatusOK {
		errorMessage = fmt.Sprintf("%s-[%d]: %s", constants.StatusError, proxiedResponse.StatusCode, response)
		logrus.WithFields(logrus.Fields{
			"response": response,
		}).Error(errorMessage)
		err := errors.New(errorMessage)
		span.SetTag(ext.Error, err)
		return false, err
	}

	//check this one if this is a response success
	logrus.WithFields(logrus.Fields{
		"response": response,
	}).Info("Send Message sucessfully")
	return true, nil
}

//ReconnectSession Reconnet session to the live agent user.
func (c *SfcChatClient) ReconnectSession(affinityToken, sessionKey, offset string) (*MessagesResponse, error) {
	// datadog tracing
	span := tracer.StartSpan("reconnect_session")
	span.SetTag(ext.AnalyticsEvent, true)
	defer span.Finish()

	var errorMessage string
	logrus.WithFields(logrus.Fields{
		"offset": offset,
	}).Info("query params received")

	//validating query params struct
	if offset == "" {
		errorMessage = fmt.Sprintf("%s ", constants.QueryParamError)
		logrus.Error(errorMessage)
		err := errors.New(errorMessage)
		span.SetTag(ext.Error, err)
		return nil, err
	}

	newRequest := c.getRequest(
		affinityToken,
		sessionKey,
		http.MethodPost,
		fmt.Sprintf("/chat/rest/System/ReconnectSession?ReconnectSession.offset=%s", offset),
		nil)
	span.SetTag(ext.ResourceName, fmt.Sprintf("%s %s", newRequest.Method, newRequest.URI))

	proxiedResponse, proxyError := c.Proxy.SendHTTPRequest(span, &newRequest)
	if proxyError != nil {
		errorMessage = fmt.Sprintf("%s : %s", constants.ForwardError, proxyError.Error())
		logrus.Error(errorMessage)
		span.SetTag(ext.Error, proxyError)
		return nil, errors.New(errorMessage)
	}

	if proxiedResponse.StatusCode != http.StatusOK {
		err := helpers.ErrorResponseMap(proxiedResponse.Body, constants.StatusError, proxiedResponse.StatusCode)
		span.SetTag(ext.Error, err)
		return nil, err
	}

	var session MessagesResponse
	readAndUnmarshalError := helpers.ReadAndUnmarshal(proxiedResponse.Body, &session)
	if readAndUnmarshalError != nil {
		errorMessage = fmt.Sprintf("%s : %s", constants.UnmarshallError, readAndUnmarshalError.Error())
		logrus.Error(errorMessage)
		span.SetTag(ext.Error, readAndUnmarshalError)
		return nil, errors.New(errorMessage)
	}

	//check this one if this is a response success
	logrus.WithFields(logrus.Fields{
		"response": session,
	}).Info("Reconnet session sucessfully")
	return &session, nil
}

//ChatEnd end chat of salesforce.
func (c *SfcChatClient) ChatEnd(affinityToken, sessionKey string) error {
	// datadog tracing
	span := tracer.StartSpan("chat_end")
	span.SetTag(ext.AnalyticsEvent, true)
	defer span.Finish()
	var errorMessage string

	newRequest := c.getRequest(
		affinityToken,
		sessionKey,
		http.MethodPost,
		"/chat/rest/Chasitor/ChatEnd",
		[]byte(bodyReason))
	span.SetTag(ext.ResourceName, fmt.Sprintf("%s %s", newRequest.Method, newRequest.URI))

	proxiedResponse, proxyError := c.Proxy.SendHTTPRequest(span, &newRequest)
	if proxyError != nil {
		errorMessage = fmt.Sprintf("%s : %s", constants.ForwardError, proxyError.Error())
		logrus.Error(errorMessage)
		span.SetTag(ext.Error, proxyError)
		return errors.New(errorMessage)
	}

	response, err := stringResponse(proxiedResponse.Body)
	if err != nil {
		span.SetTag(ext.Error, err)
		return err
	}

	if proxiedResponse.StatusCode != http.StatusOK {
		errorMessage = fmt.Sprintf("%s : %d", constants.StatusError, proxiedResponse.StatusCode)
		logrus.WithFields(logrus.Fields{
			"response": response,
		}).Error(errorMessage)
		err := errors.New(errorMessage)
		span.SetTag(ext.Error, err)
		return err
	}

	logrus.WithFields(logrus.Fields{
		"response": response,
	}).Info("Chat end sucessfully")

	return nil
}

func stringResponse(body io.ReadCloser) (string, error) {
	buf := new(bytes.Buffer)

	_, err := buf.ReadFrom(body)
	if err != nil {
		errorMessage := fmt.Sprintf("%s : %s", constants.ResponseError, err.Error())
		logrus.Error(errorMessage)
		return "", errors.New(errorMessage)
	}

	return buf.String(), nil
}

func (c *SfcChatClient) getRequest(affinityToken, sessionKey, method, uri string, body []byte) proxy.Request {
	header := make(map[string]string)
	header[apiVersionHeader] = c.ApiVersion
	header[affinityHeader] = affinityToken
	header[sessionKeyHeader] = sessionKey

	newRequest := proxy.Request{
		Method:    method,
		URI:       uri,
		HeaderMap: header,
	}

	if body != nil {
		newRequest.Body = body
	}

	return newRequest
}

func (c *SfcChatClient) UpdateToken(accessToken string) {
	c.AccessToken = accessToken
}
