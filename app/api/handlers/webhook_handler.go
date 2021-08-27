package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"yalochat.com/salesforce-integration/base/helpers"
	"yalochat.com/salesforce-integration/base/models"
)

const insertError = "There was an error inserting integration message: %s"

// webhook to save messages from integrations API
func (app *App) webhook(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	/*signature := r.Header.Get("x-yalochat-signature")
	if signature == "" {
		helpers.WriteFailedResponse(w, http.StatusUnauthorized, "x-yalochat-signature required, header invalid.")
		return
	}

	if signature != app.IntegrationsSignature {
		helpers.WriteFailedResponse(w, http.StatusUnauthorized, "x-yalochat-signature invalid, header invalid.")
		return
	}*/

	var integrationsRequest models.IntegrationsRequest
	if err := json.NewDecoder(r.Body).Decode(&integrationsRequest); err != nil {
		helpers.WriteFailedResponse(w, http.StatusBadRequest, helpers.InvalidPayload+" : "+err.Error())
		return
	}

	if err := helpers.Govalidator().Struct(integrationsRequest); err != nil {
		helpers.WriteFailedResponse(w, http.StatusBadRequest, helpers.ValidatePayloadError+" : "+err.Error())
		return
	}

	err := app.ManageManager.SaveContext(&integrationsRequest)
	if err != nil {
		helpers.WriteFailedResponse(w, http.StatusNotFound, fmt.Sprintf(insertError, err))
		return
	}

	helpers.WriteSuccessResponse(w, helpers.SuccessResponse{Message: "insert success"})
}
