package services

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"yalochat.com/salesforce-integration/app/config/envs"
	"yalochat.com/salesforce-integration/base/clients/chat"
	"yalochat.com/salesforce-integration/base/clients/login"
	"yalochat.com/salesforce-integration/base/clients/salesforce"
	"yalochat.com/salesforce-integration/base/constants"
	"yalochat.com/salesforce-integration/base/helpers"
	"yalochat.com/salesforce-integration/base/models"
)

const (
	contentLocation            = "S"
	shareType                  = "V"
	visibility                 = "allUsers"
	queryContentDocumentIDByID = `SELECT+ContentDocumentID+FROM+ContentVersion+WHERE+id+=+'@{newContentVersion.id}'`
	linkReferenceID            = "@{newQuery.records[0].ContentDocumentId}"
)

type SalesforceService struct {
	TokenPayload            login.TokenPayload
	SfcLoginClient          login.SfcLoginInterface
	SfcChatClient           chat.SfcChatInterface
	SfcClient               salesforce.SaleforceInterface
	SourceFlowBot           envs.SfcSourceFlowBot
	SfcCustomFieldsCase     map[string]string
	SfcCustomFieldsContact  map[string]string
	AccountRecordTypeId     string
	DefaultBirthDateAccount string
	RecordTypeID            string
	FirstNameContact        string
}

type SalesforceServiceInterface interface {
	CreatChat(context context.Context, contactName, organizationID, deploymentID, buttonID, caseID, contactID string) (*chat.SessionResponse, error)
	GetOrCreateContact(context context.Context, name, email, phoneNumber string, extraData map[string]interface{}) (*models.SfcContact, error)
	SendMessage(tracer.Span, string, string, chat.MessagePayload) (bool, error)
	GetMessages(mainSpan tracer.Span, affinityToken, sessionKey string) (*chat.MessagesResponse, *helpers.ErrorResponse)
	CreatCase(context context.Context, contactID, description, subject, origin, ownerID string, extraData map[string]interface{}) (string, error)
	InsertImageInCase(uri, title, mimeType, caseID string) error
	EndChat(affinityToken, sessionKey string) error
	RefreshToken()
	SearchContactComposite(email, phoneNumber string) (*models.SfcContact, *helpers.ErrorResponse)
}

func NewSalesforceService(loginClient login.SfcLoginClient, chatClient chat.SfcChatClient, salesforceClient salesforce.SalesforceClient, tokenPayload login.TokenPayload, customFieldsCase map[string]string, recordTypeID, firsNameContact string, customFieldsContact map[string]string) *SalesforceService {
	salesforceService := &SalesforceService{
		SfcLoginClient:          &loginClient,
		SfcChatClient:           &chatClient,
		SfcClient:               &salesforceClient,
		TokenPayload:            tokenPayload,
		SfcCustomFieldsCase:     customFieldsCase,
		SfcCustomFieldsContact:  customFieldsContact,
		RecordTypeID:            recordTypeID,
		DefaultBirthDateAccount: time.Now().Format(constants.DateFormatDateTime),
		FirstNameContact:        firsNameContact,
	}
	salesforceService.RefreshToken()
	return salesforceService
}

func NewContactRequest(firstName, lastName, mobilePhone, email string) *salesforce.ContactRequest {
	return &salesforce.ContactRequest{
		FirstName:   firstName,
		LastName:    lastName,
		MobilePhone: mobilePhone,
		Email:       email,
	}
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

func (s *SalesforceService) CreatChat(ctx context.Context, contactName, organizationID, deploymentID, buttonID, caseID, contactID string) (*chat.SessionResponse, error) {
	// datadog tracing
	span, _ := tracer.StartSpanFromContext(ctx, "salesforceService.CreatChat")
	span.SetTag(ext.AnalyticsEvent, true)
	span.SetTag("contactName", contactName)
	span.SetTag("organizationID", organizationID)
	span.SetTag("deploymentID", deploymentID)
	span.SetTag("buttonID", buttonID)
	span.SetTag("caseID", caseID)
	span.SetTag("contactId", contactID)
	defer span.Finish()

	session, err := s.SfcChatClient.CreateSession(span)
	if err != nil {
		span.SetTag(ext.Error, err)
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

	_, err = s.SfcChatClient.CreateChat(span, session.AffinityToken, session.Key, chatRequest)
	if err != nil {
		span.SetTag(ext.Error, err)
		return nil, err
	}
	return session, nil
}

func (s *SalesforceService) GetOrCreateContact(ctx context.Context, name, email, phoneNumber string, extraData map[string]interface{}) (*models.SfcContact, error) {
	// datadog tracing
	span, _ := tracer.StartSpanFromContext(ctx, "salesforceService.GetOrCreateContact")
	span.SetTag(ext.AnalyticsEvent, true)
	span.SetTag("email", email)
	span.SetTag("phoneNumber", phoneNumber)
	span.SetTag("name", name)
	defer span.Finish()

	contact, err := s.SfcClient.SearchContactComposite(span, email, phoneNumber)
	if err != nil {
		logrus.Errorf("Not found contact search by email or phoneNumber: [%s]-[%s]-[%s]", email, phoneNumber, err.Error.Error())
		if err.StatusCode == http.StatusUnauthorized {
			s.RefreshToken()
		}
		span.SetTag("contactFound", false)
	} else {
		span.SetTag("contactFound", true)
		return contact, nil
	}

	contact = &models.SfcContact{
		FirstName:   s.FirstNameContact,
		LastName:    name,
		Email:       email,
		MobilePhone: phoneNumber,
	}

	if s.AccountRecordTypeId != "" {
		firstName := s.FirstNameContact
		account, err := s.SfcClient.CreateAccountComposite(span, salesforce.AccountRequest{
			FirstName:         &firstName,
			LastName:          &name,
			PersonEmail:       &email,
			PersonMobilePhone: &phoneNumber,
			PersonBirthDate:   &s.DefaultBirthDateAccount,
			RecordTypeID:      &s.AccountRecordTypeId,
		})

		if err != nil {
			if err.StatusCode == http.StatusUnauthorized {
				s.RefreshToken()
			}
			span.SetTag(ext.Error, err.Error)
			return nil, errors.New(helpers.ErrorMessage("not create account", err.Error))
		}

		contact.ID = account.PersonContactId
		contact.AccountID = account.ID
		span.SetTag("createAccount", true)
		return contact, nil
	}

	payload := make(map[string]interface{})
	for key, value := range extraData {
		field, ok := s.SfcCustomFieldsContact[key]
		if ok {
			payload[field] = value
		}
	}

	contactRequest := NewContactRequest(contact.FirstName, contact.LastName, contact.MobilePhone, contact.Email)
	//validating ContactRequest Payload struct
	if err := helpers.Govalidator().Struct(contactRequest); err != nil {
		return nil, errors.New(helpers.ErrorMessage(helpers.InvalidPayload, err))
	}

	payload["FirstName"] = contactRequest.FirstName
	payload["LastName"] = contactRequest.LastName
	payload["MobilePhone"] = contactRequest.MobilePhone
	payload["Email"] = contactRequest.Email

	span.SetTag("payloadContact", fmt.Sprintf("%#v", payload))

	contactID, err := s.SfcClient.CreateContact(span, payload)
	if err != nil {
		if err.StatusCode == http.StatusUnauthorized {
			s.RefreshToken()
		}
		span.SetTag(ext.Error, err.Error)
		return nil, errors.New(helpers.ErrorMessage("not found or create contact", err.Error))
	}
	contact.ID = contactID
	return contact, nil
}

func (s *SalesforceService) SendMessage(mainSpan tracer.Span, affinityToken, sessionKey string, message chat.MessagePayload) (bool, error) {
	return s.SfcChatClient.SendMessage(mainSpan, affinityToken, sessionKey, message)
}

func (s *SalesforceService) GetMessages(mainSpan tracer.Span, affinityToken, sessionKey string) (*chat.MessagesResponse, *helpers.ErrorResponse) {
	return s.SfcChatClient.GetMessages(mainSpan, affinityToken, sessionKey)
}

func (s *SalesforceService) CreatCase(ctx context.Context, contactID, description, subject, origin, ownerID string, extraData map[string]interface{}) (string, error) {
	// datadog tracing
	span, _ := tracer.StartSpanFromContext(ctx, "salesforceService.CreatCase")
	span.SetTag(ext.AnalyticsEvent, true)
	span.SetTag("contactID", contactID)
	span.SetTag("subject", subject)
	span.SetTag("origin", origin)
	span.SetTag("ownerID", ownerID)
	span.SetTag("extraData", extraData)
	defer span.Finish()

	payload := make(map[string]interface{})
	for key, value := range extraData {
		field, ok := s.SfcCustomFieldsCase[key]
		if ok {
			payload[field] = value
		}
	}

	description = fmt.Sprintf("%s : %s", description, subject)
	if value, ok := extraData["description"]; ok {
		description = value.(string)
	}

	caseRequest := NewCaseRequest(s.RecordTypeID, contactID, subject, description, origin, ownerID)
	//validating CaseRequest Payload struct
	if err := helpers.Govalidator().Struct(caseRequest); err != nil {
		span.SetTag(ext.Error, err)
		return "", errors.New(helpers.ErrorMessage(helpers.InvalidPayload, err))
	}

	if caseRequest.RecordTypeID != "" {
		payload["RecordTypeId"] = caseRequest.RecordTypeID
	}

	if ownerID != "" {
		payload["OwnerId"] = caseRequest.OwnerID
	}

	payload["ContactId"] = caseRequest.ContactID
	payload["Description"] = caseRequest.Description
	payload["Origin"] = caseRequest.Origin
	payload["Subject"] = caseRequest.Subject

	if _, ok := payload["Priority"]; !ok {
		payload["Priority"] = caseRequest.Priority
	}
	if _, ok := payload["Status"]; !ok {
		payload["Status"] = caseRequest.Status
	}
	span.SetTag("payloadCase", fmt.Sprintf("%#v", payload))
	caseID, errorResponse := s.SfcClient.CreateCase(span, payload)

	if errorResponse != nil {
		if errorResponse.StatusCode == http.StatusUnauthorized {
			s.RefreshToken()
		}
		span.SetTag(ext.Error, errorResponse.Error)
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
	span := tracer.StartSpan("InsertImageInCase")
	span.SetTag("caseId", caseID)
	span.SetTag("title", title)
	defer span.Finish()

	resp, err := http.Get(uri)
	if err != nil {
		span.SetTag(ext.Error, err)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		err := fmt.Errorf("image not found")
		span.SetTag(ext.Error, err)
		return err
	}

	var body []byte
	if mimeType == "" {
		contentType, content, err := helpers.GetContentAndTypeByReader(resp.Body)
		if err != nil {
			span.SetTag(ext.Error, err)
			return err
		}

		mimeType = contentType
		body = helpers.StreamToByte(content)
	} else {
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			span.SetTag(ext.Error, err)
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

	_, errResponse := s.SfcClient.Composite(span, request)
	if errResponse != nil {
		span.SetTag(ext.Error, errResponse.Error)
		return errors.New(helpers.ErrorMessage("not insert image", errResponse.Error))
	}

	return nil
}

func (s *SalesforceService) EndChat(affinityToken, sessionKey string) error {
	return s.SfcChatClient.ChatEnd(affinityToken, sessionKey)
}

func (s *SalesforceService) SearchContactComposite(email, phoneNumber string) (*models.SfcContact, *helpers.ErrorResponse) {
	span := tracer.StartSpan("SearchContactComposite")
	defer span.Finish()
	return s.SfcClient.SearchContactComposite(span, email, phoneNumber)
}
