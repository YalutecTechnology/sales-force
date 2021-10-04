package services

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"yalochat.com/salesforce-integration/app/config/envs"
	"yalochat.com/salesforce-integration/base/clients/chat"
	"yalochat.com/salesforce-integration/base/clients/login"
	"yalochat.com/salesforce-integration/base/clients/salesforce"
	"yalochat.com/salesforce-integration/base/constants"
	"yalochat.com/salesforce-integration/base/helpers"
	"yalochat.com/salesforce-integration/base/models"
)

const (
	queryForContactByField     = `SELECT+id+,+firstName+,+lastName+,+mobilePhone+,+email+,+CP_BlockedChatYalo__c+FROM+Contact+WHERE+%s+=+` + "%s"
	queryForAccountByField     = `SELECT+id+,+firstName+,+lastName+,+PersonMobilePhone+,+PersonEmail+,+PersonContactId+FROM+Account+WHERE+%s+=+'%s'`
	firstNameDefault           = "Contacto Bot - "
	contentLocation            = "S"
	shareType                  = "V"
	visibility                 = "allUsers"
	queryContentDocumentIDByID = `SELECT+ContentDocumentID+FROM+ContentVersion+WHERE+id+=+'@{newContentVersion.id}'`
	linkReferenceID            = "@{newQuery.records[0].ContentDocumentId}"
)

type SalesforceService struct {
	TokenPayload   login.TokenPayload
	SfcLoginClient login.SfcLoginInterface
	SfcChatClient  chat.SfcChatInterface
	SfcClient      salesforce.SaleforceInterface
	SourceFlowBot  envs.SfcSourceFlowBot
	CustomFields   map[string]string
}

type SalesforceServiceInterface interface {
	CreatChat(contactName, organizationID, deploymentID, buttonID, caseID, contactID string) (*chat.SessionResponse, error)
	GetOrCreateContact(name, email, phoneNumber, accountRecordTypeID string) (*models.SfcContact, error)
	SendMessage(string, string, chat.MessagePayload) (bool, error)
	GetMessages(affinityToken, sessionKey string) (*chat.MessagesResponse, *helpers.ErrorResponse)
	CreatCase(recordType, contactID, description, subject, origin, ownerID string, extraData map[string]interface{}) (string, error)
	InsertImageInCase(uri, title, mimeType, caseID string) error
	EndChat(affinityToken, sessionKey string) error
}

func NewSalesforceService(loginClient login.SfcLoginClient, chatClient chat.SfcChatClient, salesforceClient salesforce.SalesforceClient, tokenPayload login.TokenPayload, customFields map[string]string) *SalesforceService {
	salesforceService := &SalesforceService{
		SfcLoginClient: &loginClient,
		SfcChatClient:  &chatClient,
		SfcClient:      &salesforceClient,
		TokenPayload:   tokenPayload,
		CustomFields:   customFields,
	}
	salesforceService.RefreshToken()
	return salesforceService
}

func NewCaseRequest(recordTypeID, contactID, subject, description, origin, ownerID string) *salesforce.CaseRequest {
	return &salesforce.CaseRequest{
		RecordTypeID: recordTypeID,
		ContactID:    contactID,
		Subject:      subject,
		Description:  description,
		OwnerID:      ownerID,
		Origin:       origin,
		Status:       "Nuevo",
		Priority:     "Medium",
	}
}

func (s *SalesforceService) CreatChat(contactName, organizationID, deploymentID, buttonID, caseID, contactID string) (*chat.SessionResponse, error) {
	session, err := s.SfcChatClient.CreateSession()
	if err != nil {
		return nil, err
	}

	chatRequest := chat.NewChatRequest(organizationID, deploymentID, session.Id, buttonID, contactName)

	if caseID != "" {
		caseDetail := chat.PreChatDetailsObject{
			Label:            "CaseId",
			Value:            caseID,
			TranscriptFields: []string{"CaseId"},
			DisplayToAgent:   true,
		}

		casePrechatEntitie := chat.PrechatEntitiesObject{
			EntityName:        "Case",
			ShowOnCreate:      true,
			LinkToEntityName:  "Case",
			LinkToEntityField: "Id",
			SaveToTranscript:  "Case",
			EntityFieldsMaps: []chat.EntityField{
				{
					FieldName:    "Id",
					Label:        "CaseId",
					DoFind:       true,
					IsExactMatch: true,
					DoCreate:     false,
				},
			},
		}

		chatRequest.PrechatDetails = append(chatRequest.PrechatDetails, caseDetail)
		chatRequest.PrechatEntities = append(chatRequest.PrechatEntities, casePrechatEntitie)
	}

	if contactID != "" {
		contactDetail := chat.PreChatDetailsObject{
			Label:            "ContactId",
			Value:            contactID,
			TranscriptFields: []string{"ContactId"},
			DisplayToAgent:   true,
		}

		contactPrechatEntitie := chat.PrechatEntitiesObject{
			EntityName:        "Contact",
			ShowOnCreate:      true,
			LinkToEntityName:  "Contact",
			LinkToEntityField: "Id",
			SaveToTranscript:  "Contact",
			EntityFieldsMaps: []chat.EntityField{
				{
					FieldName:    "Id",
					Label:        "ContactId",
					DoFind:       true,
					IsExactMatch: true,
					DoCreate:     false,
				},
			},
		}

		chatRequest.PrechatDetails = append(chatRequest.PrechatDetails, contactDetail)
		chatRequest.PrechatEntities = append(chatRequest.PrechatEntities, contactPrechatEntitie)
	}

	_, err = s.SfcChatClient.CreateChat(session.AffinityToken, session.Key, chatRequest)
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (s *SalesforceService) GetOrCreateContact(name, email, phoneNumber, accountRecordTypeID string) (*models.SfcContact, error) {
	// Search contact by email
	contact, err := s.SfcClient.SearchContact(fmt.Sprintf(queryForContactByField, "email", "%27"+email+"%27"))

	if err != nil {
		logrus.Infof("Not found contact by email : [%s]-[%s]", email, err.Error.Error())
		if err.StatusCode == http.StatusUnauthorized {
			s.RefreshToken()
		}
	} else {
		return contact, nil
	}
	// Search contact by phone
	if phoneNumber != "" {
		contact, err = s.SfcClient.SearchContact(fmt.Sprintf(queryForContactByField, "mobilePhone", "%27"+phoneNumber+"%27"))
		if err != nil {
			logrus.Infof("Not found contact by mobile phone : [%s]-[%s]", phoneNumber, err.Error.Error())
			if err.StatusCode == http.StatusUnauthorized {
				s.RefreshToken()
			}
		} else {
			return contact, nil
		}
	}

	contact = &models.SfcContact{
		FirstName:   firstNameDefault,
		LastName:    name,
		Email:       email,
		MobilePhone: phoneNumber,
	}

	if accountRecordTypeID != "" {
		firstName := firstNameDefault
		date := time.Now().Format(constants.DateFormatDateTime)
		accountID, err := s.SfcClient.CreateAccount(salesforce.AccountRequest{
			FirstName:         &firstName,
			LastName:          &name,
			PersonEmail:       &email,
			PersonMobilePhone: &phoneNumber,
			PersonBirthDate:   &date,
			RecordTypeID:      &accountRecordTypeID,
		})

		if err != nil {
			if err.StatusCode == http.StatusUnauthorized {
				s.RefreshToken()
			}
			return nil, errors.New(helpers.ErrorMessage("not create account", err.Error))
		}

		account, err := s.SfcClient.SearchAccount(fmt.Sprintf(queryForAccountByField, "id", accountID))

		if err != nil {
			logrus.Infof("Not found account by id : [%s]-[%s]", account.ID, err.Error.Error())
			if err.StatusCode == http.StatusUnauthorized {
				s.RefreshToken()
			}
		}
		contact.ID = account.PersonContactId
		contact.AccountID = accountID
		return contact, nil
	}

	contactRequest := salesforce.ContactRequest{
		FirstName:   contact.FirstName,
		LastName:    contact.LastName,
		MobilePhone: contact.MobilePhone,
		Email:       contact.Email,
	}
	contactID, err := s.SfcClient.CreateContact(contactRequest)
	if err != nil {
		if err.StatusCode == http.StatusUnauthorized {
			s.RefreshToken()
		}
		return nil, errors.New(helpers.ErrorMessage("not found or create contact", err.Error))
	}
	contact.ID = contactID
	return contact, nil
}

func (s *SalesforceService) SendMessage(affinityToken, sessionKey string, message chat.MessagePayload) (bool, error) {
	return s.SfcChatClient.SendMessage(affinityToken, sessionKey, message)
}

func (s *SalesforceService) GetMessages(affinityToken, sessionKey string) (*chat.MessagesResponse, *helpers.ErrorResponse) {
	return s.SfcChatClient.GetMessages(affinityToken, sessionKey)
}

func (s *SalesforceService) CreatCase(recordType, contactID, description, subject, origin, ownerID string, extraData map[string]interface{}) (string, error) {

	payload := make(map[string]interface{})
	for key, value := range extraData {
		field, ok := s.CustomFields[key]
		if ok {
			payload[field] = value
		}
	}

	description = description + subject
	if value, ok := extraData["description"]; ok {
		description = value.(string)
	}

	caseRequest := NewCaseRequest(recordType, contactID, subject, description, origin, ownerID)
	//validating CaseRequest Payload struct
	if err := helpers.Govalidator().Struct(caseRequest); err != nil {
		return "", errors.New(helpers.ErrorMessage(helpers.InvalidPayload, err))
	}

	payload["RecordTypeId"] = caseRequest.RecordTypeID
	payload["ContactId"] = caseRequest.ContactID
	payload["Description"] = caseRequest.Description
	payload["Origin"] = caseRequest.Origin
	payload["Priority"] = caseRequest.Priority
	payload["Status"] = caseRequest.Status
	payload["Subject"] = caseRequest.Subject
	payload["OwnerId"] = caseRequest.OwnerID
	caseID, errorResponse := s.SfcClient.CreateCase(payload)

	if errorResponse != nil {
		if errorResponse.StatusCode == http.StatusUnauthorized {
			s.RefreshToken()
		}
		return "", errorResponse.Error
	}
	return caseID, nil
}

func (s *SalesforceService) RefreshToken() {
	token, err := s.SfcLoginClient.GetToken(s.TokenPayload)
	if err != nil {
		logrus.Errorf("Could not get access token from salesforce Server : %s", err.Error())
		return
	}
	s.SfcChatClient.UpdateToken(token)
	s.SfcClient.UpdateToken(token)
	logrus.Info("Refresh token successful")
}

func (s *SalesforceService) InsertImageInCase(uri, title, mimeType, caseID string) error {
	resp, err := http.Get(uri)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("image not found")
	}

	var body []byte
	if mimeType == "" {
		contentType, content, err := helpers.GetContentAndTypeByReader(resp.Body)
		if err != nil {
			return err
		}

		mimeType = contentType
		body = helpers.StreamToByte(content)
	} else {
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
	}

	request := salesforce.CompositeRequest{
		AllOrNone:          true,
		CollateSubrequests: false,
		CompositeRequest: []salesforce.Composite{
			{
				Method: http.MethodPost,
				URL:    s.SfcClient.GetContentVersionURL(),
				Body: salesforce.ContentVersionPayload{
					Title:           title,
					ContentLocation: contentLocation,
					PathOnClient:    helpers.GetExportFilename(title, mimeType),
					VersionData:     string(helpers.Encode(body)),
				},
				ReferenceId: "newContentVersion",
			},
			{
				Method:      http.MethodGet,
				URL:         s.SfcClient.GetSearchURL(queryContentDocumentIDByID),
				ReferenceId: "newQuery",
			},
			{
				Method: http.MethodPost,
				URL:    s.SfcClient.GetDocumentLinkURL(),
				Body: salesforce.LinkDocumentPayload{
					ContentDocumentID: linkReferenceID,
					LinkedEntityID:    caseID,
					ShareType:         shareType,
					Visibility:        visibility,
				},
				ReferenceId: "newContentDocumentLink",
			},
		},
	}

	_, err = s.SfcClient.Composite(request)
	if err != nil {
		return err
	}

	return nil
}

func (s *SalesforceService) EndChat(affinityToken, sessionKey string) error {
	return s.SfcChatClient.ChatEnd(affinityToken, sessionKey)
}
