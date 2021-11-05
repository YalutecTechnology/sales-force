package models

type IntegrationsRequest struct {
	ID        string `json:"id" validate:"required"`
	Timestamp string `json:"timestamp" validate:"required"`
	Type      string `json:"type" validate:"required"`
	From      string `json:"from"`
	To        string `json:"to"`
	Audio     Media  `json:"audio,omitempty"`
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

type IntegrationsFacebook struct {
	AuthorRole  string      `json:"authorRole" validate:"required"`
	BotID       string      `json:"botId" validate:"required"`
	Message     Message     `json:"message" validate:"required"`
	MsgTracking MsgTracking `json:"msgTracking"`
	Provider    string      `json:"provider"`
	Timestamp   int64       `json:"timestamp"`
}

type Message struct {
	Entry  []Entry `json:"entry"`
	Object string  `json:"object"`
}

type Entry struct {
	ID        string      `json:"id"`
	Messaging []Messaging `json:"messaging"`
	Time      int64       `json:"time"`
}

type Messaging struct {
	Message   MessagingMessage `json:"message"`
	Recipient Recipient        `json:"recipient"`
	Sender    Recipient        `json:"sender"`
	Timestamp int64            `json:"timestamp"`
}

type MessagingMessage struct {
	Text        string       `json:"text"`
	Attachments []Attachment `json:"attachments"`
	Mid         string       `json:"mid"`
}

type Attachment struct {
	Payload Payload `json:"payload"`
	Type    string  `json:"type"`
}

type Payload struct {
	URL string `json:"url"`
}

type Recipient struct {
	ID string `json:"id"`
}

type MsgTracking struct {
}
