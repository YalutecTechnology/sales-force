package services

import (
	"errors"
	"fmt"
	"strings"

	"github.com/sirupsen/logrus"
	"yalochat.com/salesforce-integration/base/clients/chat"
	"yalochat.com/salesforce-integration/base/clients/login"
	"yalochat.com/salesforce-integration/base/clients/salesforce"
	"yalochat.com/salesforce-integration/base/helpers"
	"yalochat.com/salesforce-integration/base/models"
)

const (
	queryForContactByField = `SELECT+id+,+firstName+,+lastName+,+mobilePhone+,+email+,+CP_BlockedChatYalo__c+FROM+Contact+WHERE+%s+=+` + "%s"
	SFBFieldCustom         = "source_flow_bot"
	SFB1                   = "SFB001"
	SFB2                   = "SFB002"
	SFB3                   = "SFB003"
	SFB4                   = "SFB004"
	SFB5                   = "SFB005"
	SFB6                   = "SFB006"
	SFB1Subject            = "Estado de pedido"
	SFB2Subject            = "Devoluciones y cancelaciones"
	SFB3Subject            = "Estatus de solicitud de crédito"
	SFB4Subject            = "Consulta de estado de cuenta y abonos"
	SFB5Subject            = "Solicitud de préstamo"
	SFB6Subject            = "Otro"
	firstnameDefualt       = "Contacto Bot - "
)

type SalesforceService struct {
	SfcLoginClient login.SfcLoginInterface
	SfcChatClient  chat.SfcChatInterface
	SfcClient      salesforce.SaleforceInterface
}

type SalesforceServiceInterface interface {
	CreatChat(contactName, organizationID, deploymentID, buttonID, caseID, contactID string) (*chat.SessionResponse, error)
	GetOrCreateContact(name, email, phoneNumber string) (*models.SfcContact, error)
	SendMessage(string, string, chat.MessagePayload) (bool, error)
	GetMessages(affinityToken, sessionKey string) (*chat.MessagesResponse, *helpers.ErrorResponse)
	CreatCase(recordType, contactID, description, origin string, extraData map[string]interface{}, customFields []string) (string, error)
}

func NewSalesforceService(loginClient login.SfcLoginClient, chatClient chat.SfcChatClient, salesforceClient salesforce.SalesforceClient) *SalesforceService {
	return &SalesforceService{
		SfcLoginClient: &loginClient,
		SfcChatClient:  &chatClient,
		SfcClient:      &salesforceClient,
	}
}

func NewCaseRequest(recordTypeID, contactID, subject, description, origin string) *salesforce.CaseRequest {
	return &salesforce.CaseRequest{
		RecordTypeID: recordTypeID,
		ContactID:    contactID,
		Subject:      subject,
		Description:  description,
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

func (s *SalesforceService) GetOrCreateContact(name, email, phoneNumber string) (*models.SfcContact, error) {
	// Search contact by email
	contact, err := s.SfcClient.SearchContact(fmt.Sprintf(queryForContactByField, "email", "%27"+email+"%27"))

	if err != nil {
		logrus.Infof("Not found contact by email : [%s]-[%s]", email, err.Error())
	} else {
		return contact, nil
	}
	// Search contact by phone
	contact, err = s.SfcClient.SearchContact(fmt.Sprintf(queryForContactByField, "mobilePhone", "%27"+phoneNumber+"%27"))

	if err != nil {
		logrus.Infof("Not found contact by mobile phone : [%s]-[%s]", phoneNumber, err.Error())
	} else {
		return contact, nil
	}

	contactRequest := salesforce.ContactRequest{
		FirstName:   firstnameDefualt,
		LastName:    name,
		MobilePhone: phoneNumber,
		Email:       email,
	}
	contactID, err := s.SfcClient.CreateContact(contactRequest)
	if err != nil {
		return nil, errors.New(helpers.ErrorMessage("not found or create contact", err))
	}
	contact = &models.SfcContact{
		Id:          contactID,
		FirstName:   contactRequest.FirstName,
		LastName:    contactRequest.LastName,
		Email:       contactRequest.Email,
		MobilePhone: contactRequest.MobilePhone,
	}
	return contact, nil
}

func (s *SalesforceService) SendMessage(affinityToken, sessionKey string, message chat.MessagePayload) (bool, error) {
	return s.SfcChatClient.SendMessage(affinityToken, sessionKey, message)
}

func (s *SalesforceService) GetMessages(affinityToken, sessionKey string) (*chat.MessagesResponse, *helpers.ErrorResponse) {
	return s.SfcChatClient.GetMessages(affinityToken, sessionKey)
}

func (s *SalesforceService) CreatCase(recordType, contactID, description, origin string, extraData map[string]interface{}, customFields []string) (string, error) {
	payload := make(map[string]interface{})
	for _, field := range customFields {
		fields := strings.Split(field, ":")
		yaloField := fields[0]
		sfField := fields[1]
		value, ok := extraData[yaloField]
		if ok {
			payload[sfField] = value
		}
	}

	subject := SFB6Subject
	if value, ok := extraData[SFBFieldCustom]; ok {
		switch value {
		case SFB1:
			subject = SFB1Subject
		case SFB2:
			subject = SFB2Subject
		case SFB3:
			subject = SFB3Subject
		case SFB4:
			subject = SFB4Subject
		case SFB5:
			subject = SFB5Subject
		}
	}

	description = description + subject
	if value, ok := extraData["description"]; ok {
		description = value.(string)
	}

	caseRequest := NewCaseRequest(recordType, contactID, subject, description, origin)
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
	return s.SfcClient.CreateCase(payload)
}
