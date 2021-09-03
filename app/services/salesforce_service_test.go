package services

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"yalochat.com/salesforce-integration/base/clients/chat"
	"yalochat.com/salesforce-integration/base/clients/login"
	"yalochat.com/salesforce-integration/base/clients/proxy"
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
		mock := &proxy.Mock{}
		salesforceService := NewSalesforceService(login.SfcLoginClient{}, chat.SfcChatClient{}, salesforce.SalesforceClient{})
		salesforceService.SfcChatClient.Proxy = mock
		sessionResponse := `{"key":"ec550263-354e-477c-b773-7747ebce3f5e!1629334776994!TrfoJ67wmtlYiENsWdaUBu0xZ7M=","id":"ec550263-354e-477c-b773-7747ebce3f5e","clientPollTimeout":40,"affinityToken":"878a1fa0"}`
		sessionExpected := &chat.SessionResponse{
			Key:               "ec550263-354e-477c-b773-7747ebce3f5e!1629334776994!TrfoJ67wmtlYiENsWdaUBu0xZ7M=",
			Id:                "ec550263-354e-477c-b773-7747ebce3f5e",
			ClientPollTimeout: 40,
			AffinityToken:     "878a1fa0",
		}
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(sessionResponse))),
		}, nil)

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
		mock := &proxy.Mock{}
		salesforceService := NewSalesforceService(login.SfcLoginClient{}, chat.SfcChatClient{}, salesforce.SalesforceClient{})
		salesforceService.SfcClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"totalSize":1,"done":true,"records":[{"attributes":{"type":"Contact","url":"/services/data/v52.0/sobjects/Contact/0032300000Qzu1iAAB"},"Id":"dasfasfasd","FirstName":"contactName","LastName":"contactName","MobilePhone":"5512345678","Email":"user@example.com"}]}`))),
		}, nil)

		contact, err := salesforceService.GetOrCreateContact(contactName, email, phoneNumber)

		assert.NoError(t, err)
		assert.Equal(t, contactExpected, contact)
	})

	t.Run("Get Contact by phone Succesfull", func(t *testing.T) {
		mock := &proxy.Mock{}
		salesforceService := NewSalesforceService(login.SfcLoginClient{}, chat.SfcChatClient{}, salesforce.SalesforceClient{})
		salesforceService.SfcClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewBuffer([]byte(`{"totalSize":0,"done":true,"records":[]}`))),
		}, nil).Once().On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewBuffer([]byte(`{"totalSize":1,"done":true,"records":[{"attributes":{"type":"Contact","url":"/services/data/v52.0/sobjects/Contact/0032300000Qzu1iAAB"},"Id":"dasfasfasd","FirstName":"contactName","LastName":"contactName","MobilePhone":"5512345678","Email":"user@example.com"}]}`))),
		}, nil).Once()

		contact, err := salesforceService.GetOrCreateContact(contactName, email, phoneNumber)

		assert.NoError(t, err)
		assert.Equal(t, contactExpected, contact)
	})

	t.Run("Create Contact Succesfull", func(t *testing.T) {
		mock := &proxy.Mock{}
		salesforceService := NewSalesforceService(login.SfcLoginClient{}, chat.SfcChatClient{}, salesforce.SalesforceClient{})
		salesforceService.SfcClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"totalSize":0,"done":true,"records":[]}`))),
		}, nil).Once().On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"totalSize":0,"done":true,"records":[]}`))),
		}, nil).Once().On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusCreated,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd","success":true,"errors":[]}`))),
		}, nil).Once()

		contact, err := salesforceService.GetOrCreateContact(contactName, email, phoneNumber)

		assert.NoError(t, err)
		assert.Equal(t, contactExpected, contact)
	})

}
