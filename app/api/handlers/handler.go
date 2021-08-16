package handlers

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"
	"yalochat.com/salesforce-integration/base/helpers"
)

// WelcomeAPI get a welcome message in order to test ${appName} API
func (app *App) welcomeAPI(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	logrus.Info(w, " Welcome to API!")
	helpers.WriteSuccessResponse(w, helpers.SuccessResponse{Message: "Welcome to API!"})
}
