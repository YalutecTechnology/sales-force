package cron

import (
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"net/http"
	"time"
	"yalochat.com/salesforce-integration/app/services"
	"yalochat.com/salesforce-integration/base/cache"
)

type crons struct {
	salesforceService services.SalesforceServiceInterface
	SpecSchedule      string
	Contextschedule   string
	ContactEmail      string
	Client            string
	ContextCache      cache.ContextCache
}

func NewCron(salesforceService services.SalesforceServiceInterface,
	specSchedule, contactEmail string) *crons {
	return &crons{
		salesforceService: salesforceService,
		SpecSchedule:      specSchedule,
		ContactEmail:      contactEmail,
	}
}

func (c *crons) Run() {
	err := c.setCron()
	if err != nil {
		panic(err)
	}
}

func (c *crons) setCron() error {

	crons := cron.New()

	_, err := crons.AddFunc(c.SpecSchedule, func() {
		_, err := c.salesforceService.SearchContactComposite(c.ContactEmail, "")
		if err != nil {
			if err.StatusCode == http.StatusUnauthorized {
				c.salesforceService.RefreshToken()
			}
		}

	})

	if err != nil {
		logrus.Errorf("Error to set cron %s", err.Error())
		return err
	}

	if c.Contextschedule != "" {
		_, err = crons.AddFunc(c.Contextschedule, func() {
			logrus.WithField("Client", c.Client).Info("Clean context")
			err := c.ContextCache.CleanContextToDate(c.Client, time.Now())
			if err != nil {
				logrus.Errorf("Could not clean context %s", err.Error())
			}

		})

		if err != nil {
			logrus.Errorf("Error to context set cron %s", err.Error())
			return err
		}
	}
	crons.Start()

	return nil
}
