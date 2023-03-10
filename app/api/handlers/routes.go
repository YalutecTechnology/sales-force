package handlers

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net/http/pprof"

	"github.com/sirupsen/logrus"
	ddrouter "gopkg.in/DataDog/dd-trace-go.v1/contrib/julienschmidt/httprouter"
	"yalochat.com/salesforce-integration/app/manage"
)

// App contains the resources related to the application
type App struct {
	ManageManager         manage.ManagerI
	Client                string
	YaloUsername          string
	YaloPassword          string
	SalesforceUsername    string
	SalesforcePassword    string
	SecretKey             string
	IntegrationsSignature string
	IgnoreMessageTypes    string
}

type ApiConfig struct {
	YaloUsername          string
	YaloPassword          string
	SalesforceUsername    string
	SalesforcePassword    string
	SecretKey             string
	IntegrationsSignature string
	IgnoreMessageTypes    string
	UseProfile            bool
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

	// Define Webhooks
	managerOptions.WebhookWhatsapp = fmt.Sprintf("%s/integrations/whatsapp/webhook", apiVersion)
	managerOptions.WebhookFacebook = fmt.Sprintf("%s/integrations/facebook/webhook", apiVersion)

	manager := manage.CreateManager(managerOptions)
	yaloHash := sha1.New()
	yaloHash.Write([]byte(apiConfig.YaloPassword))
	salesforceHash := sha1.New()
	salesforceHash.Write([]byte(apiConfig.SalesforcePassword))

	app := &App{
		ManageManager:         manager,
		Client:                managerOptions.Client,
		YaloUsername:          apiConfig.YaloUsername,
		YaloPassword:          hex.EncodeToString(yaloHash.Sum(nil)),
		SalesforceUsername:    apiConfig.SalesforceUsername,
		SalesforcePassword:    hex.EncodeToString(salesforceHash.Sum(nil)),
		SecretKey:             apiConfig.SecretKey,
		IntegrationsSignature: apiConfig.IntegrationsSignature,
		IgnoreMessageTypes:    apiConfig.IgnoreMessageTypes,
	}

	if len(apiConfig.IgnoreMessageTypes) > 0 {
		logrus.WithFields(logrus.Fields{
			"types": apiConfig.IgnoreMessageTypes,
		}).Info("App configured to ignore messages of some types")
	}

	setApp(*app)

	srv.GET(fmt.Sprintf("%s/welcome", apiVersion), app.welcomeAPI)
	srv.POST(fmt.Sprintf("%s/authenticate", apiVersion), app.authenticate)
	srv.GET(fmt.Sprintf("%s/tokens/check", apiVersion), app.authorizeMiddleware(app.getUserByToken, []RoleType{Yalo, Salesforce}))
	srv.POST(fmt.Sprintf("%s/chats/connect", apiVersion), app.authorizeMiddleware(app.createChat, []RoleType{Yalo}))
	srv.POST(managerOptions.WebhookWhatsapp, app.webhook)
	srv.GET(fmt.Sprintf("%s/context/:user_id", apiVersion), app.authorizeMiddleware(app.getContext, []RoleType{Yalo}))
	srv.POST(managerOptions.WebhookFacebook, app.webhookFB)
	srv.DELETE(fmt.Sprintf("%s/chat/finish/:user_id", apiVersion), app.authorizeMiddleware(app.finishChat, []RoleType{Yalo}))
	srv.POST(fmt.Sprintf("%s/integrations/webhook/register/:provider", apiVersion), app.authorizeMiddleware(app.registerWebhook, []RoleType{Yalo}))
	srv.DELETE(fmt.Sprintf("%s/integrations/webhook/remove/:provider", apiVersion), app.authorizeMiddleware(app.removeWebhook, []RoleType{Yalo}))

	if apiConfig.UseProfile {
		//endpoint to implement profiling in the app
		srv.HandlerFunc("GET", "/debug/pprof", pprof.Profile)
		srv.HandlerFunc("GET", "/debug/pprof/cmdline", pprof.Cmdline)
		srv.HandlerFunc("GET", "/debug/pprof/profile", pprof.Profile)
		srv.HandlerFunc("GET", "/debug/pprof/symbol", pprof.Symbol)
		srv.HandlerFunc("GET", "/debug/pprof/trace", pprof.Trace)
		srv.Handler("GET", "/debug/pprof/block", pprof.Handler("block"))
		srv.Handler("GET", "/debug/pprof/goroutine", pprof.Handler("goroutine"))
		srv.Handler("GET", "/debug/pprof/heap", pprof.Handler("heap"))
		srv.Handler("GET", "/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	}

}
