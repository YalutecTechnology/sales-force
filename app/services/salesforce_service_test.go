package services

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"yalochat.com/salesforce-integration/base/clients/chat"
	"yalochat.com/salesforce-integration/base/clients/login"
	"yalochat.com/salesforce-integration/base/clients/salesforce"
	"yalochat.com/salesforce-integration/base/models"
)

const (
	contactName    = "contactName"
	organizationId = "organizationId"
	deploymentId   = "deploymentId"
	buttonId       = "buttonId"
	email          = "user@example.com"
	phoneNumber    = "5512345678"
)

func TestSalesforceService_CreatChat(t *testing.T) {

	t.Run("Create Chat Succesfull", func(t *testing.T) {
		mock := new(SfcChatInterface)

		sessionExpected := &chat.SessionResponse{
			Key:               "ec550263-354e-477c-b773-7747ebce3f5e!1629334776994!TrfoJ67wmtlYiENsWdaUBu0xZ7M=",
			Id:                "ec550263-354e-477c-b773-7747ebce3f5e",
			ClientPollTimeout: 40,
			AffinityToken:     "878a1fa0",
		}
		mock.On("CreateSession").Return(sessionExpected, nil).Once()

		request := chat.NewChatRequest(organizationId, deploymentId, sessionExpected.Id, buttonId, contactName)
		mock.On("CreateChat", sessionExpected.AffinityToken, sessionExpected.Key, request).Return(false, nil).Once()

		salesforceService := NewSalesforceService(login.SfcLoginClient{}, chat.SfcChatClient{}, salesforce.SalesforceClient{})
		salesforceService.SfcChatClient = mock

		session, err := salesforceService.CreatChat(contactName, organizationId, deploymentId, buttonId)

		assert.NoError(t, err)
		assert.Equal(t, sessionExpected, session)
	})

}

func TestSalesforceService_GetOrCreateContact(t *testing.T) {
	contactExpected := &models.SfcContact{
		Id:          "dasfasfasd",
		FirstName:   "contactName",
		LastName:    "contactName",
		Email:       "user@example.com",
		MobilePhone: "5512345678",
		Blocked:     false,
	}

	t.Run("Get Contact by email Succesfull", func(t *testing.T) {
		mock := new(SaleforceInterface)
		salesforceService := NewSalesforceService(login.SfcLoginClient{}, chat.SfcChatClient{}, salesforce.SalesforceClient{})
		salesforceService.SfcClient = mock

		mock.On("SearchContact", fmt.Sprintf(queryForContactByField, "email", "%27"+email+"%27")).Return(contactExpected, nil).Once()

		contact, err := salesforceService.GetOrCreateContact(contactName, email, phoneNumber)

		assert.NoError(t, err)
		assert.Equal(t, contactExpected, contact)
	})

	t.Run("Get Contact by phone Succesfull", func(t *testing.T) {
		mock := new(SaleforceInterface)
		salesforceService := NewSalesforceService(login.SfcLoginClient{}, chat.SfcChatClient{}, salesforce.SalesforceClient{})
		salesforceService.SfcClient = mock

		mock.On("SearchContact", fmt.Sprintf(queryForContactByField, "email", "%27"+email+"%27")).Return(nil, assert.AnError).Once()

		mock.On("SearchContact", fmt.Sprintf(queryForContactByField, "mobilePhone", "%27"+phoneNumber+"%27")).Return(contactExpected, nil).Once()

		contact, err := salesforceService.GetOrCreateContact(contactName, email, phoneNumber)

		assert.NoError(t, err)
		assert.Equal(t, contactExpected, contact)
	})

	t.Run("Create Contact Succesfull", func(t *testing.T) {
		mock := new(SaleforceInterface)
		salesforceService := NewSalesforceService(login.SfcLoginClient{}, chat.SfcChatClient{}, salesforce.SalesforceClient{})
		salesforceService.SfcClient = mock

		mock.On("SearchContact", fmt.Sprintf(queryForContactByField, "email", "%27"+email+"%27")).Return(nil, assert.AnError).Once()

		mock.On("SearchContact", fmt.Sprintf(queryForContactByField, "mobilePhone", "%27"+phoneNumber+"%27")).Return(nil, assert.AnError).Once()

		contactRequest := salesforce.ContactRequest{
			FirstName:   contactExpected.FirstName,
			LastName:    contactExpected.LastName,
			MobilePhone: phoneNumber,
			Email:       email,
		}
		mock.On("CreateContact", contactRequest).Return(contactExpected.Id, nil).Once()

		contact, err := salesforceService.GetOrCreateContact(contactName, email, phoneNumber)

		assert.NoError(t, err)
		assert.Equal(t, contactExpected, contact)
	})

}
