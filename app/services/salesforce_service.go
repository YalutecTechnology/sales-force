package services

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"yalochat.com/salesforce-integration/base/clients/chat"
	"yalochat.com/salesforce-integration/base/clients/login"
	"yalochat.com/salesforce-integration/base/clients/salesforce"
	"yalochat.com/salesforce-integration/base/helpers"
	"yalochat.com/salesforce-integration/base/models"
)

const queryForContactByField = `SELECT+id+,+firstName+,+lastName+,+mobilePhone+,+email+FROM+Contact+WHERE+%s+=+` + "%s"

type SalesforceService struct {
	SfcLoginClient *login.SfcLoginClient
	SfcChatClient  *chat.SfcChatClient
	SfcClient      *salesforce.SalesforceClient
}

type SalesforceServiceInterface interface {
	CreatChat(contactName, organizationId, deploymentId, buttonId string) (*chat.SessionResponse, error)
	GetOrCreateContact(name, email, phoneNumber string) *models.SfcContact
}

func NewSalesforceService(loginClient login.SfcLoginClient, chatClient chat.SfcChatClient, salesforceClient salesforce.SalesforceClient) *SalesforceService {
	return &SalesforceService{
		SfcLoginClient: &loginClient,
		SfcChatClient:  &chatClient,
		SfcClient:      &salesforceClient,
	}
}

func (s *SalesforceService) CreatChat(contactName, organizationId, deploymentId, buttonId string) (*chat.SessionResponse, error) {
	session, err := s.SfcChatClient.CreateSession()
	if err != nil {
		return nil, err
	}

	chatRequest := chat.NewChatRequest(organizationId, deploymentId, session.Id, buttonId, contactName)
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
		FirstName:   name,
		LastName:    name,
		MobilePhone: phoneNumber,
		Email:       email,
	}
	contactId, err := s.SfcClient.CreateContact(contactRequest)
	if err != nil {
		return nil, errors.New(helpers.ErrorMessage("not found or create contact", err))
	}
	contact = &models.SfcContact{
		Id:          contactId,
		FirstName:   contactRequest.FirstName,
		LastName:    contactRequest.LastName,
		Email:       contactRequest.Email,
		MobilePhone: contactRequest.MobilePhone,
	}
	return contact, nil
}
