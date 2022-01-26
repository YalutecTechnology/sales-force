package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
	"yalochat.com/salesforce-integration/base/helpers"
)

var timeWait float64 = 10

// WelcomeAPI get a welcome message in order to test ${appName} API
func (app *App) welcomeAPI(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	started := time.Now()
	helpers.WriteSuccessResponse(w, helpers.SuccessResponse{Message: "Starting API check..."})

	duration := time.Since(started)
	if duration.Seconds() > timeWait {
		logrus.Errorf("Error checking API health: %v", duration.Seconds())
		helpers.WriteFailedResponse(w, 500, fmt.Sprintf("Error checking API health: %v", duration.Seconds()))
	} else {
		logrus.Info("Welcome to API!")
		helpers.WriteSuccessResponse(w, helpers.SuccessResponse{Message: "Welcome to API!"})
	}
}
