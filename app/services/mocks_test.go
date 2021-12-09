// Code generated by mockery 2.7.4. DO NOT EDIT.

package services

import (
	mock "github.com/stretchr/testify/mock"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"

	//http "net/http"

	chat "yalochat.com/salesforce-integration/base/clients/chat"
	login "yalochat.com/salesforce-integration/base/clients/login"
	salesforce "yalochat.com/salesforce-integration/base/clients/salesforce"
	helpers "yalochat.com/salesforce-integration/base/helpers"
	models "yalochat.com/salesforce-integration/base/models"
)

// SfcChatInterface is an autogenerated mock type for the SfcChatInterface type
type SfcChatInterface struct {
	mock.Mock
}

// ChatEnd provides a mock function with given fields: affinityToken, sessionKey
func (_m *SfcChatInterface) ChatEnd(affinityToken string, sessionKey string) error {
	ret := _m.Called(affinityToken, sessionKey)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(affinityToken, sessionKey)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateChat provides a mock function with given fields: _a0, _a1, _a2, _a3
func (_m *SfcChatInterface) CreateChat(_a0 ddtrace.Span, _a1 string, _a2 string, _a3 chat.ChatRequest) (bool, error) {
	ret := _m.Called(_a0, _a1, _a2, _a3)

	var r0 bool
	if rf, ok := ret.Get(0).(func(ddtrace.Span, string, string, chat.ChatRequest) bool); ok {
		r0 = rf(_a0, _a1, _a2, _a3)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(ddtrace.Span, string, string, chat.ChatRequest) error); ok {
		r1 = rf(_a0, _a1, _a2, _a3)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreateSession provides a mock function with given fields: mainSpan
func (_m *SfcChatInterface) CreateSession(mainSpan ddtrace.Span) (*chat.SessionResponse, error) {
	ret := _m.Called(mainSpan)

	var r0 *chat.SessionResponse
	if rf, ok := ret.Get(0).(func(ddtrace.Span) *chat.SessionResponse); ok {
		r0 = rf(mainSpan)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*chat.SessionResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(ddtrace.Span) error); ok {
		r1 = rf(mainSpan)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetMessages provides a mock function with given fields: mainSpan, affinityToken, sessionKey
func (_m *SfcChatInterface) GetMessages(mainSpan ddtrace.Span, affinityToken string, sessionKey string) (*chat.MessagesResponse, *helpers.ErrorResponse) {
	ret := _m.Called(mainSpan, affinityToken, sessionKey)

	var r0 *chat.MessagesResponse
	if rf, ok := ret.Get(0).(func(ddtrace.Span, string, string) *chat.MessagesResponse); ok {
		r0 = rf(mainSpan, affinityToken, sessionKey)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*chat.MessagesResponse)
		}
	}

	var r1 *helpers.ErrorResponse
	if rf, ok := ret.Get(1).(func(ddtrace.Span, string, string) *helpers.ErrorResponse); ok {
		r1 = rf(mainSpan, affinityToken, sessionKey)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*helpers.ErrorResponse)
		}
	}

	return r0, r1
}

// ReconnectSession provides a mock function with given fields: affinityToken, sessionKey, offset
func (_m *SfcChatInterface) ReconnectSession(affinityToken string, sessionKey string, offset string) (*chat.MessagesResponse, error) {
	ret := _m.Called(affinityToken, sessionKey, offset)

	var r0 *chat.MessagesResponse
	if rf, ok := ret.Get(0).(func(string, string, string) *chat.MessagesResponse); ok {
		r0 = rf(affinityToken, sessionKey, offset)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*chat.MessagesResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string) error); ok {
		r1 = rf(affinityToken, sessionKey, offset)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SendMessage provides a mock function with given fields: _a0, _a1, _a2, _a3
func (_m *SfcChatInterface) SendMessage(_a0 ddtrace.Span, _a1 string, _a2 string, _a3 chat.MessagePayload) (bool, error) {
	ret := _m.Called(_a0, _a1, _a2, _a3)

	var r0 bool
	if rf, ok := ret.Get(0).(func(ddtrace.Span, string, string, chat.MessagePayload) bool); ok {
		r0 = rf(_a0, _a1, _a2, _a3)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(ddtrace.Span, string, string, chat.MessagePayload) error); ok {
		r1 = rf(_a0, _a1, _a2, _a3)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateToken provides a mock function with given fields: accessToken
func (_m *SfcChatInterface) UpdateToken(accessToken string) {
	_m.Called(accessToken)
}

// SaleforceInterface is an autogenerated mock type for the SaleforceInterface type
type SaleforceInterface struct {
	mock.Mock
}

// Composite provides a mock function with given fields: mainSpan, compositeRequest
func (_m *SaleforceInterface) Composite(mainSpan ddtrace.Span, compositeRequest salesforce.CompositeRequest) (salesforce.CompositeResponses, *helpers.ErrorResponse) {
	ret := _m.Called(mainSpan, compositeRequest)

	var r0 salesforce.CompositeResponses
	if rf, ok := ret.Get(0).(func(ddtrace.Span, salesforce.CompositeRequest) salesforce.CompositeResponses); ok {
		r0 = rf(mainSpan, compositeRequest)
	} else {
		r0 = ret.Get(0).(salesforce.CompositeResponses)
	}

	var r1 *helpers.ErrorResponse
	if rf, ok := ret.Get(1).(func(ddtrace.Span, salesforce.CompositeRequest) *helpers.ErrorResponse); ok {
		r1 = rf(mainSpan, compositeRequest)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*helpers.ErrorResponse)
		}
	}

	return r0, r1
}

// CreateAccount provides a mock function with given fields: payload
func (_m *SaleforceInterface) CreateAccount(payload salesforce.AccountRequest) (string, *helpers.ErrorResponse) {
	ret := _m.Called(payload)

	var r0 string
	if rf, ok := ret.Get(0).(func(salesforce.AccountRequest) string); ok {
		r0 = rf(payload)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 *helpers.ErrorResponse
	if rf, ok := ret.Get(1).(func(salesforce.AccountRequest) *helpers.ErrorResponse); ok {
		r1 = rf(payload)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*helpers.ErrorResponse)
		}
	}

	return r0, r1
}

// CreateAccountComposite provides a mock function with given fields: mainSpan, payload
func (_m *SaleforceInterface) CreateAccountComposite(mainSpan ddtrace.Span, payload salesforce.AccountRequest) (*models.SfcAccount, *helpers.ErrorResponse) {
	ret := _m.Called(mainSpan, payload)

	var r0 *models.SfcAccount
	if rf, ok := ret.Get(0).(func(ddtrace.Span, salesforce.AccountRequest) *models.SfcAccount); ok {
		r0 = rf(mainSpan, payload)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.SfcAccount)
		}
	}

	var r1 *helpers.ErrorResponse
	if rf, ok := ret.Get(1).(func(ddtrace.Span, salesforce.AccountRequest) *helpers.ErrorResponse); ok {
		r1 = rf(mainSpan, payload)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*helpers.ErrorResponse)
		}
	}

	return r0, r1
}

// CreateCase provides a mock function with given fields: mainSpan, payload
func (_m *SaleforceInterface) CreateCase(mainSpan ddtrace.Span, payload interface{}) (string, *helpers.ErrorResponse) {
	ret := _m.Called(mainSpan, payload)

	var r0 string
	if rf, ok := ret.Get(0).(func(ddtrace.Span, interface{}) string); ok {
		r0 = rf(mainSpan, payload)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 *helpers.ErrorResponse
	if rf, ok := ret.Get(1).(func(ddtrace.Span, interface{}) *helpers.ErrorResponse); ok {
		r1 = rf(mainSpan, payload)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*helpers.ErrorResponse)
		}
	}

	return r0, r1
}

// CreateContact provides a mock function with given fields: mainSpan, payload
func (_m *SaleforceInterface) CreateContact(mainSpan ddtrace.Span, payload interface{}) (string, *helpers.ErrorResponse) {
	ret := _m.Called(mainSpan, payload)

	var r0 string
	if rf, ok := ret.Get(0).(func(ddtrace.Span, interface{}) string); ok {
		r0 = rf(mainSpan, payload)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 *helpers.ErrorResponse
	if rf, ok := ret.Get(1).(func(ddtrace.Span, interface{}) *helpers.ErrorResponse); ok {
		r1 = rf(mainSpan, payload)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*helpers.ErrorResponse)
		}
	}

	return r0, r1
}

// CreateContentVersion provides a mock function with given fields: _a0
func (_m *SaleforceInterface) CreateContentVersion(_a0 salesforce.ContentVersionPayload) (string, error) {
	ret := _m.Called(_a0)

	var r0 string
	if rf, ok := ret.Get(0).(func(salesforce.ContentVersionPayload) string); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(salesforce.ContentVersionPayload) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetContentVersionURL provides a mock function with given fields:
func (_m *SaleforceInterface) GetContentVersionURL() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetDocumentLinkURL provides a mock function with given fields:
func (_m *SaleforceInterface) GetDocumentLinkURL() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetSearchURL provides a mock function with given fields: query
func (_m *SaleforceInterface) GetSearchURL(query string) string {
	ret := _m.Called(query)

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(query)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// LinkDocumentToCase provides a mock function with given fields: _a0
func (_m *SaleforceInterface) LinkDocumentToCase(_a0 salesforce.LinkDocumentPayload) (string, error) {
	ret := _m.Called(_a0)

	var r0 string
	if rf, ok := ret.Get(0).(func(salesforce.LinkDocumentPayload) string); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(salesforce.LinkDocumentPayload) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Search provides a mock function with given fields: _a0
func (_m *SaleforceInterface) Search(_a0 string) (*salesforce.SearchResponse, *helpers.ErrorResponse) {
	ret := _m.Called(_a0)

	var r0 *salesforce.SearchResponse
	if rf, ok := ret.Get(0).(func(string) *salesforce.SearchResponse); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*salesforce.SearchResponse)
		}
	}

	var r1 *helpers.ErrorResponse
	if rf, ok := ret.Get(1).(func(string) *helpers.ErrorResponse); ok {
		r1 = rf(_a0)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*helpers.ErrorResponse)
		}
	}

	return r0, r1
}

// SearchAccount provides a mock function with given fields: _a0
func (_m *SaleforceInterface) SearchAccount(_a0 string) (*models.SfcAccount, *helpers.ErrorResponse) {
	ret := _m.Called(_a0)

	var r0 *models.SfcAccount
	if rf, ok := ret.Get(0).(func(string) *models.SfcAccount); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.SfcAccount)
		}
	}

	var r1 *helpers.ErrorResponse
	if rf, ok := ret.Get(1).(func(string) *helpers.ErrorResponse); ok {
		r1 = rf(_a0)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*helpers.ErrorResponse)
		}
	}

	return r0, r1
}

// SearchContact provides a mock function with given fields: _a0
func (_m *SaleforceInterface) SearchContact(_a0 string) (*models.SfcContact, *helpers.ErrorResponse) {
	ret := _m.Called(_a0)

	var r0 *models.SfcContact
	if rf, ok := ret.Get(0).(func(string) *models.SfcContact); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.SfcContact)
		}
	}

	var r1 *helpers.ErrorResponse
	if rf, ok := ret.Get(1).(func(string) *helpers.ErrorResponse); ok {
		r1 = rf(_a0)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*helpers.ErrorResponse)
		}
	}

	return r0, r1
}

// SearchContactComposite provides a mock function with given fields: mainSpan, email, phoneNumber
func (_m *SaleforceInterface) SearchContactComposite(mainSpan ddtrace.Span, email string, phoneNumber string) (*models.SfcContact, *helpers.ErrorResponse) {
	ret := _m.Called(mainSpan, email, phoneNumber)

	var r0 *models.SfcContact
	if rf, ok := ret.Get(0).(func(ddtrace.Span, string, string) *models.SfcContact); ok {
		r0 = rf(mainSpan, email, phoneNumber)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.SfcContact)
		}
	}

	var r1 *helpers.ErrorResponse
	if rf, ok := ret.Get(1).(func(ddtrace.Span, string, string) *helpers.ErrorResponse); ok {
		r1 = rf(mainSpan, email, phoneNumber)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*helpers.ErrorResponse)
		}
	}

	return r0, r1
}

// SearchDocumentID provides a mock function with given fields: _a0
func (_m *SaleforceInterface) SearchDocumentID(_a0 string) (string, error) {
	ret := _m.Called(_a0)

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SearchID provides a mock function with given fields: _a0
func (_m *SaleforceInterface) SearchID(_a0 string) (string, error) {
	ret := _m.Called(_a0)

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateToken provides a mock function with given fields: accessToken
func (_m *SaleforceInterface) UpdateToken(accessToken string) {
	_m.Called(accessToken)
}

// SfcLoginInterface is an autogenerated mock type for the SfcLoginInterface type
type SfcLoginInterface struct {
	mock.Mock
}

// GetToken provides a mock function with given fields: _a0
func (_m *SfcLoginInterface) GetToken(_a0 login.TokenPayload) (string, error) {
	ret := _m.Called(_a0)

	var r0 string
	if rf, ok := ret.Get(0).(func(login.TokenPayload) string); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(login.TokenPayload) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
