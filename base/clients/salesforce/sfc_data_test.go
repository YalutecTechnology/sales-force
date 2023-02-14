package salesforce

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	"yalochat.com/salesforce-integration/base/clients/proxy"
	"yalochat.com/salesforce-integration/base/clients/salesforce/mocks"
	"yalochat.com/salesforce-integration/base/constants"
	"yalochat.com/salesforce-integration/base/models"
)

const (
	caseURL = "test"
	token   = "token"
)

var (
	name         = "name"
	phoneNumber  = "111111"
	email        = "email@example"
	recordTypeID = "reccordTypeId"
	dateBirth    = "2021-10-05T08:12:00"
)

func TestSfcData_CreateContentVersion(t *testing.T) {

	t.Run("Create ContentVersion Succesfull", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusCreated,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd"}`))),
		}, nil)
		payload := ContentVersionPayload{
			Title:           "test image",
			Description:     "A new image",
			ContentLocation: "S",
			PathOnClient:    "screnshoot.jpg",
			VersionData:     "dfhgadfhadf23rubb23",
		}
		id, err := salesforceClient.CreateContentVersion(payload)
		assert.NoError(t, err)
		assert.Equal(t, "dasfasfasd", id)
	})

	t.Run("Create ContentVersion Error validation payload", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd"}`))),
		}, nil)
		payload := ContentVersionPayload{
			Title:           "test image",
			Description:     "A new image",
			ContentLocation: "S",
			PathOnClient:    "screnshoot.jpg",
		}
		id, err := salesforceClient.CreateContentVersion(payload)

		assert.Error(t, err)
		assert.Empty(t, id)
	})

	t.Run("Create ContentVersion error SendHTTPRequest", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{}, assert.AnError)
		payload := ContentVersionPayload{
			Title:           "test image",
			Description:     "A new image",
			ContentLocation: "S",
			PathOnClient:    "screnshoot.jpg",
			VersionData:     "dfhgadfhadf23rubb23",
		}
		id, err := salesforceClient.CreateContentVersion(payload)

		assert.Error(t, err)
		assert.Empty(t, id)
	})

	t.Run("Create ContentVersion error status and unmarshallError", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester("test", token)
		salesforceClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd"}`))),
		}, nil)
		payload := ContentVersionPayload{
			Title:           "test image",
			Description:     "A new image",
			ContentLocation: "S",
			PathOnClient:    "screnshoot.jpg",
			VersionData:     "dfhgadfhadf23rubb23",
		}
		id, err := salesforceClient.CreateContentVersion(payload)

		assert.Error(t, err)
		assert.Equal(t, "Error unmarshalling the response from salesForce : json: cannot unmarshal object into Go value of type []map[string]interface {}", err.Error())
		assert.Empty(t, id)
	})

	t.Run("Create ContentVersion error status", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester("test", token)
		salesforceClient.Proxy = proxyMock
		expectedError := fmt.Sprintf("%s-[%d] : %s", constants.StatusError, http.StatusInternalServerError, `[]map[string]interface {}{map[string]interface {}{"id":"Error create content version"}}`)
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`[{"id":"Error create content version"}]`))),
		}, nil)
		payload := ContentVersionPayload{
			Title:           "test image",
			Description:     "A new image",
			ContentLocation: "S",
			PathOnClient:    "screnshoot.jpg",
			VersionData:     "dfhgadfhadf23rubb23",
		}
		id, err := salesforceClient.CreateContentVersion(payload)

		assert.Error(t, err)
		assert.Empty(t, id)
		assert.Equal(t, expectedError, err.Error())
	})
}

func TestSfcData_SearchDocumentID(t *testing.T) {

	t.Run("SearchDocumentID Succesfull", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{ "totalSize": 1,"done": true, "records": [{"attributes": {},"ContentDocumentId": "AA0"}]}`))),
		}, nil)
		query := "SELECT+ContentDocumentID+FROM+ContentVersion+WHERE+id+=+'01"
		id, err := salesforceClient.SearchDocumentID(query)

		if err != nil {
			t.Fatalf("Expected nil error, but retrieved this %#v", err)
		}

		assert.Equal(t, "AA0", id)
	})

	t.Run("SearchDocumentID Error validation query", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd"}`))),
		}, nil)
		id, err := salesforceClient.SearchDocumentID("")

		assert.Error(t, err)
		assert.Empty(t, id)
	})

	t.Run("SearchDocumentID error SendHTTPRequest", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{}, assert.AnError)

		id, err := salesforceClient.SearchDocumentID("query")

		assert.Error(t, err)
		assert.Empty(t, id)
	})

	t.Run("SearchDocumentID error status", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester("test", token)
		salesforceClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd"}`))),
		}, nil)
		id, err := salesforceClient.SearchDocumentID("query")
		assert.Error(t, err)
		assert.Empty(t, id)
	})

	t.Run("SearchDocumentID Not found", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{ "totalSize": 1,"done": true, "records": []}`))),
		}, nil)
		query := "SELECT+ContentDocumentID+FROM+ContentVersion+WHERE+id+=+'01"
		id, err := salesforceClient.SearchDocumentID(query)

		assert.Error(t, err)
		assert.Empty(t, id)
	})
}

func TestSfcData_Search(t *testing.T) {

	t.Run("Search Succesfull", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{ "totalSize": 1,"done": true, "records": [{"attributes": {},"ContentDocumentId": "AA0"}]}`))),
		}, nil)
		query := "SELECT+ContentDocumentID+FROM+ContentVersion+WHERE+id+=+'01"
		id, err := salesforceClient.Search(query)

		if err != nil {
			t.Fatalf("Expected nil error, but retrieved this %#v", err)
		}

		expectedResponse := &SearchResponse{
			TotalSize: 1,
			Done:      true,
			Records: []recordResponse{
				{
					Attributes:        map[string]interface{}{},
					ContentDocumentID: "AA0",
					Id:                ``,
				},
			},
		}
		assert.Equal(t, expectedResponse, id)
	})

	t.Run("Search Error validation query", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd"}`))),
		}, nil)
		id, err := salesforceClient.Search("")

		assert.Error(t, err.Error)
		assert.Empty(t, id)
	})

	t.Run("Search error SendHTTPRequest", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{}, assert.AnError)

		id, err := salesforceClient.Search("query")

		assert.Error(t, err.Error)
		assert.Empty(t, id)
	})

	t.Run("SearchDocumentID error status whit unmasrhall error", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester("test", token)
		salesforceClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd"}`))),
		}, nil)
		id, err := salesforceClient.Search("query")
		assert.Error(t, err.Error)
		assert.Empty(t, id)
	})

	t.Run("SearchDocumentID error status", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester("test", token)
		salesforceClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`[{"id":"dasfasfasd"}]`))),
		}, nil)
		id, err := salesforceClient.Search("query")
		assert.Error(t, err.Error)
		assert.Empty(t, id)
	})
}

func TestSfcData_SearchId(t *testing.T) {

	t.Run("SearchId Succesfull", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"totalSize":7,"done":true,"records":[{"attributes":{"type":"Contact","url":"/services/data/v52.0/sobjects/Contact/0032300000Qn8e5AAB"},"Name":"Mauricio Ruiz","LastName":"Ruiz","Id":"0032300000Qn8e5AAB"}]}`))),
		}, nil)
		query := "SELECT+name+,+lastName+,+id+FROM+Contact+WHERE+email+=+%27mauricio.ruiz@intellectsystem.net%27"
		id, err := salesforceClient.SearchID(query)

		if err != nil {
			t.Fatalf("Expected nil error, but retrieved this %#v", err)
		}

		assert.Equal(t, "0032300000Qn8e5AAB", id)
	})

	t.Run("Search Error validation query", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd"}`))),
		}, nil)
		id, err := salesforceClient.SearchID("")

		assert.Error(t, err)
		assert.Empty(t, id)
	})

	t.Run("Search error SendHTTPRequest", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{}, assert.AnError)

		id, err := salesforceClient.SearchID("query")

		assert.Error(t, err)
		assert.Empty(t, id)
	})

	t.Run("SearchDocumentID error status", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester("test", token)
		salesforceClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd"}`))),
		}, nil)
		id, err := salesforceClient.SearchID("query")
		assert.Error(t, err)
		assert.Empty(t, id)
	})

	t.Run("SearchId Not found", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"totalSize":0,"done":true,"records":[]}`))),
		}, nil)
		query := "SELECT+name+,+lastName+,+id+FROM+Contact+WHERE+email+=+%27mauricio.ruiz@intellectsystem.net%27"
		id, err := salesforceClient.SearchID(query)

		assert.Error(t, err)
		assert.Empty(t, id)
	})
}

func TestSfcData_SearchContact(t *testing.T) {

	t.Run("SearchContact Succesfull", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"totalSize":1,"done":true,"records":[{"attributes":{"type":"Contact","url":"/services/data/v52.0/sobjects/Contact/0032300000Qzu1iAAB"},"Id":"0032300000Qzu1iAAB","FirstName":"name","LastName":"lastname","MobilePhone":"55555","Email":"user@example.com"}]}`))),
		}, nil)
		contactExpected := &models.SfcContact{
			ID:          "0032300000Qzu1iAAB",
			FirstName:   "name",
			LastName:    "lastname",
			Email:       "user@example.com",
			MobilePhone: "55555",
		}
		query := "SELECT+id+,+firstName+,+lastName+,+mobilePhone+,+email+FROM+Contact+WHERE+mobilePhone+=+%277331175599%27"
		contact, err := salesforceClient.SearchContact(query)

		if err != nil {
			t.Fatalf("Expected nil error, but retrieved this %#v", err)
		}

		assert.Equal(t, contactExpected, contact)
	})

	t.Run("Search Error validation query", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd"}`))),
		}, nil)
		contact, err := salesforceClient.SearchContact("")

		assert.Error(t, err.Error)
		assert.Empty(t, contact)
	})

	t.Run("Search error SendHTTPRequest", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{}, assert.AnError)

		id, err := salesforceClient.SearchContact("query")

		assert.Error(t, err.Error)
		assert.Empty(t, id)
	})

	t.Run("SearchContact error status", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester("test", token)
		salesforceClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd"}`))),
		}, nil)
		id, err := salesforceClient.SearchContact("query")
		assert.Error(t, err.Error)
		assert.Empty(t, id)
	})

	t.Run("SearchContact not found", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"totalSize":0,"done":true,"records":[]}`))),
		}, nil)

		query := "SELECT+id+,+firstName+,+lastName+,+mobilePhone+,+email+FROM+Contact+WHERE+mobilePhone+=+%277331175599%27"
		contact, err := salesforceClient.SearchContact(query)

		assert.Error(t, err.Error)
		assert.Empty(t, contact)
	})
}

func TestSfcData_SearchAccount(t *testing.T) {

	t.Run("SearchAccount Succesfull", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body: ioutil.NopCloser(bytes.NewReader([]byte(`{
				"totalSize": 1,
				"done": true,
				"records": [
					{
						"attributes": {
							"type": "Account",
							"url": "/services/data/v52.0/sobjects/Account/0017b00000zG5WpAAK"
						},
						"Id": "0017b00000zG5WpAAK",
						"FirstName": "Contacto creado por el Bot-",
						"LastName": "Edauardo",
						"PersonMobilePhone": "5217331175598",
						"PersonEmail": "ochoa@example.com",
						"PersonContactId": "0037b00000rsgDjAAI"
					}
				]
			}`))),
		}, nil)
		accountExpected := &models.SfcAccount{
			ID:                "0017b00000zG5WpAAK",
			FirstName:         "Contacto creado por el Bot-",
			LastName:          "Edauardo",
			PersonEmail:       "ochoa@example.com",
			PersonMobilePhone: "5217331175598",
			PersonContactId:   "0037b00000rsgDjAAI",
		}
		query := "SELECT+id+,+firstName+,+lastName+,+PersonMobilePhone+,+PersonEmail+,+PersonContactId+FROM+Account+WHERE+id+=+'accountID'"
		account, err := salesforceClient.SearchAccount(query)
		assert.Nil(t, err)
		assert.Equal(t, accountExpected, account)
	})

	t.Run("Search Error validation query", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd"}`))),
		}, nil)
		account, err := salesforceClient.SearchAccount("")

		assert.Error(t, err.Error)
		assert.Empty(t, account)
	})

	t.Run("Search error SendHTTPRequest", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{}, assert.AnError)

		account, err := salesforceClient.SearchAccount("query")

		assert.Error(t, err.Error)
		assert.Empty(t, account)
	})

	t.Run("SearchContact error status", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester("test", token)
		salesforceClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd"}`))),
		}, nil)
		account, err := salesforceClient.SearchAccount("query")
		assert.Error(t, err.Error)
		assert.Empty(t, account)
	})

	t.Run("SearchAccount Not Found", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body: ioutil.NopCloser(bytes.NewReader([]byte(`{
				"totalSize": 0,
				"done": true,
				"records": []
			}`))),
		}, nil)

		query := "SELECT+id+,+firstName+,+lastName+,+PersonMobilePhone+,+PersonEmail+,+PersonContactId+FROM+Account+WHERE+id+=+'accountID'"
		account, err := salesforceClient.SearchAccount(query)
		assert.Error(t, err.Error)
		assert.Empty(t, account)
	})
}

func TestSfcData_LinkDocumentToCase(t *testing.T) {

	t.Run("LinkDocumentToCase Succesfull", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusCreated,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd"}`))),
		}, nil)
		payload := LinkDocumentPayload{
			ContentDocumentID: "g00",
			LinkedEntityID:    "AAZ",
			ShareType:         "V",
			Visibility:        "S",
		}
		id, err := salesforceClient.LinkDocumentToCase(payload)

		if err != nil {
			t.Fatalf("Expected nil error, but retrieved this %#v", err)
		}

		assert.Equal(t, "dasfasfasd", id)
	})

	t.Run("Create ContentVersion Error validation payload", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd"}`))),
		}, nil)
		payload := LinkDocumentPayload{
			ContentDocumentID: "g00",
			LinkedEntityID:    "AAZ",
			Visibility:        "S",
		}
		id, err := salesforceClient.LinkDocumentToCase(payload)

		assert.Error(t, err)
		assert.Empty(t, id)
	})

	t.Run("Create ContentVersion error SendHTTPRequest", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{}, assert.AnError)
		payload := LinkDocumentPayload{
			ContentDocumentID: "g00",
			LinkedEntityID:    "AAZ",
			ShareType:         "V",
			Visibility:        "S",
		}
		id, err := salesforceClient.LinkDocumentToCase(payload)

		assert.Error(t, err)
		assert.Empty(t, id)
	})

	t.Run("Create ContentVersion error status", func(t *testing.T) {
		expectedError := fmt.Sprintf("%s-[%d] : %s", constants.StatusError, http.StatusInternalServerError, `[]map[string]interface {}{map[string]interface {}{"id":"LinkDocument error"}}`)
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester("test", token)
		salesforceClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`[{"id":"LinkDocument error"}]`))),
		}, nil)
		payload := LinkDocumentPayload{
			ContentDocumentID: "g00",
			LinkedEntityID:    "AAZ",
			ShareType:         "V",
			Visibility:        "S",
		}
		id, err := salesforceClient.LinkDocumentToCase(payload)

		assert.Error(t, err)
		assert.Empty(t, id)
		assert.Equal(t, expectedError, err.Error())
	})
}

func TestSfcData_CreateCase(t *testing.T) {

	t.Run("Create case Succesfull", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusCreated,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd"}`))),
		}, nil)
		payload := CaseRequest{
			ContactID:   "contact id",
			Status:      "New",
			Origin:      "Web",
			Subject:     "test",
			Priority:    "Medium",
			IsEscalated: false,
			Description: "context",
		}
		span, _ := tracer.SpanFromContext(context.Background())
		id, err := salesforceClient.CreateCase(span, payload)

		if err != nil {
			t.Fatalf("Expected nil error, but retrieved this %#v", err)
		}

		assert.Equal(t, "dasfasfasd", id)
	})

	t.Run("Create case  error SendHTTPRequest", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{}, assert.AnError)
		payload := CaseRequest{
			ContactID:   "contact id",
			Status:      "New",
			Origin:      "Web",
			Subject:     "test",
			Priority:    "Medium",
			IsEscalated: false,
			Description: "context",
		}
		span, _ := tracer.SpanFromContext(context.Background())
		id, err := salesforceClient.CreateCase(span, payload)

		assert.Error(t, err.Error)
		assert.Empty(t, id)
	})

	t.Run("Create case Unmarshall response", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`error`))),
		}, nil)
		payload := CaseRequest{
			ContactID:   "contact id",
			Status:      "New",
			Origin:      "Web",
			Subject:     "test",
			Priority:    "Medium",
			IsEscalated: false,
			Description: "context",
		}
		span, _ := tracer.SpanFromContext(context.Background())
		id, err := salesforceClient.CreateCase(span, payload)

		assert.Error(t, err.Error)
		assert.Empty(t, id)
	})

	t.Run("Create case error status", func(t *testing.T) {
		expectedError := fmt.Sprintf("%s-[%d] : %s", constants.StatusError, http.StatusInternalServerError, "[map[id:Create case error status]]")
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester("test", token)
		salesforceClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`[{"id":"Create case error status"}]`))),
		}, nil)
		payload := CaseRequest{
			ContactID:   "contact id",
			Status:      "New",
			Origin:      "Web",
			Subject:     "test",
			Priority:    "Medium",
			IsEscalated: false,
			Description: "context",
		}
		span, _ := tracer.SpanFromContext(context.Background())
		id, err := salesforceClient.CreateCase(span, payload)

		assert.Error(t, err.Error)
		assert.Empty(t, id)
		assert.Equal(t, expectedError, err.Error.Error())
	})
}

func TestSfcData_CreateContact(t *testing.T) {
	span, _ := tracer.SpanFromContext(context.Background())

	t.Run("Create contact Succesfull", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusCreated,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd"}`))),
		}, nil).Once()
		payload := ContactRequest{
			FirstName:   "firstname",
			LastName:    "lasrname",
			MobilePhone: "11111111111",
			Email:       "test@mail.com",
		}
		id, err := salesforceClient.CreateContact(span, payload)

		if err != nil {
			t.Fatalf("Expected nil error, but retrieved this %#v", err)
		}

		assert.Equal(t, "dasfasfasd", id)
	})

	t.Run("Create contact error SendHTTPRequest", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{}, assert.AnError)
		payload := ContactRequest{
			FirstName:   "firstname",
			LastName:    "lasrname",
			MobilePhone: "11111111111",
			Email:       "test@mail.com",
		}
		id, err := salesforceClient.CreateContact(span, payload)

		assert.Error(t, err.Error)
		assert.Empty(t, id)
	})

	t.Run("Create contact Unmarshal response", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusCreated,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`error`))),
		}, nil)
		payload := ContactRequest{
			FirstName:   "firstname",
			LastName:    "lasrname",
			MobilePhone: "11111111111",
			Email:       "test@mail.com",
		}
		id, err := salesforceClient.CreateContact(span, payload)

		assert.Error(t, err.Error)
		assert.EqualError(t, err.Error, `Error unmarshalling the response from salesForce : invalid character 'e' looking for beginning of value`)
		assert.Empty(t, id)
	})

	t.Run("Create contact error status", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester("test", token)
		salesforceClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd"}`))),
		}, nil)
		payload := ContactRequest{
			FirstName:   "firstname",
			LastName:    "lasrname",
			MobilePhone: "11111111111",
			Email:       "test@mail.com",
		}
		id, err := salesforceClient.CreateContact(span, payload)

		assert.Error(t, err.Error)
		assert.Empty(t, id)
	})
}

func TestSfcData_CreateAccount(t *testing.T) {

	t.Run("Create account Succesfull", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusCreated,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd"}`))),
		}, nil).Once()
		payload := AccountRequest{
			FirstName:         &name,
			LastName:          &name,
			PersonEmail:       &email,
			PersonMobilePhone: &phoneNumber,
			RecordTypeID:      &recordTypeID,
			PersonBirthDate:   &dateBirth,
		}
		id, err := salesforceClient.CreateAccount(payload)
		assert.Nil(t, err)
		assert.Equal(t, "dasfasfasd", id)
	})

	t.Run("Create account  error validation payload", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd"}`))),
		}, nil)
		payload := AccountRequest{
			Name: &name,
		}
		id, err := salesforceClient.CreateAccount(payload)

		assert.Error(t, err.Error)
		assert.Empty(t, id)
	})

	t.Run("Create account  error SendHTTPRequest", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{}, assert.AnError)
		payload := AccountRequest{
			FirstName:       &name,
			PersonEmail:     &name,
			LastName:        &phoneNumber,
			RecordTypeID:    &recordTypeID,
			PersonBirthDate: &dateBirth,
		}
		id, err := salesforceClient.CreateAccount(payload)

		assert.Error(t, err.Error)
		assert.Empty(t, id)
	})

	t.Run("Create account error status", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester("test", token)
		salesforceClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd"}`))),
		}, nil)
		payload := AccountRequest{
			FirstName:       &name,
			PersonEmail:     &name,
			LastName:        &phoneNumber,
			RecordTypeID:    &recordTypeID,
			PersonBirthDate: &dateBirth,
		}
		id, err := salesforceClient.CreateAccount(payload)

		assert.Error(t, err.Error)
		assert.Empty(t, id)
	})

}

func TestSfcData_Composite(t *testing.T) {
	span, _ := tracer.SpanFromContext(context.Background())
	t.Run("Create Composite Succesfull", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = proxyMock

		expected := CompositeResponses{
			CompositeResponse: []CompositeResponse{
				{
					Body: "test",
					HTTPHeaders: HTTPHeaders{
						Location: "location",
					},
					HTTPStatusCode: http.StatusOK,
					ReferenceID:    "referenceID",
				},
			},
		}
		expectedBin, err := json.Marshal(expected)
		assert.NoError(t, err)

		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(expectedBin)),
		}, nil).Once()
		payload := CompositeRequest{
			AllOrNone:          true,
			CollateSubrequests: true,
			CompositeRequest: []Composite{
				{
					Method: "POST",
					URL:    "URL",
					Body: ContentVersionPayload{
						Title:           "title",
						Description:     "description",
						ContentLocation: "location",
						PathOnClient:    "path",
						VersionData:     "version",
					},
				},
			},
		}
		response, errResponse := salesforceClient.Composite(span, payload)

		assert.Nil(t, errResponse)
		assert.Equal(t, expected, response)
	})

	t.Run("Create Composite error validation", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = proxyMock

		payload := CompositeRequest{
			AllOrNone:          true,
			CollateSubrequests: true,
		}
		response, err := salesforceClient.Composite(span, payload)

		assert.Error(t, err.Error)
		assert.Empty(t, response)
	})

	t.Run("Create Composite error status", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = proxyMock

		expected := CompositeResponse{
			Body: "test",
			HTTPHeaders: HTTPHeaders{
				Location: "location",
			},
			HTTPStatusCode: http.StatusOK,
			ReferenceID:    "referenceID",
		}
		expectedBin, err := json.Marshal(expected)
		assert.NoError(t, err)

		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusNotFound,
			Body:       ioutil.NopCloser(bytes.NewReader(expectedBin)),
		}, nil).Once()
		payload := CompositeRequest{
			AllOrNone:          true,
			CollateSubrequests: true,
			CompositeRequest: []Composite{
				{
					Method: "POST",
					URL:    "URL",
					Body: ContentVersionPayload{
						Title:           "title",
						Description:     "description",
						ContentLocation: "location",
						PathOnClient:    "path",
						VersionData:     "version",
					},
				},
			},
		}
		response, errResponse := salesforceClient.Composite(span, payload)

		assert.Error(t, errResponse.Error)
		assert.Empty(t, response)
	})
}

func TestSfcData_UpdateToken(t *testing.T) {

	t.Run("Update token Succesfull", func(t *testing.T) {
		tokenExpected := "14525542211224"
		salesforceClient := NewSalesforceRequester(caseURL, token)

		salesforceClient.UpdateToken(tokenExpected)
		assert.Equal(t, tokenExpected, salesforceClient.AccessToken)
	})
}

func TestSfcData_SearchContactComposite(t *testing.T) {
	span, _ := tracer.SpanFromContext(context.Background())
	t.Run("SearchContactComposite Succesfull", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = proxyMock
		salesforceClient.SfcBlockedChatField = true

		response := CompositeResponses{
			CompositeResponse: []CompositeResponse{
				{
					Body: SearchResponse{
						TotalSize: 0,
						Done:      false,
						Records:   []recordResponse{},
					},
				},
				{
					Body: SearchResponse{
						TotalSize: 1,
						Done:      true,
						Records: []recordResponse{
							{
								Id:              "0032300000Qzu1iAAB",
								FirstName:       "name",
								LastName:        "lastname",
								Email:           "user@example.com",
								MobilePhone:     "55555",
								BlockedChatYalo: true,
							},
						},
					},
				},
			},
		}

		responseBin, err := json.Marshal(response)
		assert.NoError(t, err)

		stringPayload := `{"allOrNone":false,"collateSubrequests":false,"compositeRequest":[{"method":"GET","url":"/services/data/v.0/query/?q=SELECT+id+,+firstName+,+lastName+,+mobilePhone+,+email+,+CP_BlockedChatYalo__c+FROM+Contact+WHERE+email+=+'email@example'","body":null,"referenceId":"newQueryEmail"},{"method":"GET","url":"/services/data/v.0/query/?q=SELECT+id+,+firstName+,+lastName+,+mobilePhone+,+email+,+CP_BlockedChatYalo__c+FROM+Contact+WHERE+mobilePhone+=+'111111'","body":null,"referenceId":"newQueryPhone"}]}`
		header := make(map[string]string)
		header["Content-Type"] = "application/json"
		header["Authorization"] = fmt.Sprintf("Bearer %s", token)

		expectedPayload := &proxy.Request{
			Body:      []byte(stringPayload),
			Method:    http.MethodPost,
			URI:       "/services/data/v.0/composite",
			HeaderMap: header,
		}

		proxyMock.On("SendHTTPRequest", mock.Anything, expectedPayload).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(responseBin)),
		}, nil)

		contactExpected := &models.SfcContact{
			ID:          "0032300000Qzu1iAAB",
			FirstName:   "name",
			LastName:    "lastname",
			Email:       "user@example.com",
			MobilePhone: "55555",
			Blocked:     true,
		}

		contact, errResponse := salesforceClient.SearchContactComposite(span, email, phoneNumber, nil, nil)

		assert.Nil(t, errResponse)
		assert.Equal(t, contactExpected, contact)
	})

	t.Run("SearchContactComposite with custom fields", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = proxyMock
		salesforceClient.SfcBlockedChatField = true

		response := CompositeResponses{
			CompositeResponse: []CompositeResponse{
				{
					Body: SearchResponse{
						TotalSize: 0,
						Done:      true,
						Records:   []recordResponse{},
					},
				},
				{
					Body: SearchResponse{
						TotalSize: 0,
						Done:      true,
						Records:   []recordResponse{},
					},
				},
				{
					Body: SearchResponse{
						TotalSize: 1,
						Done:      true,
						Records: []recordResponse{
							{
								Id:              "0032300000Qzu1iAAB",
								FirstName:       "name",
								LastName:        "lastname",
								Email:           "another@email.com",
								MobilePhone:     "12345",
								BlockedChatYalo: true,
							},
						},
					},
				},
			},
		}

		responseBin, err := json.Marshal(response)
		assert.NoError(t, err)

		stringPayload := `{"allOrNone":false,"collateSubrequests":false,"compositeRequest":[{"method":"GET","url":"/services/data/v.0/query/?q=SELECT+id+,+firstName+,+lastName+,+mobilePhone+,+email+,+CP_BlockedChatYalo__c+FROM+Contact+WHERE+email+=+'email@example'","body":null,"referenceId":"newQueryEmail"},{"method":"GET","url":"/services/data/v.0/query/?q=SELECT+id+,+firstName+,+lastName+,+mobilePhone+,+email+,+CP_BlockedChatYalo__c+FROM+Contact+WHERE+mobilePhone+=+'111111'","body":null,"referenceId":"newQueryPhone"},{"method":"GET","url":"/services/data/v.0/query/?q=SELECT+id+,+firstName+,+lastName+,+mobilePhone+,+email+,+CP_BlockedChatYalo__c+FROM+Contact+WHERE+Account.Customer__ID__SF+=+'333333'","body":null,"referenceId":"newQueryCustomField_Account_Customer_ID_SF"}]}`
		header := make(map[string]string)
		header["Content-Type"] = "application/json"
		header["Authorization"] = fmt.Sprintf("Bearer %s", token)

		expectedPayload := &proxy.Request{
			Body:      []byte(stringPayload),
			Method:    http.MethodPost,
			URI:       "/services/data/v.0/composite",
			HeaderMap: header,
		}

		proxyMock.On("SendHTTPRequest", mock.Anything, expectedPayload).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(responseBin)),
		}, nil)
		contactExpected := &models.SfcContact{
			ID:          "0032300000Qzu1iAAB",
			FirstName:   "name",
			LastName:    "lastname",
			Email:       "another@email.com",
			MobilePhone: "12345",
			Blocked:     true,
		}

		contact, errResponse := salesforceClient.SearchContactComposite(span, email, phoneNumber, map[string]string{"customerId": "Account.Customer__ID__SF", "customerAnotherId": "Customer__Another__ID__SF"}, map[string]interface{}{"customerId": "333333", "name": "my name", "address": "my address"})

		assert.Nil(t, errResponse)
		assert.Equal(t, contactExpected, contact)
	})

	t.Run("SearchContactComposite could not find custom field data on extraData", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = proxyMock
		salesforceClient.SfcBlockedChatField = true

		response := CompositeResponses{
			CompositeResponse: []CompositeResponse{
				{
					Body: SearchResponse{
						TotalSize: 0,
						Done:      true,
						Records:   []recordResponse{},
					},
				},
				{
					Body: SearchResponse{
						TotalSize: 1,
						Done:      true,
						Records: []recordResponse{
							{
								Id:              "0032300000Qzu1iAAB",
								FirstName:       "name",
								LastName:        "lastname",
								Email:           "another@email.com",
								MobilePhone:     "12345",
								BlockedChatYalo: true,
							},
						},
					},
				},
			},
		}

		responseBin, err := json.Marshal(response)
		assert.NoError(t, err)

		stringPayload := `{"allOrNone":false,"collateSubrequests":false,"compositeRequest":[{"method":"GET","url":"/services/data/v.0/query/?q=SELECT+id+,+firstName+,+lastName+,+mobilePhone+,+email+,+CP_BlockedChatYalo__c+FROM+Contact+WHERE+email+=+'email@example'","body":null,"referenceId":"newQueryEmail"},{"method":"GET","url":"/services/data/v.0/query/?q=SELECT+id+,+firstName+,+lastName+,+mobilePhone+,+email+,+CP_BlockedChatYalo__c+FROM+Contact+WHERE+mobilePhone+=+'111111'","body":null,"referenceId":"newQueryPhone"}]}`
		header := make(map[string]string)
		header["Content-Type"] = "application/json"
		header["Authorization"] = fmt.Sprintf("Bearer %s", token)

		expectedPayload := &proxy.Request{
			Body:      []byte(stringPayload),
			Method:    http.MethodPost,
			URI:       "/services/data/v.0/composite",
			HeaderMap: header,
		}

		proxyMock.On("SendHTTPRequest", mock.Anything, expectedPayload).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(responseBin)),
		}, nil)
		contactExpected := &models.SfcContact{
			ID:          "0032300000Qzu1iAAB",
			FirstName:   "name",
			LastName:    "lastname",
			Email:       "another@email.com",
			MobilePhone: "12345",
			Blocked:     true,
		}

		contact, errResponse := salesforceClient.SearchContactComposite(span, email, phoneNumber, map[string]string{"customerId": "Customer__ID__SF"}, map[string]interface{}{"email": "my@mail.com", "name": "my name", "address": "my address"})

		assert.Nil(t, errResponse)
		assert.Equal(t, contactExpected, contact)
	})

	t.Run("SearchContactComposite notFount", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = proxyMock

		response := CompositeResponses{
			CompositeResponse: []CompositeResponse{},
		}

		responseBin, err := json.Marshal(response)
		assert.NoError(t, err)
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(responseBin)),
		}, nil)

		contact, errResponse := salesforceClient.SearchContactComposite(span, email, phoneNumber, nil, nil)

		assert.NotNil(t, errResponse)
		assert.Empty(t, contact)
	})

	t.Run("SearchContactComposite request error", func(t *testing.T) {
		expectedError := fmt.Sprintf("%s-[%d] : %s", constants.StatusError, http.StatusNotFound, "[map[compositeResponse:[]]]")
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = proxyMock

		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusNotFound,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`[{"compositeResponse":[]}]`))),
		}, nil)

		contact, errResponse := salesforceClient.SearchContactComposite(span, email, phoneNumber, nil, nil)

		assert.NotNil(t, errResponse)
		assert.Empty(t, contact)
		assert.Equal(t, expectedError, errResponse.Error.Error())
	})
}

func TestSfcData_CreateAccountComposite(t *testing.T) {
	span, _ := tracer.SpanFromContext(context.Background())
	t.Run("Create account Succesfull", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = proxyMock

		response := CompositeResponses{
			CompositeResponse: []CompositeResponse{
				{
					Body: SearchResponse{
						TotalSize: 0,
						Done:      false,
						Records:   []recordResponse{},
					},
					HTTPStatusCode: 201,
				},
				{
					Body: SearchResponse{
						TotalSize: 1,
						Done:      true,
						Records: []recordResponse{
							{
								Id:                "0032300000Qzu1iAAB",
								FirstName:         name,
								LastName:          name,
								PersonEmail:       email,
								PersonMobilePhone: phoneNumber,
								PersonContactID:   "contactID",
							},
						},
					},
					HTTPStatusCode: 200,
				},
			},
		}

		responseBin, err := json.Marshal(response)
		assert.NoError(t, err)

		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(responseBin)),
		}, nil).Once()
		payload := AccountRequest{
			FirstName:         &name,
			LastName:          &name,
			PersonEmail:       &email,
			PersonMobilePhone: &phoneNumber,
			RecordTypeID:      &recordTypeID,
			PersonBirthDate:   &dateBirth,
		}

		contactExpected := &models.SfcAccount{
			ID:                "0032300000Qzu1iAAB",
			FirstName:         name,
			LastName:          name,
			PersonEmail:       email,
			PersonMobilePhone: phoneNumber,
			PersonContactId:   "contactID",
		}
		contact, errResponse := salesforceClient.CreateAccountComposite(span, payload)
		assert.Nil(t, errResponse)
		assert.Equal(t, contactExpected, contact)
	})

	t.Run("Could not create the account", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = proxyMock

		response := CompositeResponses{
			CompositeResponse: []CompositeResponse{
				{
					Body: SearchResponse{
						TotalSize: 0,
						Done:      false,
						Records:   []recordResponse{},
					},
					HTTPStatusCode: 400,
				},
				{
					Body: SearchResponse{
						TotalSize: 1,
						Done:      true,
						Records:   []recordResponse{},
					},
					HTTPStatusCode: 400,
				},
			},
		}

		responseBin, err := json.Marshal(response)
		assert.NoError(t, err)

		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(responseBin)),
		}, nil).Once()
		payload := AccountRequest{
			FirstName:         &name,
			LastName:          &name,
			PersonEmail:       &email,
			PersonMobilePhone: &phoneNumber,
			RecordTypeID:      &recordTypeID,
			PersonBirthDate:   &dateBirth,
		}

		contact, errResponse := salesforceClient.CreateAccountComposite(span, payload)
		assert.Error(t, errResponse.Error)
		assert.Empty(t, contact)
	})

	t.Run("Create account error account not found", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = proxyMock

		response := CompositeResponses{
			CompositeResponse: []CompositeResponse{
				{
					Body: SearchResponse{
						TotalSize: 0,
						Done:      false,
						Records:   []recordResponse{},
					},
					HTTPStatusCode: 201,
				},
				{
					Body: SearchResponse{
						TotalSize: 0,
						Done:      true,
						Records:   []recordResponse{},
					},
					HTTPStatusCode: 200,
				},
			},
		}

		responseBin, err := json.Marshal(response)
		assert.NoError(t, err)

		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(responseBin)),
		}, nil).Once()
		payload := AccountRequest{
			FirstName:         &name,
			LastName:          &name,
			PersonEmail:       &email,
			PersonMobilePhone: &phoneNumber,
			RecordTypeID:      &recordTypeID,
			PersonBirthDate:   &dateBirth,
		}

		contact, errResponse := salesforceClient.CreateAccountComposite(span, payload)
		assert.Error(t, errResponse.Error)
		assert.Empty(t, contact)
	})

	t.Run("Create account error account not found ", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = proxyMock

		response := CompositeResponses{
			CompositeResponse: []CompositeResponse{
				{
					Body: SearchResponse{
						TotalSize: 0,
						Done:      false,
						Records:   []recordResponse{},
					},
					HTTPStatusCode: 201,
				},
				{
					Body: SearchResponse{
						TotalSize: 0,
						Done:      true,
						Records:   []recordResponse{},
					},
					HTTPStatusCode: 200,
				},
			},
		}

		responseBin, err := json.Marshal(response)
		assert.NoError(t, err)

		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusNotFound,
			Body:       ioutil.NopCloser(bytes.NewReader(responseBin)),
		}, nil).Once()
		payload := AccountRequest{
			FirstName:         &name,
			LastName:          &name,
			PersonEmail:       &email,
			PersonMobilePhone: &phoneNumber,
			RecordTypeID:      &recordTypeID,
			PersonBirthDate:   &dateBirth,
		}

		contact, errResponse := salesforceClient.CreateAccountComposite(span, payload)
		assert.Error(t, errResponse.Error)
		assert.Empty(t, contact)
	})

	t.Run("Create account error validation payload", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd"}`))),
		}, nil)
		payload := AccountRequest{
			Name: &name,
		}
		id, err := salesforceClient.CreateAccount(payload)

		assert.Error(t, err.Error)
		assert.Empty(t, id)
	})

	t.Run("Create account  error SendHTTPRequest", func(t *testing.T) {
		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{}, assert.AnError)
		payload := AccountRequest{
			Name:  &name,
			Phone: &phoneNumber,
		}
		id, err := salesforceClient.CreateAccount(payload)

		assert.Error(t, err.Error)
		assert.Empty(t, id)
	})

	t.Run("Create account error status", func(t *testing.T) {
		expectedError := fmt.Sprintf("%s-[%d] : %s", constants.StatusError, http.StatusInternalServerError, "[map[id:dasfasfasd]]")

		proxyMock := new(mocks.ProxyInterface)
		salesforceClient := NewSalesforceRequester("test", token)
		salesforceClient.Proxy = proxyMock
		proxyMock.On("SendHTTPRequest", mock.Anything, mock.Anything).Return(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`[{"id":"dasfasfasd"}]`))),
		}, nil)
		payload := AccountRequest{
			FirstName:       &name,
			LastName:        &name,
			PersonEmail:     &email,
			RecordTypeID:    &recordTypeID,
			PersonBirthDate: &phoneNumber,
		}
		id, err := salesforceClient.CreateAccount(payload)

		assert.Error(t, err.Error)
		assert.Empty(t, id)
		assert.Equal(t, expectedError, err.Error.Error())
	})
}

func TestSalesforceClient_GetContentVersionURL(t *testing.T) {
	type fields struct {
		AccessToken string
		APIVersion  string
		Proxy       proxy.ProxyInterface
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "success",
			fields: fields{
				APIVersion: "52",
			},

			want: "/services/data/v52.0/sobjects/ContentVersion",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cc := &SalesforceClient{
				AccessToken: tt.fields.AccessToken,
				APIVersion:  tt.fields.APIVersion,
				Proxy:       tt.fields.Proxy,
			}
			if got := cc.GetContentVersionURL(); got != tt.want {
				t.Errorf("SalesforceClient.GetContentVersionURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSalesforceClient_GetSearchURL(t *testing.T) {
	type fields struct {
		AccessToken string
		APIVersion  string
		Proxy       proxy.ProxyInterface
	}
	type args struct {
		query string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
	}{
		{
			name: "success",
			fields: fields{
				APIVersion: "52",
			},
			args: args{
				query: "test",
			},
			want: "/services/data/v52.0/query/?q=test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cc := &SalesforceClient{
				AccessToken: tt.fields.AccessToken,
				APIVersion:  tt.fields.APIVersion,
				Proxy:       tt.fields.Proxy,
			}
			if got := cc.GetSearchURL(tt.args.query); got != tt.want {
				t.Errorf("SalesforceClient.GetSearchURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSalesforceClient_GetDocumentLinkURL(t *testing.T) {
	type fields struct {
		AccessToken string
		APIVersion  string
		Proxy       proxy.ProxyInterface
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "success",
			fields: fields{
				APIVersion: "52",
			},
			want: "/services/data/v52.0/sobjects/ContentDocumentLink",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cc := &SalesforceClient{
				AccessToken: tt.fields.AccessToken,
				APIVersion:  tt.fields.APIVersion,
				Proxy:       tt.fields.Proxy,
			}
			if got := cc.GetDocumentLinkURL(); got != tt.want {
				t.Errorf("SalesforceClient.GetDocumentLinkURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
