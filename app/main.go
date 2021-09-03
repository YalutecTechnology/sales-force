package main

import (
	"fmt"
	"net/http"
	"strings"
	"time"

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

func init() {
	logrus.SetFormatter(&logrus.JSONFormatter{
		FieldMap: logrus.FieldMap{
			logrus.FieldKeyLevel: "severity",
		},
	})
	logrus.SetReportCaller(false)
}

func main() {
	logrus.Info("Initializing Integrations API application")
	var envs envs.Envs
	err := envconfig.Process("salesforce-integration", &envs)
	if err != nil {
		logrus.WithError(err).Fatal("Error with the environment configuration")
	}
	sentry.Init(sentry.ClientOptions{
		Dsn:         envs.SentryDSN,
		Environment: envs.Environment,
	})

	managerOptions := &manage.ManagerOptions{
		AppName:           envs.AppName,
		SfcClientId:       envs.SfcClientId,
		SfcClientSecret:   envs.SfcClientSecret,
		SfcUsername:       envs.SfcUsername,
		SfcPassword:       envs.SfcPassword,
		SfcSecurityToken:  envs.SfcSecurityToken,
		SfcBaseUrl:        envs.SfcBaseUrl,
		SfcLoginUrl:       envs.SfcLoginUrl,
		SfcChatUrl:        envs.SfcChatUrl,
		SfcApiVersion:     envs.SfcApiVersion,
		SfcOrganizationId: envs.SfcOrganizationId,
		SfcDeploymentId:   envs.SfcDeploymentId,
		SfcButtonId:       envs.SfcButtonId,
		SfcOwnerId:        envs.SfcOwnerId,
	}

	if len(envs.RedisMaster) > 0 {
		managerOptions.RedisOptions.FailOverOptions = &redis.FailoverOptions{
			MasterName:    envs.RedisMaster,
			SentinelAddrs: strings.Split(envs.RedisSentinelAddress, ";"),
			IdleTimeout:   time.Second * 60,
			PoolSize:      1000,
			MinIdleConns:  10,
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
