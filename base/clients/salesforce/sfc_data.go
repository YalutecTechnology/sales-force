package salesforce

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"yalochat.com/salesforce-integration/base/clients/proxy"
	"yalochat.com/salesforce-integration/base/constants"
	"yalochat.com/salesforce-integration/base/helpers"
	"yalochat.com/salesforce-integration/base/models"
)

func NewSalesforceRequester(url, token string) *SalesforceClient {
	return &SalesforceClient{
		Proxy:       proxy.NewProxy(url),
		AccessToken: token,
	}
}

//SalesforceClient settings for use a case client
type SalesforceClient struct {
	AccessToken string
	APIVersion  string
	Proxy       proxy.ProxyInterface
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
	ContactID             string      `json:"ContactId" validate:"required"`
	Description           string      `json:"Description" validate:"required"`
	Origin                string      `json:"Origin" validate:"required"`
	Priority              string      `json:"Priority" validate:"required"`
	Status                string      `json:"Status" validate:"required"`
	Subject               string      `json:"Subject" validate:"required"`
	IsEscalated           bool        `json:"IsEscalated"`
	AccountID             interface{} `json:"AccountId"`
	AssetID               interface{} `json:"AssetId"`
	SourceID              interface{} `json:"SourceId"`
	ParentID              interface{} `json:"ParentId"`
	SuppliedName          interface{} `json:"SuppliedName"`
	SuppliedEmail         interface{} `json:"SuppliedEmail"`
	SuppliedPhone         interface{} `json:"SuppliedPhone"`
	SuppliedCompany       interface{} `json:"SuppliedCompany"`
	Type                  interface{} `json:"Type"`
	Reason                interface{} `json:"Reason"`
	Comments              interface{} `json:"Comments"`
	EngineeringReqNumberC interface{} `json:"EngineeringReqNumber__c"`
	SLAViolationC         interface{} `json:"SLAViolation__c"`
	ProductC              interface{} `json:"Product__c"`
	PotentialLiabilityC   interface{} `json:"PotentialLiability__c"`
}

//SalesforceResponse handles a generic response
type SalesforceResponse struct {
	ID      string `json:"id"`
	Success bool   `json:"success"`
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
}

//SaleforceInterface handles all Saleforce's methods
type SaleforceInterface interface {
	CreateCase(payload CaseRequest) (string, error)
	Search(string) (*SearchResponse, error)
	SearchID(string) (string, error)
	SearchContact(string) (*models.SfcContact, error)
	//Methods related to upload and associate an image to a case
	CreateContentVersion(ContentVersionPayload) (string, error)
	SearchDocumentID(string) (string, error)
	LinkDocumentToCase(LinkDocumentPayload) (string, error)
	CreateContact(payload ContactRequest) (string, error)
}

//CreateContentVersion creates a new content version for a file
func (cc *SalesforceClient) CreateContentVersion(contentVersionPayload ContentVersionPayload) (string, error) {
	var errorMessage string

	logrus.WithFields(logrus.Fields{
		"payload": contentVersionPayload,
	}).Info("ContentVersionPayload received")

	//validating ContentVersionPayload struct
	if err := helpers.Govalidator().Struct(contentVersionPayload); err != nil {
		errorMessage = fmt.Sprintf("%s : %s", helpers.InvalidPayload, err.Error())
		logrus.Error(errorMessage)
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
		URI:       fmt.Sprintf("/services/data/v%s.0/sobjects/ContentVersion", cc.APIVersion),
		HeaderMap: header,
	}

	proxiedResponse, proxyError := cc.Proxy.SendHTTPRequest(&newRequest)
	if proxyError != nil {
		errorMessage = fmt.Sprintf("%s : %s", constants.ForwardError, proxyError.Error())
		logrus.Error(errorMessage)
		return "", errors.New(errorMessage)
	}

	if proxiedResponse.StatusCode != http.StatusCreated {
		responseMap := []map[string]interface{}{}
		readAndUnmarshalError := helpers.ReadAndUnmarshal(proxiedResponse.Body, &responseMap)

		if readAndUnmarshalError != nil {
			errorMessage = fmt.Sprintf("%s : %s", constants.UnmarshallError, readAndUnmarshalError.Error())
			logrus.Error(errorMessage)
			return "", errors.New(errorMessage)
		}

		errorMessage = fmt.Sprintf("%s : %d", constants.StatusError, proxiedResponse.StatusCode)
		logrus.WithFields(logrus.Fields{
			"response": responseMap,
		}).Error(errorMessage)
		return "", errors.New(errorMessage)
	}

	var response SalesforceResponse
	readAndUnmarshalError := helpers.ReadAndUnmarshal(proxiedResponse.Body, &response)

	if readAndUnmarshalError != nil {
		errorMessage = fmt.Sprintf("%s : %s", constants.UnmarshallError, readAndUnmarshalError.Error())
		logrus.Error(errorMessage)
		return "", errors.New(errorMessage)
	}

	logrus.WithFields(logrus.Fields{
		"response": response,
	}).Info("Create ContentVersion Success")

	return response.ID, nil
}

//Search for entities in salesforce
func (cc *SalesforceClient) Search(query string) (*SearchResponse, error) {
	var errorMessage string

	//validating query param
	if query == "" || len(query) < 1 {
		errorMessage = fmt.Sprintf("%s : %s", constants.QueryParamError, helpers.MissingQueryParam)
		logrus.Error(errorMessage)
		return nil, errors.New(errorMessage)
	}

	header := make(map[string]string)
	header["Content-Type"] = "application/json"
	header["Authorization"] = fmt.Sprintf("Bearer %s", cc.AccessToken)

	newRequest := proxy.Request{
		Method:    http.MethodGet,
		URI:       fmt.Sprintf("/services/data/v%s.0/query/?q=%s", cc.APIVersion, query),
		HeaderMap: header,
	}

	proxiedResponse, proxyError := cc.Proxy.SendHTTPRequest(&newRequest)
	if proxyError != nil {
		errorMessage = fmt.Sprintf("%s : %s", constants.ForwardError, proxyError.Error())
		logrus.Error(errorMessage)
		return nil, errors.New(errorMessage)
	}

	if proxiedResponse.StatusCode != http.StatusOK {
		responseMap := []map[string]interface{}{}
		readAndUnmarshalError := helpers.ReadAndUnmarshal(proxiedResponse.Body, &responseMap)

		if readAndUnmarshalError != nil {
			errorMessage = fmt.Sprintf("%s : %s", constants.UnmarshallError, readAndUnmarshalError.Error())
			logrus.Error(errorMessage)
			return nil, errors.New(errorMessage)
		}

		errorMessage = fmt.Sprintf("%s : %d", constants.StatusError, proxiedResponse.StatusCode)
		logrus.WithFields(logrus.Fields{
			"response": responseMap,
		}).Error(errorMessage)
		return nil, errors.New(errorMessage)
	}

	var response SearchResponse
	readAndUnmarshalError := helpers.ReadAndUnmarshal(proxiedResponse.Body, &response)

	if readAndUnmarshalError != nil {
		errorMessage = fmt.Sprintf("%s : %s", constants.UnmarshallError, readAndUnmarshalError.Error())
		logrus.Error(errorMessage)
		return nil, errors.New(errorMessage)
	}

	logrus.WithFields(logrus.Fields{
		"response": response,
	}).Info("Search successfully“")

	return &response, nil
}

//SearchDocumentID looks for the DocumentID of the file created
func (cc *SalesforceClient) SearchDocumentID(query string) (string, error) {
	response, err := cc.Search(query)

	if err != nil {
		return "", err
	}

	if len(response.Records) < 1 || response.Records[0].ContentDocumentID == "" {
		errorMessage := fmt.Sprintf("%s : %s", constants.RequestError, helpers.EmptyResponse)
		logrus.Error(errorMessage)
		return "", errors.New(errorMessage)
	}

	logrus.WithFields(logrus.Fields{
		"DocumentID": response.Records[0].ContentDocumentID,
	}).Info("DocumentID found successfully“")

	return response.Records[0].ContentDocumentID, nil
}

//Search the entity's identifier from Salesforce
func (cc *SalesforceClient) SearchID(query string) (string, error) {
	response, err := cc.Search(query)

	if err != nil {
		return "", err
	}

	if len(response.Records) < 1 || response.Records[0].Id == "" {
		errorMessage := fmt.Sprintf("%s : %s", constants.RequestError, helpers.EmptyResponse)
		logrus.Error(errorMessage)
		return "", errors.New(errorMessage)
	}

	logrus.WithFields(logrus.Fields{
		"ID": response.Records[0].Id,
	}).Info("ID found successfully“")

	return response.Records[0].Id, nil
}

func (cc *SalesforceClient) SearchContact(query string) (*models.SfcContact, error) {
	response, err := cc.Search(query)

	if err != nil {
		return nil, err
	}

	if len(response.Records) < 1 || response.Records[0].Id == "" {
		errorMessage := fmt.Sprintf("%s : %s", constants.RequestError, helpers.EmptyResponse)
		logrus.Error(errorMessage)
		return nil, errors.New(errorMessage)
	}

	contact := models.SfcContact{
		Id:          response.Records[0].Id,
		FirstName:   response.Records[0].FirstName,
		LastName:    response.Records[0].LastName,
		Email:       response.Records[0].Email,
		MobilePhone: response.Records[0].MobilePhone,
		Blocked:     response.Records[0].BlockedChatYalo,
	}

	logrus.WithFields(logrus.Fields{
		"contact": contact,
	}).Info("Contact found successfully“")

	return &contact, nil
}

//LinkDocumentToCase associates the file added with an valid case
func (cc *SalesforceClient) LinkDocumentToCase(linkDocumentPayload LinkDocumentPayload) (string, error) {
	var errorMessage string

	logrus.WithFields(logrus.Fields{
		"payload": linkDocumentPayload,
	}).Info("LinkDocumentPayload received")

	//validating LinkDocumentPayload struct
	if err := helpers.Govalidator().Struct(linkDocumentPayload); err != nil {
		errorMessage = fmt.Sprintf("%s : %s", helpers.InvalidPayload, err.Error())
		logrus.Error(errorMessage)
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
		URI:       fmt.Sprintf("/services/data/v%s.0/sobjects/ContentDocumentLink", cc.APIVersion),
		HeaderMap: header,
	}

	proxiedResponse, proxyError := cc.Proxy.SendHTTPRequest(&newRequest)
	if proxyError != nil {
		errorMessage = fmt.Sprintf("%s : %s", constants.ForwardError, proxyError.Error())
		logrus.Error(errorMessage)
		return "", errors.New(errorMessage)
	}

	if proxiedResponse.StatusCode != http.StatusCreated {
		responseMap := []map[string]interface{}{}
		readAndUnmarshalError := helpers.ReadAndUnmarshal(proxiedResponse.Body, &responseMap)

		if readAndUnmarshalError != nil {
			errorMessage = fmt.Sprintf("%s : %s", constants.UnmarshallError, readAndUnmarshalError.Error())
			logrus.Error(errorMessage)
			return "", errors.New(errorMessage)
		}

		errorMessage = fmt.Sprintf("%s : %d", constants.StatusError, proxiedResponse.StatusCode)
		logrus.WithFields(logrus.Fields{
			"response": responseMap,
		}).Error(errorMessage)
		return "", errors.New(errorMessage)
	}

	var response SalesforceResponse
	readAndUnmarshalError := helpers.ReadAndUnmarshal(proxiedResponse.Body, &response)

	if readAndUnmarshalError != nil {
		errorMessage = fmt.Sprintf("%s : %s", constants.UnmarshallError, readAndUnmarshalError.Error())
		logrus.Error(errorMessage)
		return "", errors.New(errorMessage)
	}

	logrus.WithFields(logrus.Fields{
		"response": response,
	}).Info("Linked Document Success")

	return response.ID, nil
}

//CreateCase Create case for Salesforce Requests
func (cc *SalesforceClient) CreateCase(payload CaseRequest) (string, error) {
	var errorMessage string

	logrus.WithFields(logrus.Fields{
		"payload": payload,
	}).Info("Payload received")

	//validating CaseRequest Payload struct
	if err := helpers.Govalidator().Struct(payload); err != nil {
		errorMessage = fmt.Sprintf("%s : %s", helpers.InvalidPayload, err.Error())
		logrus.Error(errorMessage)
		return "", errors.New(errorMessage)
	}

	//building request to send through proxy
	requestBytes, _ := json.Marshal(payload)

	header := make(map[string]string)
	header["Content-Type"] = "application/json"
	header["Authorization"] = fmt.Sprintf("Bearer %s", cc.AccessToken)

	newRequest := proxy.Request{
		Body:      requestBytes,
		Method:    http.MethodPost,
		URI:       fmt.Sprintf("/services/data/v%s.0/sobjects/Contact", cc.APIVersion),
		HeaderMap: header,
	}

	proxiedResponse, proxyError := cc.Proxy.SendHTTPRequest(&newRequest)
	if proxyError != nil {
		errorMessage = fmt.Sprintf("%s : %s", constants.ForwardError, proxyError.Error())
		logrus.Error(errorMessage)
		return "", errors.New(errorMessage)
	}

	var response SalesforceResponse
	readAndUnmarshalError := helpers.ReadAndUnmarshal(proxiedResponse.Body, &response)
	if readAndUnmarshalError != nil {
		errorMessage = fmt.Sprintf("%s : %s", constants.UnmarshallError, readAndUnmarshalError.Error())
		logrus.Error(errorMessage)
		return "", errors.New(errorMessage)

	}

	if proxiedResponse.StatusCode != 200 {
		return "", helpers.ErrorResponseMap(proxiedResponse.Body, constants.StatusError, proxiedResponse.StatusCode)
	}

	logrus.WithFields(logrus.Fields{
		"response": response,
	}).Info("Create case success")

	return response.ID, nil
}

//CreateContact Create contact for Salesforce Requests
func (cc *SalesforceClient) CreateContact(payload ContactRequest) (string, error) {
	var errorMessage string

	logrus.WithFields(logrus.Fields{
		"payload": payload,
	}).Info("Payload received")

	//validating ContactRequest Payload struct
	if err := helpers.Govalidator().Struct(payload); err != nil {
		errorMessage = fmt.Sprintf("%s : %s", helpers.InvalidPayload, err.Error())
		logrus.Error(errorMessage)
		return "", errors.New(errorMessage)
	}

	//building request to send through proxy
	requestBytes, _ := json.Marshal(payload)

	header := make(map[string]string)
	header["Content-Type"] = "application/json"
	header["Authorization"] = fmt.Sprintf("Bearer %s", cc.AccessToken)

	newRequest := proxy.Request{
		Body:      requestBytes,
		Method:    http.MethodPost,
		URI:       fmt.Sprintf("/services/data/v%s.0/sobjects/Contact", cc.APIVersion),
		HeaderMap: header,
	}

	proxiedResponse, proxyError := cc.Proxy.SendHTTPRequest(&newRequest)
	if proxyError != nil {
		errorMessage = fmt.Sprintf("%s : %s", constants.ForwardError, proxyError.Error())
		logrus.Error(errorMessage)
		return "", errors.New(errorMessage)
	}

	var response SalesforceResponse
	readAndUnmarshalError := helpers.ReadAndUnmarshal(proxiedResponse.Body, &response)
	if readAndUnmarshalError != nil {
		errorMessage = fmt.Sprintf("%s : %s", constants.UnmarshallError, readAndUnmarshalError.Error())
		logrus.Error(errorMessage)
		return "", errors.New(errorMessage)

	}

	if proxiedResponse.StatusCode != http.StatusCreated {
		return "", helpers.ErrorResponseMap(proxiedResponse.Body, constants.StatusError, proxiedResponse.StatusCode)
	}

	logrus.WithFields(logrus.Fields{
		"response": response,
	}).Info("create contact success")

	return response.ID, nil
}
