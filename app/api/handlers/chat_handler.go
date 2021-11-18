package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"net/http"
	"yalochat.com/salesforce-integration/base/constants"
	"yalochat.com/salesforce-integration/base/events"

	"github.com/julienschmidt/httprouter"
	"yalochat.com/salesforce-integration/app/manage"
	"yalochat.com/salesforce-integration/base/helpers"
)

type ChatPayload struct {
	UserID      string                 `json:"userID" validate:"required"`
	Name        string                 `json:"name" validate:"required"`
	Provider    string                 `json:"provider" validate:"required"`
	BotSlug     string                 `json:"botSlug" validate:"required"`
	BotId       string                 `json:"botId" validate:"required"`
	Email       string                 `json:"email" validate:"required"`
	PhoneNumber string                 `json:"phoneNumber"`
	ExtraData   map[string]interface{} `json:"extraData"`
}

// Connect and create chat between user and salesforce
func (app *App) createChat(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	// datadog tracing
	span, _ := tracer.StartSpanFromContext(r.Context(), "chats.connect")
	span.SetTag(ext.ResourceName, fmt.Sprintf("%s %s", r.Method, r.URL.RequestURI()))
	span.SetTag(ext.AnalyticsEvent, true)
	defer span.Finish()

	logFields := logrus.Fields{
		constants.TraceIdKey: span.Context().TraceID(),
		constants.SpanIdKey:  span.Context().SpanID(),
		events.Params:        params,
	}

	logrus.WithFields(logFields).Info("Create new chat")
	var chatPayload = &ChatPayload{}
	var errorMessage string
	//unmarshalling payload
	if err := json.NewDecoder(r.Body).Decode(&chatPayload); err != nil {
		errorMessage = helpers.ErrorMessage(helpers.InvalidPayload, err)
		span.SetTag(ext.Error, err)
		span.SetTag(ext.ErrorDetails, errorMessage)
		logrus.Error(errorMessage)
		helpers.WriteFailedResponse(w, http.StatusBadRequest, errorMessage)
		return
	}

	logFields[events.Payload] = chatPayload
	span.SetTag(events.Payload, fmt.Sprintf("%#v", chatPayload))
	//validating payload struct
	if err := helpers.Govalidator().Struct(chatPayload); err != nil {
		errorMessage = helpers.ErrorMessage(helpers.ValidatePayloadError, err)
		span.SetTag(ext.Error, err)
		span.SetTag(ext.ErrorDetails, errorMessage)
		logrus.WithFields(logFields).Error(errorMessage)
		helpers.WriteFailedResponse(w, http.StatusBadRequest, errorMessage)
		return
	}

	// Create Interconnection between yalo and salesforce
	interconnection := &manage.Interconnection{
		UserID:      chatPayload.UserID,
		Name:        chatPayload.Name,
		Provider:    manage.Provider(chatPayload.Provider),
		BotSlug:     chatPayload.BotSlug,
		BotID:       chatPayload.BotId,
		PhoneNumber: chatPayload.PhoneNumber,
		Email:       chatPayload.Email,
		ExtraData:   chatPayload.ExtraData,
	}

	logFields[events.Interconnection] = interconnection
	span.SetTag(events.Interconnection, fmt.Sprintf("%#v", interconnection))
	if err := app.ManageManager.CreateChat(r.Context(), interconnection); err != nil {
		errorMessage = err.Error()
		span.SetTag(ext.Error, err)
		span.SetTag(ext.ErrorDetails, errorMessage)
		logrus.WithFields(logFields).Error(errorMessage)
		helpers.WriteFailedResponse(w, http.StatusNotFound, errorMessage)
		return
	}

	helpers.WriteSuccessResponse(w, helpers.SuccessResponse{Message: "Chat created succefully"})
}

// Connect and end the chat between the user and the sales force
func (app *App) finishChat(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	userID := params.ByName("user_id")

	// Finished chat
	if err := app.ManageManager.FinishChat(userID); err != nil {
		helpers.WriteFailedResponse(w, http.StatusNotFound, err.Error())
		return
	}

	helpers.WriteSuccessResponse(w, helpers.SuccessResponse{Message: "Chat finished successfully"})
}
