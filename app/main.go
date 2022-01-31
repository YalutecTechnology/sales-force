package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"

	"github.com/getsentry/sentry-go"
	"github.com/go-redis/redis"
	"github.com/kelseyhightower/envconfig"
	"github.com/sirupsen/logrus"
	ddrouter "gopkg.in/DataDog/dd-trace-go.v1/contrib/julienschmidt/httprouter"
	"yalochat.com/salesforce-integration/app/api/handlers"
	"yalochat.com/salesforce-integration/app/config/envs"
	"yalochat.com/salesforce-integration/app/manage"
)

var httpServer http.Server

// defaultServiceName hols the default service name used when DD env variable is not set
const defaultServiceName = "salesforce-integration"

// ddServiceEnvVar is the env Var used by DD library to load service name
const ddServiceEnvVar = "DD_SERVICE"

// secondsToShutdown are the seconds to shutdown de app
const secondsToShutdown = 3

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyLevel: "severity",
		},
	})
	logrus.SetReportCaller(false)
}

func main() {
	go shutdownHandler()
	logrus.Info("Initializing Integrations API application")
	var envs envs.Envs
	err := envconfig.Process("salesforce-integration", &envs)
	if err != nil {
		logrus.WithError(err).Fatal("Error with the environment configuration")
	}

	// datadog tracer start
	if len(os.Getenv(ddServiceEnvVar)) == 0 {
		logrus.Info("DD_SERVICE envar not exists")
		tracer.Start(tracer.WithService(defaultServiceName), tracer.WithAnalytics(true))
	} else {
		logrus.Info("DD_SERVICE envar exists")
		tracer.Start(tracer.WithAnalytics(true))
	}
	defer func() {
		tracer.Stop()
		logrus.Info("Tracing stopped correctly")
	}()

	sentry.Init(sentry.ClientOptions{
		Dsn:         envs.SentryDSN,
		Environment: envs.Environment,
	})

	managerOptions := &manage.ManagerOptions{
		AppName:                    envs.AppName,
		Client:                     envs.Client,
		BlockedUserState:           envs.BlockedUserState,
		BotrunnerUrl:               envs.BotrunnerUrl,
		BotrunnerToken:             envs.BotrunnerToken,
		BotrunnerTimeout:           envs.BotrunnerTimeout,
		TimeoutState:               envs.TimeoutState,
		SuccessState:               envs.SuccessState,
		SfcClientID:                envs.SfcClientID,
		SfcClientSecret:            envs.SfcClientSecret,
		SfcUsername:                envs.SfcUsername,
		SfcPassword:                envs.SfcPassword,
		SfcSecurityToken:           envs.SfcSecurityToken,
		SfcBaseUrl:                 envs.SfcBaseUrl,
		SfcLoginUrl:                envs.SfcLoginUrl,
		SfcChatUrl:                 envs.SfcChatUrl,
		SfcApiVersion:              envs.SfcApiVersion,
		SfcOrganizationID:          envs.SfcOrganizationId,
		SfcDeploymentID:            envs.SfcDeploymentId,
		SfcRecordTypeID:            envs.SfcRecordTypeId,
		SfcAccountRecordTypeID:     envs.SfcAccountRecordTypeId,
		SfcDefaultBirthDateAccount: envs.SfcDefaultBirthDateAccount,
		SfcCustomFieldsCase:        envs.SfcCustomFieldsCase,
		SfcCustomFieldsContact:     envs.SfcCustomFieldsContact,
		SfcSourceFlowField:         envs.SfcSourceFlowField,
		SfcSourceFlowBot:           envs.SfcSourceFlowBot,
		SfcCodePhoneRemove:         envs.SfcCodePhoneRemove,
		IntegrationsUrl:            envs.IntegrationsBaseUrl,
		IntegrationsWAChannel:      envs.IntegrationsWAChannel,
		IntegrationsFBChannel:      envs.IntegrationsFBChannel,
		IntegrationsWABotID:        envs.IntegrationsWABotID,
		IntegrationsFBBotID:        envs.IntegrationsFBBotID,
		IntegrationsFBToken:        envs.IntegrationsFBBotJWT,
		IntegrationsWAToken:        envs.IntegrationsWABotJWT,
		IntegrationsSignature:      envs.IntegrationsSignature,
		IntegrationsWABotPhone:     envs.IntegrationsWABotPhone,
		IntegrationsFBBotPhone:     envs.IntegrationsFBBotPhone,
		WebhookBaseUrl:             envs.WebhookBaseUrl,
		Environment:                envs.Environment,
		KeywordsRestart:            envs.KeywordsRestart,
		SfcBlockedChatField:        envs.SfcBlockedChatField,
		StudioNGUrl:                envs.StudioNGUrl,
		StudioNGToken:              envs.StudioNGToken,
		StudioNGTimeout:            envs.StudioNGTimeout,
		SpecSchedule:               envs.SpecSchedule,
		MaxRetries:                 envs.MaxRetries,
		CleanContextSchedule:       envs.CleanContextSchedule,
		IntegrationsRateLimit:      envs.IntegrationChanRateLimit,
		SalesforceRateLimit:        envs.SaleforceChanRateLimit,
		Messages:                   envs.Messages,
		Timezone:                   envs.Timezone,
		SendImageNameInMessage:     envs.SendImageNameInMessage,
		KafkaHost:                  envs.KafkaHost,
		KafkaPort:                  envs.KafkaPort,
		KafkaUser:                  envs.KafkaUser,
		KafkaPassword:              envs.KafkaPassword,
		KafkaTopic:                 envs.KafkaTopic,
	}

	if len(envs.RedisMaster) > 0 {
		managerOptions.RedisOptions.FailOverOptions = &redis.FailoverOptions{
			MasterName:    envs.RedisMaster,
			SentinelAddrs: strings.Split(envs.RedisSentinelAddress, ";"),
			IdleTimeout:   time.Second * 60,
			PoolSize:      1000,
			MinIdleConns:  10,
			ReadTimeout:   time.Second * 15,
		}
	}
	if len(envs.RedisAddress) > 0 {
		managerOptions.RedisOptions.Options = &redis.Options{
			Addr: envs.RedisAddress,
		}
	}

	srv := ddrouter.New(ddrouter.WithServiceName("salesforce-integration.http"))
	handlers.API(srv, managerOptions, handlers.ApiConfig{
		YaloUsername:          envs.YaloUsername,
		YaloPassword:          envs.YaloPassword,
		SalesforceUsername:    envs.SalesforceUsername,
		SalesforcePassword:    envs.SalesforcePassword,
		SecretKey:             envs.SecretKey,
		IntegrationsSignature: envs.IntegrationsSignature,
	})

	httpServer = http.Server{
		Addr:    fmt.Sprintf("%s:%s", envs.Host, envs.Port),
		Handler: srv,
	}

	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logrus.WithError(err).Fatal("Error salesforce-integration API HTTP server")
	}
}

// shutdownHandler triggers application shutdown.

func shutdownHandler() {
	// signChan channel is used to transmit signal notifications.
	signChan := make(chan os.Signal, 1)

	// Catch and relay certain signal(s) to signChan channel.
	signal.Notify(signChan, os.Interrupt, syscall.SIGTERM)

	// Blocking until a signal is sent over signChan channel.
	sig := <-signChan

	logrus.Infof("Cleanup started with %s signal", sig)
	time.Sleep(time.Duration(secondsToShutdown) * time.Second)
	logrus.Infof("Shoutdown and Cleanup completed in %d seconds", secondsToShutdown)
	os.Exit(1)
}
