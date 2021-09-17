package services

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"yalochat.com/salesforce-integration/base/clients/chat"
	"yalochat.com/salesforce-integration/base/clients/login"
	"yalochat.com/salesforce-integration/base/clients/salesforce"
	"yalochat.com/salesforce-integration/base/models"
)

const (
	contactName       = "contactName"
	organizationID    = "organizationId"
	deploymentID      = "deploymentId"
	buttonID          = "buttonId"
	email             = "user@example.com"
	phoneNumber       = "5512345678"
	uri               = "https://icon-icons.com/downloadimage.php?id=142968&root=2348/PNG/32/&file=size_maximize_icon_142968.png"
	mimeType          = "image/png"
	caseID            = "caseID"
	contentVersionID  = "contentVersionID"
	contentDocumentID = "contentDocumentID"
	title             = "title"
	versionData       = "iVBORw0KGgoAAAANSUhEUgAAACAAAAAgCAMAAABEpIrGAAAABGdBTUEAALGPC/xhBQAAAAFzUkdCAK7OHOkAAAAgY0hSTQAAeiYAAICEAAD6AAAAgOgAAHUwAADqYAAAOpgAABdwnLpRPAAAAHJQTFRFAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA////mlzVjAAAACR0Uk5TACtqaUMF7v7twmjrYAHI72QCbfNvBOo9RPFGBk5MS0jyPmPJVqU8vQAAAAFiS0dEJcMByQ8AAAAJcEhZcwABOvYAATr2ATqxVzoAAAC2SURBVDjLvdPJEoIwDAbg0BUKyI4ouPf9n1Fo0XG0yYnxP/T0zSRpG4ANEjHuw4QMAm5fUQIHejliAiTMcE6VSLMcbZJbvSusKisc2Jo1hJiBgZYQM+DQEcIBSgjlLmgR/T4E5DC4+WdxOJKPMk4nCf/IOJ3pQpdr03afzX/n1lsv1vF/Ut1jL/wFhkTpBQpWAQYFXrAaB04UD40DyLN0+YhhIAXjhiU4EOq9BxSwdIl1FyPYIE+RZBNEN1CCzQAAACV0RVh0ZGF0ZTpjcmVhdGUAMjAyMC0wNS0wN1QwOTowNzo0MCswMTowMJZQazMAAAAldEVYdGRhdGU6bW9kaWZ5ADIwMjAtMDUtMDdUMDk6MDc6NDArMDE6MDDnDdOPAAAARnRFWHRzb2Z0d2FyZQBJbWFnZU1hZ2ljayA2LjcuOC05IDIwMTktMDItMDEgUTE2IGh0dHA6Ly93d3cuaW1hZ2VtYWdpY2sub3JnQXviyAAAABh0RVh0VGh1bWI6OkRvY3VtZW50OjpQYWdlcwAxp/+7LwAAABh0RVh0VGh1bWI6OkltYWdlOjpoZWlnaHQANTEywNBQUQAAABd0RVh0VGh1bWI6OkltYWdlOjpXaWR0aAA1MTIcfAPcAAAAGXRFWHRUaHVtYjo6TWltZXR5cGUAaW1hZ2UvcG5nP7JWTgAAABd0RVh0VGh1bWI6Ok1UaW1lADE1ODg4Mzg4NjDthjAYAAAAEnRFWHRUaHVtYjo6U2l6ZQA1LjNLQkLfeornAAAASXRFWHRUaHVtYjo6VVJJAGZpbGU6Ly8uL3VwbG9hZHMvNTYvNDJhTzRoSC8yMzQ4L3NpemVfbWF4aW1pemVfaWNvbl8xNDI5NjgucG5nCRVuAQAAAABJRU5ErkJggg=="
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

func TestSalesforceService_InsertImageInCase(t *testing.T) {
	t.Run("Insert image in case success", func(t *testing.T) {
		salesforceMock := new(SaleforceInterface)
		salesforceService := NewSalesforceService(login.SfcLoginClient{}, chat.SfcChatClient{}, salesforce.SalesforceClient{})
		salesforceService.SfcClient = salesforceMock

		salesforceMock.On("GetContentVersionURL").Return("contentVersionURL").Once()

		salesforceMock.On("GetSearchURL", queryContentDocumentIDByID).Return("searchURL").Once()

		salesforceMock.On("GetDocumentLinkURL").Return("documentLinkURL").Once()

		request := salesforce.CompositeRequest{
			AllOrNone:          true,
			CollateSubrequests: false,
			CompositeRequest: []salesforce.Composite{
				{
					Method: http.MethodPost,
					URL:    "contentVersionURL",
					Body: salesforce.ContentVersionPayload{
						Title:           title,
						ContentLocation: "S",
						PathOnClient:    title + ".png",
						VersionData:     versionData,
					},
					ReferenceId: "newContentVersion",
				},
				{
					Method:      http.MethodGet,
					URL:         "searchURL",
					ReferenceId: "newQuery",
				},
				{
					Method: http.MethodPost,
					URL:    "documentLinkURL",
					Body: salesforce.LinkDocumentPayload{
						ContentDocumentID: linkReferenceID,
						LinkedEntityID:    caseID,
						ShareType:         shareType,
						Visibility:        visibility,
					},
				},
			},
		}
		salesforceMock.On("Composite", request).Return(salesforce.CompositeResponse{}, nil).Once()

		err := salesforceService.InsertImageInCase(uri, title, mimeType, caseID)

		assert.NoError(t, err)
	})

	t.Run("Insert image in case error get image", func(t *testing.T) {
		salesforceMock := new(SaleforceInterface)
		salesforceService := NewSalesforceService(login.SfcLoginClient{}, chat.SfcChatClient{}, salesforce.SalesforceClient{})
		salesforceService.SfcClient = salesforceMock

		contentVersion := salesforce.ContentVersionPayload{
			Title:           title,
			ContentLocation: "S",
			PathOnClient:    title + ".png",
			VersionData:     versionData,
		}
		salesforceMock.On("CreateContentVersion", contentVersion).Return("", assert.AnError).Once()

		err := salesforceService.InsertImageInCase("error", title, mimeType, caseID)

		assert.Error(t, err)
	})

	t.Run("Insert image in case error composite", func(t *testing.T) {
		salesforceMock := new(SaleforceInterface)
		salesforceService := NewSalesforceService(login.SfcLoginClient{}, chat.SfcChatClient{}, salesforce.SalesforceClient{})
		salesforceService.SfcClient = salesforceMock

		salesforceMock.On("GetContentVersionURL").Return("contentVersionURL").Once()

		salesforceMock.On("GetSearchURL", queryContentDocumentIDByID).Return("searchURL").Once()

		salesforceMock.On("GetDocumentLinkURL").Return("documentLinkURL").Once()

		request := salesforce.CompositeRequest{
			AllOrNone:          true,
			CollateSubrequests: false,
			CompositeRequest: []salesforce.Composite{
				{
					Method: http.MethodPost,
					URL:    "contentVersionURL",
					Body: salesforce.ContentVersionPayload{
						Title:           title,
						ContentLocation: "S",
						PathOnClient:    title + ".png",
						VersionData:     versionData,
					},
					ReferenceId: "newContentVersion",
				},
				{
					Method:      http.MethodGet,
					URL:         "searchURL",
					ReferenceId: "newQuery",
				},
				{
					Method: http.MethodPost,
					URL:    "documentLinkURL",
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
		salesforceMock.On("Composite", request).Return(salesforce.CompositeResponse{}, assert.AnError).Once()

		err := salesforceService.InsertImageInCase(uri, title, mimeType, caseID)

		assert.Error(t, err)
	})

}
