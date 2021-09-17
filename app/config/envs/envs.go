package envs

// Envs represents the list of well known env vars used by the app
type Envs struct {
	AppName               string   `default:"salesforce-integration" split_words:"true"`
	Host                  string   `required:"true" split_words:"true" default:"localhost"`
	Port                  string   `required:"true" split_words:"true" default:"8080"`
	SentryDSN             string   `default:"" split_words:"true"`
	Environment           string   `default:"dev" split_words:"true"`
	MainContextTimeOut    int16    `default:"10" split_words:"true"`
	RedisAddress          string   `split_words:"true"`
	RedisMaster           string   `split_words:"true"`
	RedisSentinelAddress  string   `split_words:"true"`
	BlockedUserState      string   `required:"true" split_words:"true" default:"from-sf-blocked"`
	TimeoutState          string   `required:"true" split_words:"true" default:"from-sf-timeout"`
	SuccessState          string   `required:"true" split_words:"true" default:"from-sf-success"`
	YaloUsername          string   `required:"true" split_words:"true" default:"yaloUser"`
	YaloPassword          string   `required:"true" split_words:"true"`
	SalesforceUsername    string   `required:"true" split_words:"true" default:"salesforceUser"`
	SalesforcePassword    string   `required:"true" split_words:"true"`
	SecretKey             string   `required:"true" split_words:"true"`
	BotrunnerUrl          string   `split_words:"true"`
	BotrunnerToken        string   `split_words:"true" default:""`
	BotrunnerTimeout      int      `split_words:"true" default:"4"`
	SfcClientId           string   `split_words:"true"`
	SfcClientSecret       string   `split_words:"true"`
	SfcUsername           string   `split_words:"true"`
	SfcPassword           string   `split_words:"true"`
	SfcSecurityToken      string   `split_words:"true"`
	SfcBaseUrl            string   `split_words:"true"`
	SfcChatUrl            string   `split_words:"true"`
	SfcLoginUrl           string   `split_words:"true"`
	SfcApiVersion         string   `split_words:"true" default:"52"`
	SfcOrganizationId     string   `split_words:"true"`
	SfcDeploymentId       string   `split_words:"true"`
	SfcWAButtonId         string   `split_words:"true"`
	SfcFBButtonId         string   `split_words:"true"`
	SfcWAOwnerId          string   `split_words:"true"`
	SfcFBOwnerId          string   `split_words:"true"`
	SfcRecordTypeId       string   `split_words:"true"`
	SfcCustomFieldsCase   []string `split_words:"true"`
	IntegrationsChannel   string   `split_words:"true" default:"outgoing_webhook"`
	IntegrationsBotId     string   `split_words:"true"`
	IntegrationsBotJWT    string   `split_words:"true"`
	IntegrationsBaseUrl   string   `split_words:"true"`
	IntegrationsSignature string   `split_words:"true"`
	WebhookBaseUrl        string   `split_words:"true"`
	IntegrationsBotPhone  string   `split_words:"true"`
	KeywordsRestart       []string `split_words:"true" default:"coppelbot,regresar,reiniciar,restart"`
}
