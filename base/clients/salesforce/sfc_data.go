package salesforce

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"yalochat.com/salesforce-integration/base/events"

	"github.com/sirupsen/logrus"
	"yalochat.com/salesforce-integration/base/clients/proxy"
	"yalochat.com/salesforce-integration/base/constants"
	"yalochat.com/salesforce-integration/base/helpers"
	"yalochat.com/salesforce-integration/base/models"
)

const (
	AutoAsssingHeader = "Sforce-Auto-Assign"
	//TODO: the CP_BlockedChatYalo__c field is customized for Coppel and that in other integrations we must remove from the query
	queryForContactByField = `SELECT+id+,+firstName+,+lastName+,+mobilePhone+,+email+%sFROM+Contact+WHERE+%s+=+'%s'`
	queryForAccountByField = `SELECT+id+,+firstName+,+lastName+,+PersonMobilePhone+,+PersonEmail+,+PersonContactId+FROM+Account+WHERE+%s+=+'%s'`
	blockedChatYaloField   = ",+CP_BlockedChatYalo__c+"
)

func NewSalesforceRequester(url, token string) *SalesforceClient {
	return &SalesforceClient{
		Proxy:       proxy.NewProxy(url, 30, 3, 1, 30),
		AccessToken: token,
	}
}

//SalesforceClient settings for use a case client
type SalesforceClient struct {
	AccessToken         string
	APIVersion          string
	Proxy               proxy.ProxyInterface
	SfcBlockedChatField bool
}

//ContentVersionPayload handles a content version response
type ContentVersionPayload struct {
	Title           string `json:"Title" validate:"required"`
	Description     string `json:"Description" validate:"omitempty"`
	ContentLocation string `json:"ContentLocation" validate:"required"`
	PathOnClient    string `json:"PathOnClient" validate:"required"`
	VersionData     string `json:"VersionData" validate:"required"`
}

//LinkDocumentPayload handles a link document response
type LinkDocumentPayload struct {
	ContentDocumentID string `json:"ContentDocumentId" validate:"required"`
	LinkedEntityID    string `json:"LinkedEntityId" validate:"required"`
	ShareType         string `json:"shareType" validate:"required"`
	Visibility        string `json:"visibility" validate:"required"`
}

type CaseRequest struct {
	RecordTypeID    string      `json:"RecordTypeId"`
	ContactID       string      `json:"ContactId" validate:"required"`
	OwnerID         string      `json:"OwnerId"`
	Description     string      `json:"Description" validate:"required"`
	Origin          string      `json:"Origin" validate:"required"`
	Priority        string      `json:"Priority" validate:"required"`
	Status          string      `json:"Status" validate:"required"`
	Subject         string      `json:"Subject" validate:"required"`
	IsEscalated     bool        `json:"IsEscalated"`
	AccountID       string      `json:"AccountId"`
	AssetID         string      `json:"AssetId"`
	SourceID        string      `json:"SourceId"`
	ParentID        string      `json:"ParentId"`
	SuppliedName    interface{} `json:"SuppliedName"`
	SuppliedEmail   interface{} `json:"SuppliedEmail"`
	SuppliedPhone   interface{} `json:"SuppliedPhone"`
	SuppliedCompany interface{} `json:"SuppliedCompany"`
	Type            interface{} `json:"Type"`
	Reason          interface{} `json:"Reason"`
	Comments        string      `json:"Comments"`
}

//SalesforceResponse handles a generic response
type SalesforceResponse struct {
	ID      string        `json:"id"`
	Success bool          `json:"success"`
	Errors  []interface{} `json:"errors"`
}

type recordResponse struct {
	Attributes        map[string]interface{} `json:"attributes"`
	ContentDocumentID string                 `json:"ContentDocumentId"`
	Id                string                 `json:"Id"`
	FirstName         string                 `json:"FirstName"`
	LastName          string                 `json:"LastName"`
	Email             string                 `json:"Email"`
	MobilePhone       string                 `json:"MobilePhone"`
	BlockedChatYalo   bool                   `json:"CP_BlockedChatYalo__c"`
	PersonContactID   string                 `json:"PersonContactId"`
	PersonEmail       string                 `json:"PersonEmail"`
	PersonMobilePhone string                 `json:"PersonMobilePhone"`
}

//SearchResponse handles search document response
type SearchResponse struct {
	TotalSize int64            `json:"totalSize"`
	Done      bool             `json:"done"`
	Records   []recordResponse `json:"records"`
}

//ContactRequest handles search document response
type ContactRequest struct {
	FirstName   string `json:"FirstName" validate:"required"`
	LastName    string `json:"LastName" validate:"required"`
	MobilePhone string `json:"MobilePhone"`
	Email       string `json:"Email" validate:"required"`
	AccountID   string `json:"AccountId"`
}

//AccountRequest for create account in salesforce
type AccountRequest struct {
	Name              *string `json:"Name,omitempty"`
	Phone             *string `json:"Phone,omitempty"`
	PersonEmail       *string `json:"PersonEmail,omitempty" validate:"required"`
	PersonMobilePhone *string `json:"PersonMobilePhone,omitempty"`
	FirstName         *string `json:"FirstName,omitempty" validate:"required"`
	LastName          *string `json:"LastName,omitempty" validate:"required"`
	RecordTypeID      *string `json:"RecordTypeId,omitempty" validate:"required"`
	PersonBirthDate   *string `json:"PersonBirthDate,omitempty" validate:"required"`
}

//CompositeRequest struct to request compose
type CompositeRequest struct {
	AllOrNone          bool        `json:"allOrNone"`
	CollateSubrequests bool        `json:"collateSubrequests"`
	CompositeRequest   []Composite `json:"compositeRequest" validate:"required"`
}

//Composite struct to request compose
type Composite struct {
	Method      string      `json:"method" validate:"required"`
	URL         string      `json:"url" validate:"required"`
	Body        interface{} `json:"body"`
	ReferenceId string      `json:"referenceId" validate:"required"`
}

type CompositeResponse struct {
	Body           interface{} `json:"body"`
	HTTPHeaders    HTTPHeaders `json:"httpHeaders"`
	HTTPStatusCode int64       `json:"httpStatusCode"`
	ReferenceID    string      `json:"referenceId"`
}
type CompositeResponses struct {
	CompositeResponse []CompositeResponse `json:"compositeResponse"`
}

type HTTPHeaders struct {
	Location string `json:"Location"`
}

//SaleforceInterface handles all Saleforce's methods
type SaleforceInterface interface {
	CreateCase(mainSpan tracer.Span, payload interface{}) (string, *helpers.ErrorResponse)
	Search(string) (*SearchResponse, *helpers.ErrorResponse)
	SearchID(string) (string, error)
	SearchContact(string) (*models.SfcContact, *helpers.ErrorResponse)
	SearchContactComposite(mainSpan tracer.Span, email, phoneNumber string, sfcCustomFieldsToSearchContact map[string]string, extraData map[string]interface{}) (*models.SfcContact, *helpers.ErrorResponse)
	SearchAccount(string) (*models.SfcAccount, *helpers.ErrorResponse)
	//Methods related to upload and associate an image to a case
	CreateContentVersion(ContentVersionPayload) (string, error)
	SearchDocumentID(string) (string, error)
	LinkDocumentToCase(LinkDocumentPayload) (string, error)
	CreateContact(mainSpan tracer.Span, payload interface{}) (string, *helpers.ErrorResponse)
	CreateAccount(payload AccountRequest) (string, *helpers.ErrorResponse)
	CreateAccountComposite(mainSpan tracer.Span, payload interface{}) (*models.SfcAccount, *helpers.ErrorResponse)
	Composite(mainSpan tracer.Span, compositeRequest CompositeRequest) (CompositeResponses, *helpers.ErrorResponse)
	GetContentVersionURL() string
	GetSearchURL(query string) string
	GetDocumentLinkURL() string
	UpdateToken(accessToken string)
}

//CreateContentVersion creates a new content version for a file
func (cc *SalesforceClient) CreateContentVersion(contentVersionPayload ContentVersionPayload) (string, error) {
	// datadog tracing
	span := tracer.StartSpan("create_content_version")
	span.SetTag(ext.AnalyticsEvent, true)
	span.SetTag(events.Payload, fmt.Sprintf("%#v", contentVersionPayload))
	defer span.Finish()
	uri := fmt.Sprintf("/services/data/v%s.0/sobjects/ContentVersion", cc.APIVersion)
	span.SetTag(ext.ResourceName, fmt.Sprintf("%s %s", http.MethodPost, uri))
	var errorMessage string

	//validating ContentVersionPayload struct
	if err := helpers.Govalidator().Struct(contentVersionPayload); err != nil {
		errorMessage = fmt.Sprintf("%s : %s", helpers.InvalidPayload, err.Error())
		logrus.Error(errorMessage)
		span.SetTag(ext.Error, err)
		return "", errors.New(errorMessage)
	}

	//building request to send through proxy
	requestBytes, _ := json.Marshal(contentVersionPayload)

	header := make(map[string]string)
	header["Content-Type"] = "application/json"
	header["Authorization"] = fmt.Sprintf("Bearer %s", cc.AccessToken)

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
		return "", errors.New(errorMessage)
	}

	if proxiedResponse.StatusCode != http.StatusCreated {
		responseMap := []map[string]interface{}{}
		readAndUnmarshalError := helpers.ReadAndUnmarshal(proxiedResponse.Body, &responseMap)

		if readAndUnmarshalError != nil {
			errorMessage = fmt.Sprintf("%s : %s", constants.UnmarshallError, readAndUnmarshalError.Error())
			logrus.Error(errorMessage)
			span.SetTag(ext.Error, readAndUnmarshalError)
			return "", errors.New(errorMessage)
		}

		errorMessage = fmt.Sprintf("%s-[%d] : %#v", constants.StatusError, proxiedResponse.StatusCode, responseMap)
		logrus.WithFields(logrus.Fields{
			"response": responseMap,
		}).Error(errorMessage)
		err := errors.New(errorMessage)
		span.SetTag(ext.Error, err)
		return "", err
	}

	var response SalesforceResponse
	readAndUnmarshalError := helpers.ReadAndUnmarshal(proxiedResponse.Body, &response)

	if readAndUnmarshalError != nil {
		errorMessage = fmt.Sprintf("%s : %s", constants.UnmarshallError, readAndUnmarshalError.Error())
		logrus.Error(errorMessage)
		span.SetTag(ext.Error, readAndUnmarshalError)
		return "", errors.New(errorMessage)
	}

	logrus.WithFields(logrus.Fields{
		"response": response,
	}).Info("Create ContentVersion Success")

	return response.ID, nil
}

//Search for entities in salesforce
func (cc *SalesforceClient) Search(query string) (*SearchResponse, *helpers.ErrorResponse) {
	// datadog tracing
	span := tracer.StartSpan("search")
	span.SetTag(ext.AnalyticsEvent, true)
	defer span.Finish()
	uri := fmt.Sprintf("/services/data/v%s.0/query/?q=%s", cc.APIVersion, query)
	span.SetTag(ext.ResourceName, fmt.Sprintf("%s %s", http.MethodPost, uri))
	var errorMessage string

	//validating query param
	if query == "" || len(query) < 1 {
		errorMessage = fmt.Sprintf("%s : %s", constants.QueryParamError, helpers.MissingQueryParam)
		logrus.Error(errorMessage)
		errorResponse := &helpers.ErrorResponse{Error: errors.New(errorMessage), StatusCode: 0}
		span.SetTag(ext.Error, errorResponse.Error)
		return nil, errorResponse
	}

	header := make(map[string]string)
	header["Content-Type"] = "application/json"
	header["Authorization"] = fmt.Sprintf("Bearer %s", cc.AccessToken)

	newRequest := proxy.Request{
		Method:    http.MethodGet,
		URI:       uri,
		HeaderMap: header,
	}

	proxiedResponse, proxyError := cc.Proxy.SendHTTPRequest(span, &newRequest)
	if proxyError != nil {
		errorMessage = fmt.Sprintf("%s : %s", constants.ForwardError, proxyError.Error())
		logrus.Error(errorMessage)
		return nil, &helpers.ErrorResponse{Error: errors.New(errorMessage), StatusCode: 0}
	}

	if proxiedResponse.StatusCode != http.StatusOK {
		responseMap := []map[string]interface{}{}
		readAndUnmarshalError := helpers.ReadAndUnmarshal(proxiedResponse.Body, &responseMap)

		if readAndUnmarshalError != nil {
			errorMessage = fmt.Sprintf("%s : %s", constants.UnmarshallError, readAndUnmarshalError.Error())
			logrus.Error(errorMessage)
			return nil, &helpers.ErrorResponse{Error: errors.New(errorMessage), StatusCode: proxiedResponse.StatusCode}
		}

		errorMessage = fmt.Sprintf("%s : %d", constants.StatusError, proxiedResponse.StatusCode)
		logrus.WithFields(logrus.Fields{
			"response": responseMap,
		}).Error(errorMessage)
		return nil, &helpers.ErrorResponse{Error: errors.New(errorMessage), StatusCode: proxiedResponse.StatusCode}
	}

	var response SearchResponse
	readAndUnmarshalError := helpers.ReadAndUnmarshal(proxiedResponse.Body, &response)

	if readAndUnmarshalError != nil {
		errorMessage = fmt.Sprintf("%s : %s", constants.UnmarshallError, readAndUnmarshalError.Error())
		logrus.Error(errorMessage)
		return nil, &helpers.ErrorResponse{Error: errors.New(errorMessage), StatusCode: http.StatusOK}
	}

	logrus.WithFields(logrus.Fields{
		"response": response,
	}).Info("Search successfully???")

	return &response, nil
}

//SearchDocumentID looks for the DocumentID of the file created
func (cc *SalesforceClient) SearchDocumentID(query string) (string, error) {
	response, err := cc.Search(query)

	if err != nil {
		return "", err.Error
	}

	if len(response.Records) < 1 || response.Records[0].ContentDocumentID == "" {
		errorMessage := fmt.Sprintf("%s : %s", constants.RequestError, helpers.EmptyResponse)
		logrus.Error(errorMessage)
		return "", errors.New(errorMessage)
	}

	logrus.WithFields(logrus.Fields{
		"DocumentID": response.Records[0].ContentDocumentID,
	}).Info("DocumentID found successfully???")

	return response.Records[0].ContentDocumentID, nil
}

//SearchID the entity's identifier from Salesforce
func (cc *SalesforceClient) SearchID(query string) (string, error) {
	response, err := cc.Search(query)

	if err != nil {
		return "", err.Error
	}

	if len(response.Records) < 1 || response.Records[0].Id == "" {
		errorMessage := fmt.Sprintf("%s : %s", constants.RequestError, helpers.EmptyResponse)
		logrus.Error(errorMessage)
		return "", errors.New(errorMessage)
	}

	logrus.WithFields(logrus.Fields{
		"ID": response.Records[0].Id,
	}).Info("ID found successfully???")

	return response.Records[0].Id, nil
}

func (cc *SalesforceClient) SearchContact(query string) (*models.SfcContact, *helpers.ErrorResponse) {
	response, err := cc.Search(query)

	if err != nil {
		return nil, err
	}

	if len(response.Records) < 1 || response.Records[0].Id == "" {
		errorMessage := fmt.Sprintf("%s : %s", "contact not found", helpers.EmptyResponse)
		logrus.Error(errorMessage)
		return nil, &helpers.ErrorResponse{Error: errors.New(errorMessage), StatusCode: http.StatusNotFound}
	}

	contact := models.SfcContact{
		ID:          response.Records[0].Id,
		FirstName:   response.Records[0].FirstName,
		LastName:    response.Records[0].LastName,
		Email:       response.Records[0].Email,
		MobilePhone: response.Records[0].MobilePhone,
		Blocked:     response.Records[0].BlockedChatYalo,
	}

	logrus.WithFields(logrus.Fields{
		"contact": contact,
	}).Info("Contact found successfully???")

	return &contact, nil
}

func (cc *SalesforceClient) SearchContactComposite(mainSpan tracer.Span, email, phoneNumber string, sfcCustomFieldsToSearchContact map[string]string, extraData map[string]interface{}) (*models.SfcContact, *helpers.ErrorResponse) {
	spanContext := events.GetSpanContextFromSpan(mainSpan)
	span := tracer.StartSpan("SearchContactComposite", tracer.ChildOf(spanContext))
	span.SetTag(ext.AnalyticsEvent, true)
	span.SetTag("email", email)
	span.SetTag("phoneNumber", phoneNumber)
	defer span.Finish()

	blockedChat := ""
	if cc.SfcBlockedChatField {
		blockedChat = blockedChatYaloField
	}
	request := CompositeRequest{
		CompositeRequest: []Composite{
			{
				Method:      http.MethodGet,
				URL:         fmt.Sprintf("/services/data/v%s.0/query/?q=%s", cc.APIVersion, fmt.Sprintf(queryForContactByField, blockedChat, "email", email)),
				ReferenceId: "newQueryEmail",
			},
		},
	}

	if phoneNumber != "" {
		request.CompositeRequest = append(request.CompositeRequest, Composite{
			Method:      http.MethodGet,
			URL:         fmt.Sprintf("/services/data/v%s.0/query/?q=%s", cc.APIVersion, fmt.Sprintf(queryForContactByField, blockedChat, "mobilePhone", phoneNumber)),
			ReferenceId: "newQueryPhone",
		})
	}

	if sfcCustomFieldsToSearchContact != nil {
		for key, field := range sfcCustomFieldsToSearchContact {
			value, ok := extraData[key]
			if ok {
				logrus.WithFields(logrus.Fields{
					field: value,
				}).Info("Searching contact with custom field")

				reg, _ := regexp.Compile("[^a-zA-Z0-9]+")
				fieldForReference := reg.ReplaceAllString(field, "_")

				request.CompositeRequest = append(request.CompositeRequest, Composite{
					Method:      http.MethodGet,
					URL:         fmt.Sprintf("/services/data/v%s.0/query/?q=%s", cc.APIVersion, fmt.Sprintf(queryForContactByField, blockedChat, field, value)),
					ReferenceId: "newQueryCustomField_" + fieldForReference,
				})
			}
		}
	}

	compositeResponses, err := cc.Composite(span, request)
	if err != nil {
		if err.StatusCode != http.StatusUnauthorized {
			span.SetTag("unauthorized-error", err)
		} else {
			span.SetTag(ext.Error, err)
		}
		return nil, err
	}

	contact := models.SfcContact{}
	for _, compositeResponse := range compositeResponses.CompositeResponse {
		response := SearchResponse{}
		responseMap := compositeResponse.Body.(map[string]interface{})
		responseBin, _ := json.Marshal(responseMap)
		json.Unmarshal(responseBin, &response)

		if response.TotalSize == 0 {
			continue
		}

		contact = models.SfcContact{
			ID:          response.Records[0].Id,
			FirstName:   response.Records[0].FirstName,
			LastName:    response.Records[0].LastName,
			Email:       response.Records[0].Email,
			MobilePhone: response.Records[0].MobilePhone,
			Blocked:     response.Records[0].BlockedChatYalo,
		}
		break
	}

	if contact.ID == "" {
		message := fmt.Sprintf("%s : %s", "contact not found", helpers.EmptyResponse)
		logrus.Info(message)
		errorResponse := &helpers.ErrorResponse{Error: errors.New(message), StatusCode: http.StatusNotFound}
		span.SetTag(ext.ResourceName, errorResponse.Error)
		return nil, errorResponse
	}

	logrus.WithFields(logrus.Fields{
		"contact": contact,
	}).Info("Contact found successfully???")

	return &contact, nil
}

func (cc *SalesforceClient) SearchAccount(query string) (*models.SfcAccount, *helpers.ErrorResponse) {
	response, err := cc.Search(query)

	if err != nil {
		return nil, err
	}

	if len(response.Records) < 1 || response.Records[0].Id == "" {
		errorMessage := fmt.Sprintf("%s : %s", "account not found", helpers.EmptyResponse)
		logrus.Error(errorMessage)
		return nil, &helpers.ErrorResponse{Error: errors.New(errorMessage), StatusCode: http.StatusNotFound}
	}

	personAccount := models.SfcAccount{
		ID:                response.Records[0].Id,
		FirstName:         response.Records[0].FirstName,
		LastName:          response.Records[0].LastName,
		PersonEmail:       response.Records[0].PersonEmail,
		PersonMobilePhone: response.Records[0].PersonMobilePhone,
		PersonContactId:   response.Records[0].PersonContactID,
	}

	logrus.WithFields(logrus.Fields{
		"account": personAccount,
	}).Info("Account found successfully???")

	return &personAccount, nil
}

//LinkDocumentToCase associates the file added with an valid case
func (cc *SalesforceClient) LinkDocumentToCase(linkDocumentPayload LinkDocumentPayload) (string, error) {
	// datadog tracing
	span := tracer.StartSpan("link_document_to_case")
	span.SetTag(ext.AnalyticsEvent, true)
	span.SetTag(events.Payload, linkDocumentPayload)
	defer span.Finish()
	uri := fmt.Sprintf("/services/data/v%s.0/sobjects/ContentDocumentLink", cc.APIVersion)
	span.SetTag(ext.ResourceName, fmt.Sprintf("%s %s", http.MethodPost, uri))
	var errorMessage string

	logrus.WithFields(logrus.Fields{
		"payload": linkDocumentPayload,
	}).Info("LinkDocumentPayload received")

	//validating LinkDocumentPayload struct
	if err := helpers.Govalidator().Struct(linkDocumentPayload); err != nil {
		errorMessage = fmt.Sprintf("%s : %s", helpers.InvalidPayload, err.Error())
		logrus.Error(errorMessage)
		span.SetTag(ext.Error, err)
		return "", errors.New(errorMessage)
	}

	//building request to send through proxy
	requestBytes, _ := json.Marshal(linkDocumentPayload)

	header := make(map[string]string)
	header["Content-Type"] = "application/json"
	header["Authorization"] = fmt.Sprintf("Bearer %s", cc.AccessToken)

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
		return "", errors.New(errorMessage)
	}

	if proxiedResponse.StatusCode != http.StatusCreated {
		responseMap := []map[string]interface{}{}
		readAndUnmarshalError := helpers.ReadAndUnmarshal(proxiedResponse.Body, &responseMap)

		if readAndUnmarshalError != nil {
			errorMessage = fmt.Sprintf("%s : %s", constants.UnmarshallError, readAndUnmarshalError.Error())
			logrus.Error(errorMessage)
			span.SetTag(ext.Error, readAndUnmarshalError)
			return "", errors.New(errorMessage)
		}

		errorMessage = fmt.Sprintf("%s-[%d] : %#v", constants.StatusError, proxiedResponse.StatusCode, responseMap)
		logrus.WithFields(logrus.Fields{
			"response": responseMap,
		}).Error(errorMessage)
		span.SetTag(ext.Error, errors.New(fmt.Sprintf("[%d] - %#v", proxiedResponse.StatusCode, responseMap)))
		return "", errors.New(errorMessage)
	}

	var response SalesforceResponse
	readAndUnmarshalError := helpers.ReadAndUnmarshal(proxiedResponse.Body, &response)

	if readAndUnmarshalError != nil {
		errorMessage = fmt.Sprintf("%s : %s", constants.UnmarshallError, readAndUnmarshalError.Error())
		logrus.Error(errorMessage)
		span.SetTag(ext.Error, readAndUnmarshalError)
		return "", errors.New(errorMessage)
	}

	logrus.WithFields(logrus.Fields{
		"response": response,
	}).Info("Linked Document Success")

	return response.ID, nil
}

//CreateCase Create case for Salesforce Requests
func (cc *SalesforceClient) CreateCase(mainSpan tracer.Span, payload interface{}) (string, *helpers.ErrorResponse) {
	// datadog tracing
	spanContext := events.GetSpanContextFromSpan(mainSpan)
	span := tracer.StartSpan("create_case", tracer.ChildOf(spanContext))
	span.SetTag(ext.AnalyticsEvent, true)
	span.SetTag("payload", payload)
	defer span.Finish()
	var errorMessage string

	logrus.WithFields(logrus.Fields{
		"payload": payload,
	}).Info("Payload received")

	//building request to send through proxy
	requestBytes, _ := json.Marshal(payload)

	header := make(map[string]string)
	header["Content-Type"] = "application/json"
	header["Authorization"] = fmt.Sprintf("Bearer %s", cc.AccessToken)
	header[AutoAsssingHeader] = "false"

	newRequest := proxy.Request{
		Body:      requestBytes,
		Method:    http.MethodPost,
		URI:       fmt.Sprintf("/services/data/v%s.0/sobjects/Case", cc.APIVersion),
		HeaderMap: header,
	}
	span.SetTag(ext.ResourceName, fmt.Sprintf("%s %s", newRequest.Method, newRequest.URI))

	proxiedResponse, proxyError := cc.Proxy.SendHTTPRequest(span, &newRequest)
	if proxyError != nil {
		errorMessage = fmt.Sprintf("%s : %s", constants.ForwardError, proxyError.Error())
		logrus.Error(errorMessage)
		span.SetTag(ext.Error, proxyError)
		return "", &helpers.ErrorResponse{Error: errors.New(errorMessage), StatusCode: 0}
	}

	if proxiedResponse.StatusCode != http.StatusCreated {
		errorResponse := helpers.GetErrorResponseArrayMap(proxiedResponse.Body, constants.StatusError, proxiedResponse.StatusCode)
		span.SetTag(ext.Error, errorResponse.Error)
		return "", errorResponse
	}

	var response SalesforceResponse
	readAndUnmarshalError := helpers.ReadAndUnmarshal(proxiedResponse.Body, &response)
	if readAndUnmarshalError != nil {
		errorMessage = fmt.Sprintf("%s : %s", constants.UnmarshallError, readAndUnmarshalError.Error())
		logrus.Error(errorMessage)
		span.SetTag(ext.Error, readAndUnmarshalError)
		return "", &helpers.ErrorResponse{Error: errors.New(errorMessage), StatusCode: http.StatusCreated}
	}

	logrus.WithFields(logrus.Fields{
		"response": response,
	}).Info("Create case success")

	return response.ID, nil
}

//CreateContact Create contact for Salesforce Requests
func (cc *SalesforceClient) CreateContact(mainSpan tracer.Span, payload interface{}) (string, *helpers.ErrorResponse) {
	// datadog tracing
	spanContext := events.GetSpanContextFromSpan(mainSpan)
	span := tracer.StartSpan("create_contact", tracer.ChildOf(spanContext))
	span.SetTag(ext.AnalyticsEvent, true)
	span.SetTag(events.Payload, fmt.Sprintf("%#v", payload))
	defer span.Finish()
	uri := fmt.Sprintf("/services/data/v%s.0/sobjects/Contact", cc.APIVersion)
	span.SetTag(ext.ResourceName, fmt.Sprintf("%s %s", http.MethodPost, uri))
	var errorMessage string

	logrus.WithFields(logrus.Fields{
		"payload": payload,
	}).Info("Payload received")

	//building request to send through proxy
	requestBytes, _ := json.Marshal(payload)

	header := make(map[string]string)
	header["Content-Type"] = "application/json"
	header["Authorization"] = fmt.Sprintf("Bearer %s", cc.AccessToken)

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
		return "", &helpers.ErrorResponse{Error: errors.New(errorMessage), StatusCode: 0}
	}

	if proxiedResponse.StatusCode != http.StatusCreated {
		errorResponse := helpers.GetErrorResponse(proxiedResponse.Body, constants.StatusError, proxiedResponse.StatusCode)
		span.SetTag(ext.Error, errorResponse.Error)
		return "", errorResponse
	}

	var response SalesforceResponse
	readAndUnmarshalError := helpers.ReadAndUnmarshal(proxiedResponse.Body, &response)
	if readAndUnmarshalError != nil {
		errorMessage = fmt.Sprintf("%s : %s", constants.UnmarshallError, readAndUnmarshalError.Error())
		logrus.Error(errorMessage)
		span.SetTag(ext.Error, readAndUnmarshalError)
		return "", &helpers.ErrorResponse{Error: errors.New(errorMessage), StatusCode: proxiedResponse.StatusCode}

	}

	logrus.WithFields(logrus.Fields{
		"response": response,
	}).Info("create contact success")

	return response.ID, nil
}

//CreateAccount Create account for Salesforce Requests
func (cc *SalesforceClient) CreateAccount(payload AccountRequest) (string, *helpers.ErrorResponse) {
	// datadog tracing
	span := tracer.StartSpan("create_account")
	span.SetTag(ext.AnalyticsEvent, true)
	span.SetTag(events.Payload, fmt.Sprintf("%#v", payload))
	defer span.Finish()
	uri := fmt.Sprintf("/services/data/v%s.0/sobjects/Account", cc.APIVersion)
	span.SetTag(ext.ResourceName, fmt.Sprintf("%s %s", http.MethodPost, uri))
	var errorMessage string

	logrus.WithFields(logrus.Fields{
		"payload": payload,
	}).Info("Payload received")

	//validating AccountRequest Payload struct
	if err := helpers.Govalidator().Struct(payload); err != nil {
		errorMessage = fmt.Sprintf("%s : %s", helpers.InvalidPayload, err.Error())
		logrus.Error(errorMessage)
		span.SetTag(ext.Error, err)
		return "", &helpers.ErrorResponse{Error: errors.New(errorMessage), StatusCode: http.StatusBadRequest}
	}

	//building request to send through proxy
	requestBytes, _ := helpers.MarshalJSON(payload)

	header := make(map[string]string)
	header["Content-Type"] = "application/json"
	header["Authorization"] = fmt.Sprintf("Bearer %s", cc.AccessToken)

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
		return "", &helpers.ErrorResponse{Error: errors.New(errorMessage), StatusCode: 0}
	}

	if proxiedResponse.StatusCode != http.StatusCreated {
		errorResponse := helpers.GetErrorResponseArrayMap(proxiedResponse.Body, constants.StatusError, proxiedResponse.StatusCode)
		span.SetTag(ext.Error, errorResponse.Error)
		return "", errorResponse
	}

	var response SalesforceResponse
	readAndUnmarshalError := helpers.ReadAndUnmarshal(proxiedResponse.Body, &response)
	if readAndUnmarshalError != nil {
		errorMessage = fmt.Sprintf("%s : %s", constants.UnmarshallError, readAndUnmarshalError.Error())
		logrus.Error(errorMessage)
		span.SetTag(ext.Error, readAndUnmarshalError)
		return "", &helpers.ErrorResponse{Error: errors.New(errorMessage), StatusCode: proxiedResponse.StatusCode}
	}

	logrus.WithFields(logrus.Fields{
		"response": response,
	}).Info("create account success")

	return response.ID, nil
}

func (cc *SalesforceClient) CreateAccountComposite(mainSpan tracer.Span, payload interface{}) (*models.SfcAccount, *helpers.ErrorResponse) {
	spanContext := events.GetSpanContextFromSpan(mainSpan)
	span := tracer.StartSpan("CreateAccountComposite", tracer.ChildOf(spanContext))
	span.SetTag(ext.AnalyticsEvent, true)
	defer span.Finish()
	request := CompositeRequest{
		CompositeRequest: []Composite{
			{
				Method:      http.MethodPost,
				URL:         fmt.Sprintf("/services/data/v%s.0/sobjects/Account", cc.APIVersion),
				Body:        payload,
				ReferenceId: "newAccount",
			},
			{
				Method:      http.MethodGet,
				URL:         fmt.Sprintf("/services/data/v%s.0/query/?q=%s", cc.APIVersion, fmt.Sprintf(queryForAccountByField, "id", "@{newAccount.id}")),
				ReferenceId: "newQuery",
			},
		},
	}

	compositeResponses, err := cc.Composite(span, request)
	if err != nil {
		span.SetTag(ext.Error, err)
		return nil, err
	}

	createAccountRequest := compositeResponses.CompositeResponse[0]
	if createAccountRequest.HTTPStatusCode >= http.StatusBadRequest {
		errorMessage := "Could not create the account"
		span.SetTag("salesforce-response", fmt.Sprintf("%+v", compositeResponses.CompositeResponse))
		span.SetTag(ext.Error, errorMessage)
		return nil, &helpers.ErrorResponse{Error: errors.New(errorMessage), StatusCode: int(createAccountRequest.HTTPStatusCode)}
	}

	response := SearchResponse{}
	responseMap := compositeResponses.CompositeResponse[1].Body.(map[string]interface{})
	responseBin, _ := json.Marshal(responseMap)
	json.Unmarshal(responseBin, &response)

	if response.TotalSize == 0 {
		errorMessage := fmt.Sprintf("%s : %s", "account not found", helpers.EmptyResponse)
		logrus.Error(errorMessage)
		errorResponse := &helpers.ErrorResponse{Error: errors.New(errorMessage), StatusCode: http.StatusNotFound}
		span.SetTag(ext.Error, errorResponse.Error)
		return nil, errorResponse
	}

	personAccount := models.SfcAccount{
		ID:                response.Records[0].Id,
		FirstName:         response.Records[0].FirstName,
		LastName:          response.Records[0].LastName,
		PersonEmail:       response.Records[0].PersonEmail,
		PersonMobilePhone: response.Records[0].PersonMobilePhone,
		PersonContactId:   response.Records[0].PersonContactID,
	}

	logrus.WithFields(logrus.Fields{
		"account": personAccount,
	}).Info("Account found successfully???")

	return &personAccount, nil
}

//Composite create a composite request
func (cc *SalesforceClient) Composite(mainSpan tracer.Span, compositeRequest CompositeRequest) (CompositeResponses, *helpers.ErrorResponse) {
	// datadog tracing
	spanContext := events.GetSpanContextFromSpan(mainSpan)
	span := tracer.StartSpan("composite", tracer.ChildOf(spanContext))
	span.SetTag(ext.AnalyticsEvent, true)
	span.SetTag(events.Payload, fmt.Sprintf("%#v", compositeRequest))
	defer span.Finish()
	uri := fmt.Sprintf("/services/data/v%s.0/composite", cc.APIVersion)
	span.SetTag(ext.ResourceName, fmt.Sprintf("%s %s", http.MethodPost, uri))
	var errorMessage string

	//validating CompositeRequest struct
	if err := helpers.Govalidator().Struct(compositeRequest); err != nil {
		errorMessage = fmt.Sprintf("%s : %s", helpers.InvalidPayload, err.Error())
		logrus.Error(errorMessage)
		span.SetTag(ext.Error, err)
		return CompositeResponses{}, &helpers.ErrorResponse{Error: errors.New(errorMessage), StatusCode: http.StatusBadRequest}
	}

	//building request to send through proxy
	requestBytes, _ := json.Marshal(compositeRequest)

	header := make(map[string]string)
	header["Content-Type"] = "application/json"
	header["Authorization"] = fmt.Sprintf("Bearer %s", cc.AccessToken)

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
		return CompositeResponses{}, &helpers.ErrorResponse{Error: errors.New(errorMessage), StatusCode: http.StatusNotFound}
	}

	if proxiedResponse.StatusCode != http.StatusOK {
		return CompositeResponses{}, helpers.GetErrorResponseArrayMap(proxiedResponse.Body, constants.StatusError, proxiedResponse.StatusCode)
	}

	response := CompositeResponses{}
	readAndUnmarshalError := helpers.ReadAndUnmarshal(proxiedResponse.Body, &response)

	if readAndUnmarshalError != nil {
		errorMessage = fmt.Sprintf("%s : %s", constants.UnmarshallError, readAndUnmarshalError.Error())
		logrus.Error(errorMessage)
		return CompositeResponses{}, &helpers.ErrorResponse{Error: errors.New(errorMessage), StatusCode: http.StatusInternalServerError}
	}

	logrus.WithFields(logrus.Fields{
		"response": response,
	}).Info("Create composite success")

	return response, nil
}

func (cc *SalesforceClient) GetContentVersionURL() string {
	return fmt.Sprintf("/services/data/v%s.0/sobjects/ContentVersion", cc.APIVersion)
}
func (cc *SalesforceClient) GetSearchURL(query string) string {
	return fmt.Sprintf("/services/data/v%s.0/query/?q=%s", cc.APIVersion, query)
}

func (cc *SalesforceClient) GetDocumentLinkURL() string {
	return fmt.Sprintf("/services/data/v%s.0/sobjects/ContentDocumentLink", cc.APIVersion)
}

func (cc *SalesforceClient) UpdateToken(accessToken string) {
	cc.AccessToken = accessToken
}
