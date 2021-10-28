package cron

import (
	"net/http"

	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"yalochat.com/salesforce-integration/app/services"
)

type crons struct {
	salesforceService services.SalesforceServiceInterface
	SpecSchedule      string
	ContactEmail      string
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

	crons.Start()

	return nil
}
