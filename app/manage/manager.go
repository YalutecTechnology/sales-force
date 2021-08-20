package manage

import (
	"github.com/sirupsen/logrus"
	"yalochat.com/salesforce-integration/base/cache"
	"yalochat.com/salesforce-integration/base/clients/chat"
	"yalochat.com/salesforce-integration/base/clients/login"
	"yalochat.com/salesforce-integration/base/clients/proxy"
	"yalochat.com/salesforce-integration/base/clients/salesforce"
)

// Manager controls the process of the app
type Manager struct {
	clientName     string
	sfcLoginClient *login.SfcLoginClient
	sfcChatClient  *chat.SfcChatClient
	sfcClient      *salesforce.SalesforceClient
}

// ManagerOptions holds configurations for the interactions manager
type ManagerOptions struct {
	AppName           string
	RedisOptions      cache.RedisOptions
	SfcClientId       string
	SfcClientSecret   string
	SfcUsername       string
	SfcPassword       string
	SfcSecurityToken  string
	SfcBaseUrl        string
	SfcChatUrl        string
	SfcLoginUrl       string
	SfcApiVersion     string
	SfcOrganizationId string
	SfcDeploymentId   string
	SfcButtonId       string
	SfcOwnerId        string
}

// CreateManager retrieves an agents manager
func CreateManager(config *ManagerOptions) *Manager {
	_, err := cache.NewRedisCache(&config.RedisOptions)

	if err != nil {
		logrus.WithError(err).Error("Error initializing Redis Manager")
	}

	sfcLoginClient := &login.SfcLoginClient{
		Proxy: proxy.NewProxy(config.SfcLoginUrl),
	}

	// Get token for salesforce
	tokenPayload := login.TokenPayload{
		ClientId:     config.SfcClientId,
		ClientSecret: config.SfcClientSecret,
		Username:     config.SfcUsername,
		Password:     config.SfcPassword + config.SfcSecurityToken,
	}

	token, err := sfcLoginClient.GetToken(tokenPayload)
	if err != nil {
		logrus.Errorf("Could not get access token from salesforce Server : %s", err.Error())
	}

	sfcChatClient := &chat.SfcChatClient{
		Proxy:       proxy.NewProxy(config.SfcChatUrl),
		ApiVersion:  config.SfcApiVersion,
		AccessToken: token,
	}

	salesforceClient := &salesforce.SalesforceClient{
		Proxy:       proxy.NewProxy(config.SfcBaseUrl),
		APIVersion:  config.SfcApiVersion,
		AccessToken: token,
	}

	m := &Manager{
		clientName:     config.AppName,
		sfcLoginClient: sfcLoginClient,
		sfcChatClient:  sfcChatClient,
		sfcClient:      salesforceClient,
	}
	return m
}
