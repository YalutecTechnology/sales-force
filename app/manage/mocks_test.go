// Code generated by mockery 2.7.4. DO NOT EDIT.

package manage

import (
	mock "github.com/stretchr/testify/mock"

	"yalochat.com/salesforce-integration/base/cache"
	chat "yalochat.com/salesforce-integration/base/clients/chat"
	integrations "yalochat.com/salesforce-integration/base/clients/integrations"
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

// BotRunnerInterface is an autogenerated mock type for the BotRunnerInterface type
type BotRunnerInterface struct {
	mock.Mock
}

// SendTo provides a mock function with given fields: object
func (_m *BotRunnerInterface) SendTo(object map[string]interface{}) (bool, error) {
	ret := _m.Called(object)

	var r0 bool
	if rf, ok := ret.Get(0).(func(map[string]interface{}) bool); ok {
		r0 = rf(object)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(map[string]interface{}) error); ok {
		r1 = rf(object)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SalesforceServiceInterface is an autogenerated mock type for the SalesforceServiceInterface type
type SalesforceServiceInterface struct {
	mock.Mock
}

// CreatCase provides a mock function with given fields: contactID, description, subject, origin, ownerID, extraData
func (_m *SalesforceServiceInterface) CreatCase(contactID string, description string, subject string, origin string, ownerID string, extraData map[string]interface{}) (string, error) {
	ret := _m.Called(contactID, description, subject, origin, ownerID, extraData)

	var r0 string
	if rf, ok := ret.Get(0).(func(string, string, string, string, string, map[string]interface{}) string); ok {
		r0 = rf(contactID, description, subject, origin, ownerID, extraData)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string, string, string, map[string]interface{}) error); ok {
		r1 = rf(contactID, description, subject, origin, ownerID, extraData)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreatChat provides a mock function with given fields: contactName, organizationID, deploymentID, buttonID, caseID, contactID
func (_m *SalesforceServiceInterface) CreatChat(contactName string, organizationID string, deploymentID string, buttonID string, caseID string, contactID string) (*chat.SessionResponse, error) {
	ret := _m.Called(contactName, organizationID, deploymentID, buttonID, caseID, contactID)

	var r0 *chat.SessionResponse
	if rf, ok := ret.Get(0).(func(string, string, string, string, string, string) *chat.SessionResponse); ok {
		r0 = rf(contactName, organizationID, deploymentID, buttonID, caseID, contactID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*chat.SessionResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string, string, string, string, string) error); ok {
		r1 = rf(contactName, organizationID, deploymentID, buttonID, caseID, contactID)
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

// InsertImageInCase provides a mock function with given fields: uri, title, mimeType, caseID
func (_m *SalesforceServiceInterface) InsertImageInCase(uri string, title string, mimeType string, caseID string) error {
	ret := _m.Called(uri, title, mimeType, caseID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, string, string) error); ok {
		r0 = rf(uri, title, mimeType, caseID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// EndChat provides a mock function with given fields: affinityToken, sessionKey
func (_m *SalesforceServiceInterface) EndChat(affinityToken string, sessionKey string) error {
	ret := _m.Called(affinityToken, sessionKey)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(affinityToken, sessionKey)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SearchContactComposite provides a mock function with given fields: email, phoneNumber
func (_m *SalesforceServiceInterface) SearchContactComposite(email string, phoneNumber string) (*models.SfcContact, *helpers.ErrorResponse) {
	ret := _m.Called(email, phoneNumber)

	var r0 *models.SfcContact
	if rf, ok := ret.Get(0).(func(string, string) *models.SfcContact); ok {
		r0 = rf(email, phoneNumber)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.SfcContact)
		}
	}

	var r1 *helpers.ErrorResponse
	if rf, ok := ret.Get(1).(func(string, string) *helpers.ErrorResponse); ok {
		r1 = rf(email, phoneNumber)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*helpers.ErrorResponse)
		}
	}

	return r0, r1
}

// RefreshToken provides a mock function with given fields:
func (_m *SalesforceServiceInterface) RefreshToken() {
	_m.Called()
}

// InterconnectionCache is an autogenerated mock type for the InterconnectionCache type
type InterconnectionCache struct {
	mock.Mock
}

// DeleteAllInterconnections provides a mock function with given fields:
func (_m *InterconnectionCache) DeleteAllInterconnections() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteInterconnection provides a mock function with given fields: _a0
func (_m *InterconnectionCache) DeleteInterconnection(_a0 cache.Interconnection) (bool, error) {
	ret := _m.Called(_a0)

	var r0 bool
	if rf, ok := ret.Get(0).(func(cache.Interconnection) bool); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(cache.Interconnection) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RetrieveAllInterconnections provides a mock function with given fields: client
func (_m *InterconnectionCache) RetrieveAllInterconnections(client string) *[]cache.Interconnection {
	ret := _m.Called(client)

	var r0 *[]cache.Interconnection
	if rf, ok := ret.Get(0).(func(string) *[]cache.Interconnection); ok {
		r0 = rf(client)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*[]cache.Interconnection)
		}
	}

	return r0
}

// RetrieveInterconnection provides a mock function with given fields: _a0
func (_m *InterconnectionCache) RetrieveInterconnection(_a0 cache.Interconnection) (*cache.Interconnection, error) {
	ret := _m.Called(_a0)

	var r0 *cache.Interconnection
	if rf, ok := ret.Get(0).(func(cache.Interconnection) *cache.Interconnection); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*cache.Interconnection)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(cache.Interconnection) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RetrieveInterconnectionActiveByUserId provides a mock function with given fields: userId
func (_m *InterconnectionCache) RetrieveInterconnectionActiveByUserId(userId string) *cache.Interconnection {
	ret := _m.Called(userId)

	var r0 *cache.Interconnection
	if rf, ok := ret.Get(0).(func(string) *cache.Interconnection); ok {
		r0 = rf(userId)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*cache.Interconnection)
		}
	}

	return r0
}

// StoreInterconnection provides a mock function with given fields: _a0
func (_m *InterconnectionCache) StoreInterconnection(_a0 cache.Interconnection) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(cache.Interconnection) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// IntegrationInterface is an autogenerated mock type for the IntegrationInterface type
type IntegrationInterface struct {
	mock.Mock
}

// SendMessage provides a mock function with given fields: messagePayload, provider
func (_m *IntegrationInterface) SendMessage(messagePayload interface{}, provider string) (*integrations.SendMessageResponse, error) {
	ret := _m.Called(messagePayload, provider)

	var r0 *integrations.SendMessageResponse
	if rf, ok := ret.Get(0).(func(interface{}, string) *integrations.SendMessageResponse); ok {
		r0 = rf(messagePayload, provider)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*integrations.SendMessageResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(interface{}, string) error); ok {
		r1 = rf(messagePayload, provider)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// WebhookRegister provides a mock function with given fields: HealthcheckPayload
func (_m *IntegrationInterface) WebhookRegister(HealthcheckPayload integrations.HealthcheckPayload) (*integrations.HealthcheckResponse, error) {
	ret := _m.Called(HealthcheckPayload)

	var r0 *integrations.HealthcheckResponse
	if rf, ok := ret.Get(0).(func(integrations.HealthcheckPayload) *integrations.HealthcheckResponse); ok {
		r0 = rf(HealthcheckPayload)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*integrations.HealthcheckResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(integrations.HealthcheckPayload) error); ok {
		r1 = rf(HealthcheckPayload)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// WebhookRemove provides a mock function with given fields: removeWebhookPayload
func (_m *IntegrationInterface) WebhookRemove(removeWebhookPayload integrations.RemoveWebhookPayload) (bool, error) {
	ret := _m.Called(removeWebhookPayload)

	var r0 bool
	if rf, ok := ret.Get(0).(func(integrations.RemoveWebhookPayload) bool); ok {
		r0 = rf(removeWebhookPayload)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(integrations.RemoveWebhookPayload) error); ok {
		r1 = rf(removeWebhookPayload)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// IMessageCache is an autogenerated mock type for the IMessageCache type
type IMessageCache struct {
	mock.Mock
}

// IsRepeatedMessage provides a mock function with given fields: _a0
func (_m *IMessageCache) IsRepeatedMessage(_a0 string) bool {
	ret := _m.Called(_a0)

	var r0 bool
	if rf, ok := ret.Get(0).(func(string) bool); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// StudioNGInterface is an autogenerated mock type for the StudioNGInterface type
type StudioNGInterface struct {
	mock.Mock
}

// SendTo provides a mock function with given fields: state, userID
func (_m *StudioNGInterface) SendTo(state string, userID string) error {
	ret := _m.Called(state, userID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(state, userID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
