package handlers

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"

	"github.com/sirupsen/logrus"
	ddrouter "gopkg.in/DataDog/dd-trace-go.v1/contrib/julienschmidt/httprouter"
	"yalochat.com/salesforce-integration/app/manage"
)

// App contains the resources related to the application
type App struct {
	ManageManager         *manage.Manager
	YaloUsername          string
	YaloPassword          string
	SalesforceUsername    string
	SalesforcePassword    string
	SecretKey             string
	IntegrationsSignature string
}

type ApiConfig struct {
	YaloUsername          string
	YaloPassword          string
	SalesforceUsername    string
	SalesforcePassword    string
	SecretKey             string
	IntegrationsSignature string
}

const apiVersion = "/v1"

//These methods are only helpers to be able to access the app in the tests
var app App

func getApp() *App {
	return &app
}
func setApp(newApp App) {
	app = newApp
}

// API Rest interface for app
func API(srv *ddrouter.Router, managerOptions *manage.ManagerOptions, apiConfig ApiConfig) {
	logrus.Info("API initilizated")
	manager := manage.CreateManager(managerOptions)
	yaloHash := sha1.New()
	yaloHash.Write([]byte(apiConfig.YaloPassword))
	salesforceHash := sha1.New()
	salesforceHash.Write([]byte(apiConfig.SalesforcePassword))

	app := &App{
		ManageManager:         manager,
		YaloUsername:          apiConfig.YaloUsername,
		YaloPassword:          hex.EncodeToString(yaloHash.Sum(nil)),
		SalesforceUsername:    apiConfig.SalesforceUsername,
		SalesforcePassword:    hex.EncodeToString(salesforceHash.Sum(nil)),
		SecretKey:             apiConfig.SecretKey,
		IntegrationsSignature: apiConfig.IntegrationsSignature,
	}
	setApp(*app)

	srv.GET(fmt.Sprintf("%s/welcome", apiVersion), app.welcomeAPI)
	srv.POST(fmt.Sprintf("%s/authenticate", apiVersion), app.authenticate)
	srv.GET(fmt.Sprintf("%s/tokens/check", apiVersion), app.authorizeMiddleware(app.getUserByToken, []RoleType{Yalo, Salesforce}))
	srv.POST(fmt.Sprintf("%s/chats/connect", apiVersion), app.authorizeMiddleware(app.createChat, []RoleType{Yalo}))
	srv.POST(fmt.Sprintf("%s/integrations/webhook", apiVersion), app.webhook)
}
