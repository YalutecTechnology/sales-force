package models

import (
	"encoding/json"
	"fmt"
)

type MessageTemplate struct {
	WaitAgent          string `json:"waitAgent"`
	QueuePosition      string `json:"queuePosition"`
	WaitTime           string `json:"waitTime"`
	WelcomeTemplate    string `json:"welcomeTemplate"`
	Context            string `json:"context"`
	DescriptionCase    string `json:"descriptionCase"`
	UploadImageError   string `json:"uploadImageError"`
	UploadImageSuccess string `json:"uploadImageSuccess"`
	UploadFileError    string `json:"uploadFileError"`
	UploadFileSuccess  string `json:"uploadFileSuccess"`
	UploadAudioError   string `json:"uploadAudioError"`
	UploadAudioSuccess string `json:"uploadAudioSuccess"`
	FirstNameContact   string `json:"firstNameContact"`
	ClientLabel        string `json:"clientLabel"`
	BotLabel           string `json:"botLabel"`
}

// Decode Decoder this function deserializes the struct by the envconfig Decoder interface implementation
func (sd *MessageTemplate) Decode(value string) error {
	var messageTemplate = &MessageTemplate{}

	err := json.Unmarshal([]byte(value), &messageTemplate)
	if err != nil {
		return fmt.Errorf("invalid map json: %w", err)
	}
	*sd = *messageTemplate

	return nil
}
