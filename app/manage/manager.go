package manage

import (
	"github.com/sirupsen/logrus"
	"yalochat.com/salesforce-integration/base/cache"
	"yalochat.com/salesforce-integration/base/clients/proxy"
	"yalochat.com/salesforce-integration/base/clients/salesforce"
)

// Manager controls the process of the app
type Manager struct {
	clientName       string
	salesforceClient *salesforce.SalesforceClient
}

// ManagerOptions holds configurations for the agents manager
type ManagerOptions struct {
	AppName               string
	RedisOptions          cache.RedisOptions
	ClentId               string
	ClientSecret          string
	SalesforceApiUsername string
	SalesforceApiPassword string
	SalesforceUrl         string
	SalesforceLoginUrl    string
	SalesforceCaseUrl     string
	SalesforceApiVersion  int
}

// CreateManager retrieves an agents manager
func CreateManager(config *ManagerOptions) *Manager {
	_, err := cache.NewRedisCache(&config.RedisOptions)

	if err != nil {
		logrus.WithError(err).Error("Error initializing Redis Manager")
	}

	salesforceClient := &salesforce.SalesforceClient{
		Proxy:         &proxy.Proxy{},
		LoginURL:      config.SalesforceLoginUrl,
		CaseURL:       config.SalesforceCaseUrl,
		SalesforceURL: config.SalesforceUrl,
		ApiVersion:    config.SalesforceApiVersion,
	}

	// Get token for salesforce
	payload := salesforce.TokenPayload{
		ClientId:     config.ClentId,
		ClientSecret: config.ClientSecret,
		Username:     config.SalesforceApiUsername,
		Password:     config.SalesforceApiPassword,
	}
	if _, err := salesforceClient.GetToken(payload); err != nil {
		logrus.Errorf("Could not get access token from salesforce Server : %s", err.Error())
	}

	session, err := salesforceClient.CreateSession()
	if err != nil {
		logrus.Errorf("Could not create session from salesforce Server : %s", err.Error())
	} else {
		logrus.Infof("Session Created: %v", session)
	}

	m := &Manager{
		clientName:       config.AppName,
		salesforceClient: salesforceClient,
	}
	return m
}
