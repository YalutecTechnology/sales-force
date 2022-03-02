package envs

import (
	"encoding/json"
	"fmt"
	"strings"

	"yalochat.com/salesforce-integration/base/models"
)

// Envs represents the list of well known env vars used by the app
type Envs struct {
	AppName              string            `default:"salesforce-integration" split_words:"true"`
	Client               string            `default:"salesforce" split_words:"true"`
	Host                 string            `required:"true" split_words:"true" default:"localhost"`
	Port                 string            `required:"true" split_words:"true" default:"8080"`
	SentryDSN            string            `default:"" split_words:"true"`
	Environment          string            `default:"dev" split_words:"true"`
	RedisAddress         string            `split_words:"true"`
	RedisMaster          string            `split_words:"true"`
	RedisSentinelAddress string            `split_words:"true"`
	BlockedUserState     map[string]string `required:"true" split_words:"true" default:"whatsapp:from-sf-blocked,facebook:from-sf-blocked"`
	TimeoutState         map[string]string `required:"true" split_words:"true" default:"whatsapp:from-sf-timeout,facebook:from-sf-timeout"`
	SuccessState         map[string]string `required:"true" split_words:"true" default:"whatsapp:from-sf-success,facebook:from-sf-success"`
	YaloUsername         string            `required:"true" split_words:"true" default:"yaloUser"`
	YaloPassword         string            `required:"true" split_words:"true"`
	SalesforceUsername   string            `required:"true" split_words:"true" default:"salesforceUser"`
	SalesforcePassword   string            `required:"true" split_words:"true"`
	SecretKey            string            `required:"true" split_words:"true"`
	BotrunnerUrl         string            `split_words:"true"`
	BotrunnerToken       string            `split_words:"true" default:""`
	BotrunnerTimeout     int               `split_words:"true" default:"4"`
	StudioNGUrl          string            `split_words:"true"`
	StudioNGToken        string            `split_words:"true"`
	StudioNGTimeout      int               `split_words:"true" default:"4"`
	SfcClientID          string            `split_words:"true"`
	SfcClientSecret      string            `split_words:"true"`
	SfcUsername          string            `split_words:"true"`
	SfcPassword          string            `split_words:"true"`
	SfcSecurityToken     string            `split_words:"true"`
	SfcBaseUrl           string            `split_words:"true"`
	SfcChatUrl           string            `split_words:"true"`
	SfcLoginUrl          string            `split_words:"true"`
	SfcApiVersion        string            `split_words:"true" default:"52"`
	SfcOrganizationId    string            `split_words:"true"`
	SfcDeploymentId      string            `split_words:"true"`
	SfcRecordTypeId      string            `split_words:"true"`
	// Only if this value exists will person accounts be created instead of contacts in salesforce
	SfcAccountRecordTypeId     string                 `split_words:"true"`
	SfcDefaultBirthDateAccount string                 `split_words:"true" default:"1921-01-01T00:00:00"`
	SfcCustomFieldsCase        map[string]string      `split_words:"true"`
	SfcCustomFieldsContact     map[string]string      `split_words:"true"`
	SfcSourceFlowBot           SfcSourceFlowBot       `required:"true" split_words:"true"`
	SfcSourceFlowField         string                 `required:"true" split_words:"true" default:"source_flow_bot"`
	SfcBlockedChatField        bool                   `split_words:"true" default:"false"`
	SfcCodePhoneRemove         []string               `split_words:"true" default:"521,52"`
	IntegrationsWAChannel      string                 `split_words:"true" default:"outgoing_webhook"`
	IntegrationsFBChannel      string                 `split_words:"true" default:"passthrough"`
	IntegrationsWABotID        string                 `split_words:"true"`
	IntegrationsFBBotID        string                 `split_words:"true"`
	IntegrationsWABotJWT       string                 `split_words:"true"`
	IntegrationsFBBotJWT       string                 `split_words:"true"`
	IntegrationsBaseUrl        string                 `split_words:"true"`
	IntegrationsSignature      string                 `split_words:"true"`
	WebhookBaseUrl             string                 `split_words:"true"`
	IntegrationsWABotPhone     string                 `split_words:"true"`
	IntegrationsFBBotPhone     string                 `split_words:"true"`
	KeywordsRestart            []string               `split_words:"true" default:"coppelbot,regresar,reiniciar,restart"`
	SpecSchedule               string                 `split_words:"true" default:"@every 59m"`
	MaxRetries                 int                    `split_words:"true" default:"2"`
	CleanContextSchedule       string                 `split_words:"true" default:"0 9 * * *"`
	IntegrationChanRateLimit   float64                `split_words:"true" default:"20"`
	SaleforceChanRateLimit     float64                `split_words:"true" default:"20"`
	Messages                   models.MessageTemplate `split_words:"true" required:"true" default:"{\"waitAgent\":\"Esperando un agente\",\"welcomeTemplate\":\"Hola soy %s y necesito ayuda\",\"context\":\"Contexto\",\"DescriptionCase\":\"Caso levantado por el Bot\",\"uploadImageError\":\"Imagen no enviada\",\"uploadImageSuccess\":\"**El usuario adjunto una imagen al caso**\",\"uploadFileError\":\"Archivo no enviado\",\"uploadFileSuccess\":\"**El usuario adjunto un archivo al caso**\",\"queuePosition\":\"Posici\u00F3n en la cola\",\"waitTime\":\"Tiempo de espera\",\"firstNameContact\":\"Contacto Bot - \",\"clientLabel\":\"Cliente\",\"botLabel\":\"Bot\"}"`
	Timezone                   string                 `required:"true" default:"America/Mexico_City"`
	SendImageNameInMessage     bool                   `split_words:"true" default:"false"`
	KafkaHost                  string                 `required:"true" split_words:"true"`
	KafkaPort                  string                 `required:"true" split_words:"true"`
	KafkaUser                  string                 `required:"true" split_words:"true"`
	KafkaPassword              string                 `required:"true" split_words:"true"`
	KafkaTopic                 string                 `required:"true" split_words:"true"`
}

type Provider struct {
	ButtonID string `json:"button_id"`
	OwnerID  string `json:"owner_id"`
}

type SourceFlowBot struct {
	Subject   string              `json:"subject"`
	Providers map[string]Provider `json:"providers"`
}

type SfcSourceFlowBot map[string]SourceFlowBot

//Decode Decoder this function deserializes the struct by the envconfig Decoder interface implementation
func (sd *SfcSourceFlowBot) Decode(value string) error {
	providerMap := map[string]SourceFlowBot{}

	pairs := strings.Split(value, ";")
	for _, pair := range pairs {
		sourceFlowBotData := SourceFlowBot{}
		kvpair := strings.Split(pair, "=")
		if len(kvpair) != 2 {
			return fmt.Errorf("invalid map item: %q", pair)
		}

		err := json.Unmarshal([]byte(kvpair[1]), &sourceFlowBotData)
		if err != nil {
			return fmt.Errorf("invalid map json: %w", err)
		}

		providerMap[kvpair[0]] = sourceFlowBotData

	}
	*sd = SfcSourceFlowBot(providerMap)

	return nil
}
