package cron

import (
	"net/http"
	"time"

	cronV3 "github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"yalochat.com/salesforce-integration/app/services"
	"yalochat.com/salesforce-integration/base/cache"
)

type cron struct {
	salesforceService services.SalesforceServiceInterface
	SpecSchedule      string
	Contextschedule   string
	ContactEmail      string
	Client            string
	ContextCache      cache.IContextCache
}

func NewCron(
	salesforceService services.SalesforceServiceInterface,
	specSchedule,
	contactEmail string,
) *cron {
	return &cron{
		salesforceService: salesforceService,
		SpecSchedule:      specSchedule,
		ContactEmail:      contactEmail,
	}
}

func (c *cron) Run() {
	err := c.setCron()
	if err != nil {
		panic(err)
	}
}

func (c *cron) setCron() error {

	crons := cronV3.New()

	_, err := crons.AddFunc(c.SpecSchedule, func() {
		_, err := c.salesforceService.SearchContactComposite(c.ContactEmail, "", nil, nil)
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
