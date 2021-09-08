// Code generated by mockery 2.7.4. DO NOT EDIT.

package manage

import (
	mock "github.com/stretchr/testify/mock"

	"yalochat.com/salesforce-integration/base/cache"
	chat "yalochat.com/salesforce-integration/base/clients/chat"
	helpers "yalochat.com/salesforce-integration/base/helpers"
	models "yalochat.com/salesforce-integration/base/models"
)

// ContextCache is an autogenerated mock type for the ContextCache type
type ContextCacheMock struct {
	mock.Mock
}

// RetrieveContext provides a mock function with given fields: userID
func (_m *ContextCacheMock) RetrieveContext(userID string) []cache.Context {
	ret := _m.Called(userID)

	var r0 []cache.Context
	if rf, ok := ret.Get(0).(func(string) []cache.Context); ok {
		r0 = rf(userID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]cache.Context)
		}
	}

	return r0
}

// StoreContext provides a mock function with given fields: _a0
func (_m *ContextCacheMock) StoreContext(_a0 cache.Context) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(cache.Context) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SalesforceServiceInterface is an autogenerated mock type for the SalesforceServiceInterface type
type SalesforceServiceInterface struct {
	mock.Mock
}

// CreatChat provides a mock function with given fields: contactName, organizationId, deploymentId, buttonId
func (_m *SalesforceServiceInterface) CreatChat(contactName string, organizationId string, deploymentId string, buttonId string) (*chat.SessionResponse, error) {
	ret := _m.Called(contactName, organizationId, deploymentId, buttonId)

	var r0 *chat.SessionResponse
	if rf, ok := ret.Get(0).(func(string, string, string, string) *chat.SessionResponse); ok {
		r0 = rf(contactName, organizationId, deploymentId, buttonId)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*chat.SessionResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string, string) error); ok {
		r1 = rf(contactName, organizationId, deploymentId, buttonId)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetMessages provides a mock function with given fields: affinityToken, sessionKey
func (_m *SalesforceServiceInterface) GetMessages(affinityToken string, sessionKey string) (*chat.MessagesResponse, *helpers.ErrorResponse) {
	ret := _m.Called(affinityToken, sessionKey)

	var r0 *chat.MessagesResponse
	if rf, ok := ret.Get(0).(func(string, string) *chat.MessagesResponse); ok {
		r0 = rf(affinityToken, sessionKey)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*chat.MessagesResponse)
		}
	}

	var r1 *helpers.ErrorResponse
	if rf, ok := ret.Get(1).(func(string, string) *helpers.ErrorResponse); ok {
		r1 = rf(affinityToken, sessionKey)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*helpers.ErrorResponse)
		}
	}

	return r0, r1
}

// GetOrCreateContact provides a mock function with given fields: name, email, phoneNumber
func (_m *SalesforceServiceInterface) GetOrCreateContact(name string, email string, phoneNumber string) (*models.SfcContact, error) {
	ret := _m.Called(name, email, phoneNumber)

	var r0 *models.SfcContact
	if rf, ok := ret.Get(0).(func(string, string, string) *models.SfcContact); ok {
		r0 = rf(name, email, phoneNumber)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.SfcContact)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string) error); ok {
		r1 = rf(name, email, phoneNumber)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SendMessage provides a mock function with given fields: _a0, _a1, _a2
func (_m *SalesforceServiceInterface) SendMessage(_a0 string, _a1 string, _a2 chat.MessagePayload) (bool, error) {
	ret := _m.Called(_a0, _a1, _a2)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string, string, chat.MessagePayload) bool); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, chat.MessagePayload) error); ok {
		r1 = rf(_a0, _a1, _a2)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
