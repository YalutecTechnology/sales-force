package models

type IntegrationsRequest struct {
	ID        string `json:"id" validate:"required"`
	Timestamp string `json:"timestamp" validate:"required"`
	Type      string `json:"type" validate:"required"`
	From      string `json:"from"`
	To        string `json:"to"`
	Voice     Media  `json:"voice,omitempty"`
	Document  Media  `json:"document,omitempty"`
	Image     Media  `json:"image,omitempty"`
	Text      Text   `json:"text,omitempty"`
}

type Media struct {
	URL      string `json:"url,omitempty"`
	MIMEType string `json:"mimeType,omitempty"`
	Caption  string `json:"caption,omitempty"`
}

type Text struct {
	Body string `json:"body"`
}
