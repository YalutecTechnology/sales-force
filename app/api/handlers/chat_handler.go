package handlers

import (
	"encoding/json"
	"net/http"

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
	var chatPayload = &ChatPayload{}
	//unmarshalling payload
	if err := json.NewDecoder(r.Body).Decode(&chatPayload); err != nil {
		helpers.WriteFailedResponse(w, http.StatusBadRequest, helpers.ErrorMessage(helpers.InvalidPayload, err))
		return
	}

	//validating payload struct
	if err := helpers.Govalidator().Struct(chatPayload); err != nil {
		helpers.WriteFailedResponse(w, http.StatusBadRequest, helpers.ErrorMessage(helpers.ValidatePayloadError, err))
		return
	}

	// Create Interconnection between yalo and salesforce
	interconnection := &manage.Interconnection{
		UserID:      chatPayload.UserID,
		Name:        chatPayload.Name,
		Provider:    manage.Provider(chatPayload.Provider),
		BotSlug:     chatPayload.BotSlug,
		BotId:       chatPayload.BotId,
		PhoneNumber: chatPayload.PhoneNumber,
		Email:       chatPayload.Email,
		ExtraData:   chatPayload.ExtraData,
	}
	if err := app.ManageManager.CreateChat(interconnection); err != nil {
		helpers.WriteFailedResponse(w, http.StatusNotFound, err.Error())
		return
	}

	helpers.WriteSuccessResponse(w, helpers.SuccessResponse{Message: "Chat created succefully"})
}
