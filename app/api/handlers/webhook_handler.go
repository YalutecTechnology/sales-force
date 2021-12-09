package handlers

import (
	"encoding/json"
	"fmt"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"net/http"
	"yalochat.com/salesforce-integration/base/constants"
	"yalochat.com/salesforce-integration/base/events"

	"github.com/julienschmidt/httprouter"
	"github.com/sirupsen/logrus"

	"yalochat.com/salesforce-integration/base/helpers"
	"yalochat.com/salesforce-integration/base/models"
)

const insertError = "There was an error inserting integration message"

// webhook to save messages from integrations API
func (app *App) webhook(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// datadog tracing
	span, _ := tracer.SpanFromContext(r.Context())
	span.SetOperationName("receive_message_wa")
	span.SetTag(ext.ResourceName, fmt.Sprintf("%s %s", r.Method, r.URL.RequestURI()))
	span.SetTag(ext.AnalyticsEvent, true)
	defer span.Finish()

	logFields := logrus.Fields{
		constants.TraceIdKey: span.Context().TraceID(),
		constants.SpanIdKey:  span.Context().SpanID(),
		events.Params:        params,
	}
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
		errorMessage := helpers.ErrorMessage(helpers.InvalidPayload, err)
		logrus.WithFields(logFields).WithError(err).Error(errorMessage)
		span.SetTag(ext.Error, err)
		span.SetTag(ext.ErrorDetails, errorMessage)
		span.SetTag(ext.HTTPCode, http.StatusBadRequest)
		helpers.WriteFailedResponse(w, http.StatusBadRequest, errorMessage)
		return
	}

	logFields[events.Payload] = integrationsRequest
	span.SetTag(events.Payload, fmt.Sprintf("%#v", integrationsRequest))
	if err := helpers.Govalidator().Struct(integrationsRequest); err != nil {
		errorMessage := helpers.ErrorMessage(helpers.ValidatePayloadError, err)
		logrus.WithFields(logFields).WithError(err).Error(errorMessage)
		span.SetTag(ext.Error, err)
		span.SetTag(ext.ErrorDetails, errorMessage)
		span.SetTag(ext.HTTPCode, http.StatusBadRequest)
		helpers.WriteFailedResponse(w, http.StatusBadRequest, errorMessage)
		return
	}

	err := app.ManageManager.SaveContext(r.Context(), &integrationsRequest)
	if err != nil {
		errorMessage := helpers.ErrorMessage(insertError, err)
		logrus.WithFields(logFields).WithError(err).Error(errorMessage)
		span.SetTag(ext.Error, err)
		span.SetTag(ext.ErrorDetails, errorMessage)
		span.SetTag(ext.HTTPCode, http.StatusInternalServerError)
		helpers.WriteFailedResponse(w, http.StatusInternalServerError, errorMessage)
		return
	}

	span.SetTag(ext.HTTPCode, http.StatusOK)
	helpers.WriteSuccessResponse(w, helpers.SuccessResponse{Message: "insert success"})
}

func (app *App) getContext(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	userID := params.ByName("user_id")

	if userID == "" {
		helpers.WriteFailedResponse(w, http.StatusBadRequest, helpers.MissingParam+" : user_id")
		return
	}

	ctx := app.ManageManager.GetContextByUserID(userID)

	helpers.WriteSuccessResponse(w, ctx)
}

// webhookFB to save messages from integrations API
func (app *App) webhookFB(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	span, _ := tracer.SpanFromContext(r.Context())
	span.SetOperationName("receive_message_fb")
	span.SetTag(ext.ResourceName, fmt.Sprintf("%s %s", r.Method, r.URL.RequestURI()))
	span.SetTag(ext.AnalyticsEvent, true)
	defer span.Finish()

	logFields := logrus.Fields{
		constants.TraceIdKey: span.Context().TraceID(),
		constants.SpanIdKey:  span.Context().SpanID(),
		events.Params:        params,
	}
	var integrationsRequest models.IntegrationsFacebook
	if err := json.NewDecoder(r.Body).Decode(&integrationsRequest); err != nil {
		errorMessage := helpers.ErrorMessage(helpers.InvalidPayload, err)
		logrus.WithFields(logFields).WithError(err).Error(errorMessage)
		span.SetTag(ext.Error, err)
		span.SetTag(ext.ErrorDetails, errorMessage)
		span.SetTag(ext.HTTPCode, http.StatusBadRequest)
		helpers.WriteFailedResponse(w, http.StatusBadRequest, errorMessage)
		return
	}

	logFields[events.Payload] = integrationsRequest
	span.SetTag(events.Payload, fmt.Sprintf("%#v", integrationsRequest))
	if err := helpers.Govalidator().Struct(integrationsRequest); err != nil {
		errorMessage := helpers.ErrorMessage(helpers.ValidatePayloadError, err)
		logrus.WithFields(logFields).WithError(err).Error(errorMessage)
		span.SetTag(ext.Error, err)
		span.SetTag(ext.ErrorDetails, errorMessage)
		span.SetTag(ext.HTTPCode, http.StatusBadRequest)
		helpers.WriteFailedResponse(w, http.StatusBadRequest, errorMessage)
		return
	}

	err := app.ManageManager.SaveContextFB(r.Context(), &integrationsRequest)
	if err != nil {
		errorMessage := helpers.ErrorMessage(insertError, err)
		logrus.WithFields(logFields).WithError(err).Error(errorMessage)
		span.SetTag(ext.Error, err)
		span.SetTag(ext.ErrorDetails, errorMessage)
		span.SetTag(ext.HTTPCode, http.StatusInternalServerError)
		helpers.WriteFailedResponse(w, http.StatusInternalServerError, errorMessage)
		return
	}

	span.SetTag(ext.HTTPCode, http.StatusOK)
	helpers.WriteSuccessResponse(w, helpers.SuccessResponse{Message: "insert success"})
}

// registerWebhook Register webhook to intagrations
func (app *App) registerWebhook(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	provider := params.ByName("provider")

	err := app.ManageManager.RegisterWebhookInIntegrations(provider)
	if err != nil {
		errorMessage := "error register webhook"
		logrus.WithFields(logrus.Fields{
			"provider": provider,
		}).WithError(err).Error(errorMessage)
		helpers.WriteFailedResponse(w, http.StatusInternalServerError, errorMessage)
		return
	}

	helpers.WriteSuccessResponse(w, helpers.SuccessResponse{Message: "Register webhook success with provider : " + provider})
}

// removeWebhook Remove webhook to intagrations
func (app *App) removeWebhook(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	provider := params.ByName("provider")

	err := app.ManageManager.RemoveWebhookInIntegrations(provider)
	if err != nil {
		errorMessage := "error remove webhook"
		logrus.WithFields(logrus.Fields{
			"provider": provider,
		}).WithError(err).Error(errorMessage)
		helpers.WriteFailedResponse(w, http.StatusInternalServerError, errorMessage)
		return
	}

	helpers.WriteSuccessResponse(w, helpers.SuccessResponse{Message: "Remove webhook success with provider : " + provider})
}
