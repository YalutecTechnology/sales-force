package cron

import (
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	helpers "yalochat.com/salesforce-integration/base/helpers"
	models "yalochat.com/salesforce-integration/base/models"
)

const (
	oneSecond = 1*time.Second + 50*time.Millisecond
	email     = "email@example.com"
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

		saleforceService := new(SalesforceServiceInterface)
		contact := &models.SfcContact{
			Email:   email,
			Blocked: true,
		}
		saleforceService.On("SearchContactComposite", email, "").Return(contact, nil).Once()

		cronService := NewCron(saleforceService, "@every 1s", email)
		err := cronService.setCron()

		assert.NoError(t, err)

	})

	t.Run("Should exec cron error services", func(t *testing.T) {
		wg := &sync.WaitGroup{}
		wg.Add(1)

		saleforceService := new(SalesforceServiceInterface)
		contact := &models.SfcContact{}
		saleforceService.On("SearchContactComposite", email, "").Return(contact, &helpers.ErrorResponse{
			StatusCode: http.StatusUnauthorized,
		}).Once()

		saleforceService.On("RefreshToken").Once().After(oneSecond)

		cronService := NewCron(saleforceService, "@every 1s", email)
		err := cronService.setCron()

		assert.NoError(t, err)

	})

	<-time.After(oneSecond)

}
