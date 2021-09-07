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
	organizationID = "organizationID"
	deploymentID   = "deploymentID"
	buttonID       = "buttonID"
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

		request := chat.ChatRequest{OrganizationId: "organizationID", DeploymentId: "deploymentID", ButtonId: "buttonID", SessionId: "ec550263-354e-477c-b773-7747ebce3f5e", UserAgent: "Yalo Bot", Language: "es-MX", ScreenResolution: "1900x1080", VisitorName: "contactName", PrechatDetails: []chat.PreChatDetailsObject{chat.PreChatDetailsObject{Label: "CaseId", Value: "caseId", DisplayToAgent: true, TranscriptFields: []string{"CaseId"}}, chat.PreChatDetailsObject{Label: "ContactId", Value: "contactId", DisplayToAgent: true, TranscriptFields: []string{"ContactId"}}}, PrechatEntities: []chat.PrechatEntitiesObject{chat.PrechatEntitiesObject{EntityName: "Case", LinkToEntityName: "Case", LinkToEntityField: "Id", SaveToTranscript: "Case", ShowOnCreate: true, EntityFieldsMaps: []chat.EntityField{chat.EntityField{FieldName: "Id", Label: "CaseId", DoFind: true, IsExactMatch: true, DoCreate: false}}}, chat.PrechatEntitiesObject{EntityName: "Contact", LinkToEntityName: "Contact", LinkToEntityField: "Id", SaveToTranscript: "Contact", ShowOnCreate: true, EntityFieldsMaps: []chat.EntityField{chat.EntityField{FieldName: "Id", Label: "ContactId", DoFind: true, IsExactMatch: true, DoCreate: false}}}}, ReceiveQueueUpdates: true, IsPost: true}
		mock.On("CreateChat", sessionExpected.AffinityToken, sessionExpected.Key, request).Return(false, nil).Once()

		salesforceService := NewSalesforceService(login.SfcLoginClient{}, chat.SfcChatClient{}, salesforce.SalesforceClient{})
		salesforceService.SfcChatClient = mock

		session, err := salesforceService.CreatChat(contactName, organizationID, deploymentID, buttonID, "caseId", "contactId")

		assert.NoError(t, err)
		assert.Equal(t, sessionExpected, session)
	})

}

func TestSalesforceService_GetOrCreateContact(t *testing.T) {
	contactExpected := &models.SfcContact{
		Id:          "dasfasfasd",
		FirstName:   firstnameDefualt,
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

func TestSalesforceService_CreatCase(t *testing.T) {
	t.Run("Create case Succesfull", func(t *testing.T) {
		caseIDExpected := "14224111"
		mock := new(SaleforceInterface)
		salesforceService := NewSalesforceService(login.SfcLoginClient{}, chat.SfcChatClient{}, salesforce.SalesforceClient{})
		salesforceService.SfcClient = mock

		payload := map[string]interface{}{"CP__source_flow_bot__c": "SFB001", "ContactId": "contactId", "Description": "Caso creado por yalo : Estado de pedido", "Origin": "whatsapp", "Priority": "Medium", "RecordTypeId": "recordTypeId", "Status": "Nuevo", "Subject": "Estado de pedido"}
		mock.On("CreateCase", payload).Return(caseIDExpected, nil).Once()

		customFields := []string{"source_flow_bot:CP__source_flow_bot__c"}
		caseId, err := salesforceService.CreatCase("recordTypeId", "contactId", "Caso creado por yalo : ", "whatsapp", map[string]interface{}{"source_flow_bot": "SFB001"}, customFields)

		assert.NoError(t, err)
		assert.Equal(t, caseIDExpected, caseId)
	})

}
