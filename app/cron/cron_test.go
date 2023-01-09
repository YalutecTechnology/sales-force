package cron

import (
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/stretchr/testify/assert"
	"yalochat.com/salesforce-integration/app/cron/mocks"
	"yalochat.com/salesforce-integration/base/helpers"
	"yalochat.com/salesforce-integration/base/models"
)

const (
	waitTime      = 1*time.Second + 20*time.Millisecond
	email         = "email@example.com"
	client        = "client"
	expresionCron = "0 9 * * *"
)

func Test_crons_Run(t *testing.T) {
	t.Run("Should set a cron instance", func(t *testing.T) {
		cronService := NewCron(nil, "@every 1h30m", "test")
		cronService.Run()
	})
}

func Test_crons_setCron(t *testing.T) {
	t.Run("Should set a cron error", func(t *testing.T) {
		cronService := NewCron(nil, "", "test")
		err := cronService.setCron()
		assert.Error(t, err)
	})

	t.Run("Should exec cron", func(t *testing.T) {
		wg := &sync.WaitGroup{}
		wg.Add(1)

		saleforceService := new(mocks.SalesforceServiceInterface)
		contact := &models.SfcContact{
			Email:   email,
			Blocked: true,
		}
		saleforceService.On("SearchContactComposite", email, "", map[string]string(nil), map[string]interface{}(nil)).Return(contact, nil).Times(10)

		contextCacheMock := new(mocks.IContextCache)
		contextCacheMock.On("CleanContextToDate", client, mock.Anything).Return(nil).Times(10)

		cronService := NewCron(saleforceService, "@every 1s", email)
		cronService.Client = client
		cronService.Contextschedule = "@every 1s"
		cronService.ContextCache = contextCacheMock
		err := cronService.setCron()

		time.Sleep(waitTime)
		assert.NoError(t, err)

	})

	t.Run("Should exec cron error services", func(t *testing.T) {
		wg := &sync.WaitGroup{}
		wg.Add(1)

		saleforceService := new(mocks.SalesforceServiceInterface)
		contact := &models.SfcContact{}
		saleforceService.On("SearchContactComposite", email, "", map[string]string(nil), map[string]interface{}(nil)).Return(contact, &helpers.ErrorResponse{
			StatusCode: http.StatusUnauthorized,
		}).Times(10)

		saleforceService.On("RefreshToken").Times(10)

		contextCacheMock := new(mocks.IContextCache)
		contextCacheMock.On("CleanContextToDate", client, mock.Anything).Return(assert.AnError).Times(10)

		cronService := NewCron(saleforceService, "@every 1s", email)
		cronService.Client = client
		cronService.Contextschedule = "@every 1s"
		cronService.ContextCache = contextCacheMock
		err := cronService.setCron()

		time.Sleep(waitTime)
		assert.NoError(t, err)

	})

}
