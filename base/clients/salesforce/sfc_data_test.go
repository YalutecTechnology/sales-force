package salesforce

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"yalochat.com/salesforce-integration/base/clients/proxy"
	"yalochat.com/salesforce-integration/base/models"
)

const (
	caseURL = "test"
	token   = "token"
)

var (
	name         string = "name"
	phoneNumber  string = "111111"
	email        string = "email@example"
	recordTypeID string = "reccordTypeId"
	dateBirth    string = "2021-10-05T08:12:00"
)

func TestCaseClient_CreateContentVersion(t *testing.T) {

	t.Run("Create ContentVersion Succesfull", func(t *testing.T) {
		mock := &proxy.Mock{}
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
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
		mock := &proxy.Mock{}
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
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
		mock := &proxy.Mock{}
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{}, assert.AnError)
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

	t.Run("Create ContentVersion error status", func(t *testing.T) {
		mock := &proxy.Mock{}
		salesforceClient := NewSalesforceRequester("test", token)
		salesforceClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
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
		assert.Empty(t, id)
	})
}

func TestCaseClient_SearchDocumentID(t *testing.T) {

	t.Run("SearchDocumentID Succesfull", func(t *testing.T) {
		mock := &proxy.Mock{}
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
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
		mock := &proxy.Mock{}
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd"}`))),
		}, nil)
		id, err := salesforceClient.SearchDocumentID("")

		assert.Error(t, err)
		assert.Empty(t, id)
	})

	t.Run("SearchDocumentID error SendHTTPRequest", func(t *testing.T) {
		mock := &proxy.Mock{}
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{}, assert.AnError)

		id, err := salesforceClient.SearchDocumentID("query")

		assert.Error(t, err)
		assert.Empty(t, id)
	})

	t.Run("SearchDocumentID error status", func(t *testing.T) {
		mock := &proxy.Mock{}
		salesforceClient := NewSalesforceRequester("test", token)
		salesforceClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd"}`))),
		}, nil)
		id, err := salesforceClient.SearchDocumentID("query")
		assert.Error(t, err)
		assert.Empty(t, id)
	})
}

func TestCaseClient_Search(t *testing.T) {

	t.Run("Search Succesfull", func(t *testing.T) {
		mock := &proxy.Mock{}
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
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
		mock := &proxy.Mock{}
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd"}`))),
		}, nil)
		id, err := salesforceClient.Search("")

		assert.Error(t, err.Error)
		assert.Empty(t, id)
	})

	t.Run("Search error SendHTTPRequest", func(t *testing.T) {
		mock := &proxy.Mock{}
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{}, assert.AnError)

		id, err := salesforceClient.Search("query")

		assert.Error(t, err.Error)
		assert.Empty(t, id)
	})

	t.Run("SearchDocumentID error status", func(t *testing.T) {
		mock := &proxy.Mock{}
		salesforceClient := NewSalesforceRequester("test", token)
		salesforceClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd"}`))),
		}, nil)
		id, err := salesforceClient.Search("query")
		assert.Error(t, err.Error)
		assert.Empty(t, id)
	})
}

func TestCaseClient_SearchId(t *testing.T) {

	t.Run("SearchId Succesfull", func(t *testing.T) {
		mock := &proxy.Mock{}
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
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
		mock := &proxy.Mock{}
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd"}`))),
		}, nil)
		id, err := salesforceClient.SearchID("")

		assert.Error(t, err)
		assert.Empty(t, id)
	})

	t.Run("Search error SendHTTPRequest", func(t *testing.T) {
		mock := &proxy.Mock{}
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{}, assert.AnError)

		id, err := salesforceClient.SearchID("query")

		assert.Error(t, err)
		assert.Empty(t, id)
	})

	t.Run("SearchDocumentID error status", func(t *testing.T) {
		mock := &proxy.Mock{}
		salesforceClient := NewSalesforceRequester("test", token)
		salesforceClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd"}`))),
		}, nil)
		id, err := salesforceClient.SearchID("query")
		assert.Error(t, err)
		assert.Empty(t, id)
	})
}

func TestCaseClient_SearchContact(t *testing.T) {

	t.Run("SearchContact Succesfull", func(t *testing.T) {
		mock := &proxy.Mock{}
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
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
		mock := &proxy.Mock{}
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd"}`))),
		}, nil)
		contact, err := salesforceClient.SearchContact("")

		assert.Error(t, err.Error)
		assert.Empty(t, contact)
	})

	t.Run("Search error SendHTTPRequest", func(t *testing.T) {
		mock := &proxy.Mock{}
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{}, assert.AnError)

		id, err := salesforceClient.SearchContact("query")

		assert.Error(t, err.Error)
		assert.Empty(t, id)
	})

	t.Run("SearchContact error status", func(t *testing.T) {
		mock := &proxy.Mock{}
		salesforceClient := NewSalesforceRequester("test", token)
		salesforceClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd"}`))),
		}, nil)
		id, err := salesforceClient.SearchContact("query")
		assert.Error(t, err.Error)
		assert.Empty(t, id)
	})
}

func TestCaseClient_SearchAccount(t *testing.T) {

	t.Run("SearchAccount Succesfull", func(t *testing.T) {
		mock := &proxy.Mock{}
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
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
		mock := &proxy.Mock{}
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd"}`))),
		}, nil)
		account, err := salesforceClient.SearchAccount("")

		assert.Error(t, err.Error)
		assert.Empty(t, account)
	})

	t.Run("Search error SendHTTPRequest", func(t *testing.T) {
		mock := &proxy.Mock{}
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{}, assert.AnError)

		account, err := salesforceClient.SearchAccount("query")

		assert.Error(t, err.Error)
		assert.Empty(t, account)
	})

	t.Run("SearchContact error status", func(t *testing.T) {
		mock := &proxy.Mock{}
		salesforceClient := NewSalesforceRequester("test", token)
		salesforceClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd"}`))),
		}, nil)
		account, err := salesforceClient.SearchAccount("query")
		assert.Error(t, err.Error)
		assert.Empty(t, account)
	})
}

func TestCaseClient_LinkDocumentToCase(t *testing.T) {

	t.Run("LinkDocumentToCase Succesfull", func(t *testing.T) {
		mock := &proxy.Mock{}
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
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
		mock := &proxy.Mock{}
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
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
		mock := &proxy.Mock{}
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{}, assert.AnError)
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
		mock := &proxy.Mock{}
		salesforceClient := NewSalesforceRequester("test", token)
		salesforceClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd"}`))),
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
	})
}

func TestCaseClient_CreateCase(t *testing.T) {

	t.Run("Create case Succesfull", func(t *testing.T) {
		mock := &proxy.Mock{}
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
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
		id, err := salesforceClient.CreateCase(payload)

		if err != nil {
			t.Fatalf("Expected nil error, but retrieved this %#v", err)
		}

		assert.Equal(t, "dasfasfasd", id)
	})

	t.Run("Create case  error SendHTTPRequest", func(t *testing.T) {
		mock := &proxy.Mock{}
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{}, assert.AnError)
		payload := CaseRequest{
			ContactID:   "contact id",
			Status:      "New",
			Origin:      "Web",
			Subject:     "test",
			Priority:    "Medium",
			IsEscalated: false,
			Description: "context",
		}
		id, err := salesforceClient.CreateCase(payload)

		assert.Error(t, err.Error)
		assert.Empty(t, id)
	})

	// t.Run("Create case unmarsahal response", func(t *testing.T) {
	// 	mock := &proxy.Mock{}
	// 	salesforceClient := NewSalesforceRequester(caseURL,token)
	// 	salesforceClient.Proxy = mock
	// 	mock.On("SendHTTPRequest").Return(&http.Response{
	// 		StatusCode: http.StatusOK,
	// 		Body:       ioutil.NopCloser(bytes.NewReader([]byte(`error`))),
	// 	}, nil)
	// 	payload := CaseRequest{
	// 		ContactID:   "contact id",
	// 		Status:      "New",
	// 		Origin:      "Web",
	// 		Subject:     "test",
	// 		Priority:    "Medium",
	// 		IsEscalated: false,
	// 		Description: "context",
	// 	}
	// 	id, err := salesforceClient.CreateCase(payload)

	// 	assert.Error(t, err)
	// 	assert.Empty(t, id)
	// })

	t.Run("Create case error status", func(t *testing.T) {
		mock := &proxy.Mock{}
		salesforceClient := NewSalesforceRequester("test", token)
		salesforceClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusInternalServerError,
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
		id, err := salesforceClient.CreateCase(payload)

		assert.Error(t, err.Error)
		assert.Empty(t, id)
	})
}

func TestCaseClient_CreateContact(t *testing.T) {

	t.Run("Create contact Succesfull", func(t *testing.T) {
		mock := &proxy.Mock{}
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusCreated,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd"}`))),
		}, nil).Once()
		payload := ContactRequest{
			FirstName:   "firstname",
			LastName:    "lasrname",
			MobilePhone: "11111111111",
			Email:       "test@mail.com",
		}
		id, err := salesforceClient.CreateContact(payload)

		if err != nil {
			t.Fatalf("Expected nil error, but retrieved this %#v", err)
		}

		assert.Equal(t, "dasfasfasd", id)
	})

	t.Run("Create contact  error validation payload", func(t *testing.T) {
		mock := &proxy.Mock{}
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd"}`))),
		}, nil)
		payload := ContactRequest{
			FirstName:   "firstname",
			LastName:    "lasrname",
			MobilePhone: "11111111111",
		}
		id, err := salesforceClient.CreateContact(payload)

		assert.Error(t, err.Error)
		assert.Empty(t, id)
	})

	t.Run("Create contact  error SendHTTPRequest", func(t *testing.T) {
		mock := &proxy.Mock{}
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{}, assert.AnError)
		payload := ContactRequest{
			FirstName:   "firstname",
			LastName:    "lasrname",
			MobilePhone: "11111111111",
			Email:       "test@mail.com",
		}
		id, err := salesforceClient.CreateContact(payload)

		assert.Error(t, err.Error)
		assert.Empty(t, id)
	})

	// t.Run("Create contact unmarsahal response", func(t *testing.T) {
	// 	mock := &proxy.Mock{}
	// 	salesforceClient := NewSalesforceRequester(caseURL,token)
	// 	salesforceClient.Proxy = mock
	// 	mock.On("SendHTTPRequest").Return(&http.Response{
	// 		StatusCode: http.StatusOK,
	// 		Body:       ioutil.NopCloser(bytes.NewReader([]byte(`error`))),
	// 	}, nil)
	// 	payload := ContactRequest{
	// 	FirstName: "firstname",
	// 	LastName: "lasrname",
	// 	MobilePhone: "11111111111",
	// 	Email: "test@mail.com",
	// }
	// 	id, err := salesforceClient.CreateContact(payload)

	// 	assert.Error(t, err)
	// 	assert.Empty(t, id)
	// })

	t.Run("Create contact error status", func(t *testing.T) {
		mock := &proxy.Mock{}
		salesforceClient := NewSalesforceRequester("test", token)
		salesforceClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd"}`))),
		}, nil)
		payload := ContactRequest{
			FirstName:   "firstname",
			LastName:    "lasrname",
			MobilePhone: "11111111111",
			Email:       "test@mail.com",
		}
		id, err := salesforceClient.CreateContact(payload)

		assert.Error(t, err.Error)
		assert.Empty(t, id)
	})
}

func TestCaseClient_CreateAccount(t *testing.T) {

	t.Run("Create account Succesfull", func(t *testing.T) {
		mock := &proxy.Mock{}
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
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
		mock := &proxy.Mock{}
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
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
		mock := &proxy.Mock{}
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{}, assert.AnError)
		payload := AccountRequest{
			Name:  &name,
			Phone: &phoneNumber,
		}
		id, err := salesforceClient.CreateAccount(payload)

		assert.Error(t, err.Error)
		assert.Empty(t, id)
	})

	t.Run("Create account error status", func(t *testing.T) {
		mock := &proxy.Mock{}
		salesforceClient := NewSalesforceRequester("test", token)
		salesforceClient.Proxy = mock
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd"}`))),
		}, nil)
		payload := AccountRequest{
			Name:  &name,
			Phone: &phoneNumber,
		}
		id, err := salesforceClient.CreateAccount(payload)

		assert.Error(t, err.Error)
		assert.Empty(t, id)
	})
}

func TestCaseClient_Composite(t *testing.T) {

	t.Run("Create Composite Succesfull", func(t *testing.T) {
		mock := &proxy.Mock{}
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = mock

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

		mock.On("SendHTTPRequest").Return(&http.Response{
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
		response, errResponse := salesforceClient.Composite(payload)

		assert.Nil(t, errResponse)
		assert.Equal(t, expected, response)
	})

	t.Run("Create Composite error validation", func(t *testing.T) {
		mock := &proxy.Mock{}
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = mock

		payload := CompositeRequest{
			AllOrNone:          true,
			CollateSubrequests: true,
		}
		response, err := salesforceClient.Composite(payload)

		assert.Error(t, err.Error)
		assert.Empty(t, response)
	})

	t.Run("Create Composite error status", func(t *testing.T) {
		mock := &proxy.Mock{}
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = mock

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

		mock.On("SendHTTPRequest").Return(&http.Response{
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
		response, errResponse := salesforceClient.Composite(payload)

		assert.Error(t, errResponse.Error)
		assert.Empty(t, response)
	})
}

func TestCaseClient_UpdateToken(t *testing.T) {

	t.Run("Update token Succesfull", func(t *testing.T) {
		tokenExpected := "14525542211224"
		salesforceClient := NewSalesforceRequester(caseURL, token)

		salesforceClient.UpdateToken(tokenExpected)
		assert.Equal(t, tokenExpected, salesforceClient.AccessToken)
	})
}

func TestCaseClient_SearchContactComposite(t *testing.T) {

	t.Run("SearchContactComposite Succesfull", func(t *testing.T) {
		mock := &proxy.Mock{}
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = mock

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
		mock.On("SendHTTPRequest").Return(&http.Response{
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

		contact, errResponse := salesforceClient.SearchContactComposite(email, phoneNumber)

		assert.Nil(t, errResponse)
		assert.Equal(t, contactExpected, contact)
	})

	t.Run("SearchContactComposite notFount", func(t *testing.T) {
		mock := &proxy.Mock{}
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = mock

		response := CompositeResponses{
			CompositeResponse: []CompositeResponse{},
		}

		responseBin, err := json.Marshal(response)
		assert.NoError(t, err)
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader(responseBin)),
		}, nil)

		contact, errResponse := salesforceClient.SearchContactComposite(email, phoneNumber)

		assert.NotNil(t, errResponse)
		assert.Empty(t, contact)
	})

	t.Run("SearchContactComposite request error", func(t *testing.T) {
		mock := &proxy.Mock{}
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = mock

		response := CompositeResponses{
			CompositeResponse: []CompositeResponse{},
		}

		responseBin, err := json.Marshal(response)
		assert.NoError(t, err)
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusNotFound,
			Body:       ioutil.NopCloser(bytes.NewReader(responseBin)),
		}, nil)

		contact, errResponse := salesforceClient.SearchContactComposite(email, phoneNumber)

		assert.NotNil(t, errResponse)
		assert.Empty(t, contact)
	})
}

func TestCaseClient_CreateAccountComposite(t *testing.T) {
	t.Run("Create account Succesfull", func(t *testing.T) {
		mock := &proxy.Mock{}
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = mock

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
								Id:                "0032300000Qzu1iAAB",
								FirstName:         name,
								LastName:          name,
								PersonEmail:       email,
								PersonMobilePhone: phoneNumber,
								PersonContactID:   "contactID",
							},
						},
					},
				},
			},
		}

		responseBin, err := json.Marshal(response)
		assert.NoError(t, err)

		mock.On("SendHTTPRequest").Return(&http.Response{
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
		contact, errResponse := salesforceClient.CreateAccountComposite(payload)
		assert.Nil(t, errResponse)
		assert.Equal(t, contactExpected, contact)
	})

	t.Run("Create account error account not found ", func(t *testing.T) {
		mock := &proxy.Mock{}
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = mock

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
						TotalSize: 0,
						Done:      true,
						Records:   []recordResponse{},
					},
				},
			},
		}

		responseBin, err := json.Marshal(response)
		assert.NoError(t, err)

		mock.On("SendHTTPRequest").Return(&http.Response{
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

		contact, errResponse := salesforceClient.CreateAccountComposite(payload)
		assert.Error(t, errResponse.Error)
		assert.Empty(t, contact)
	})

	t.Run("Create account error account not found ", func(t *testing.T) {
		mock := &proxy.Mock{}
		salesforceClient := NewSalesforceRequester(caseURL, token)
		salesforceClient.Proxy = mock

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
						TotalSize: 0,
						Done:      true,
						Records:   []recordResponse{},
					},
				},
			},
		}

		responseBin, err := json.Marshal(response)
		assert.NoError(t, err)

		mock.On("SendHTTPRequest").Return(&http.Response{
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

		contact, errResponse := salesforceClient.CreateAccountComposite(payload)
		assert.Error(t, errResponse.Error)
		assert.Empty(t, contact)
	})

	// t.Run("Create account  error validation payload", func(t *testing.T) {
	// 	mock := &proxy.Mock{}
	// 	salesforceClient := NewSalesforceRequester(caseURL, token)
	// 	salesforceClient.Proxy = mock
	// 	mock.On("SendHTTPRequest").Return(&http.Response{
	// 		StatusCode: http.StatusOK,
	// 		Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd"}`))),
	// 	}, nil)
	// 	payload := AccountRequest{
	// 		Name: &name,
	// 	}
	// 	id, err := salesforceClient.CreateAccount(payload)

	// 	assert.Error(t, err.Error)
	// 	assert.Empty(t, id)
	// })

	// t.Run("Create account  error SendHTTPRequest", func(t *testing.T) {
	// 	mock := &proxy.Mock{}
	// 	salesforceClient := NewSalesforceRequester(caseURL, token)
	// 	salesforceClient.Proxy = mock
	// 	mock.On("SendHTTPRequest").Return(&http.Response{}, assert.AnError)
	// 	payload := AccountRequest{
	// 		Name:  &name,
	// 		Phone: &phoneNumber,
	// 	}
	// 	id, err := salesforceClient.CreateAccount(payload)

	// 	assert.Error(t, err.Error)
	// 	assert.Empty(t, id)
	// })

	// t.Run("Create account error status", func(t *testing.T) {
	// 	mock := &proxy.Mock{}
	// 	salesforceClient := NewSalesforceRequester("test", token)
	// 	salesforceClient.Proxy = mock
	// 	mock.On("SendHTTPRequest").Return(&http.Response{
	// 		StatusCode: http.StatusInternalServerError,
	// 		Body:       ioutil.NopCloser(bytes.NewReader([]byte(`{"id":"dasfasfasd"}`))),
	// 	}, nil)
	// 	payload := AccountRequest{
	// 		Name:  &name,
	// 		Phone: &phoneNumber,
	// 	}
	// 	id, err := salesforceClient.CreateAccount(payload)

	// 	assert.Error(t, err.Error)
	// 	assert.Empty(t, id)
	// })
}
