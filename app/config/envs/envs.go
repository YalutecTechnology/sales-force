package envs

// Envs represents the list of well known env vars used by the app
type Envs struct {
	AppName              string `default:"salesforce-integration" split_words:"true"`
	Host                 string `required:"true" split_words:"true" default:"localhost"`
	Port                 string `required:"true" split_words:"true" default:"8080"`
	SentryDSN            string `default:"" split_words:"true"`
	Environment          string `default:"dev" split_words:"true"`
	MainContextTimeOut   int16  `default:"10" split_words:"true"`
	RedisAddress         string `split_words:"true" default:"localhost:6379"`
	RedisMaster          string `split_words:"true"`
	RedisSentinelAddress string `split_words:"true"`
	YaloUsername         string `required:"true" split_words:"true" default:"yaloUser"`
	YaloPassword         string `required:"true" split_words:"true"`
	SalesforceUsername   string `required:"true" split_words:"true" default:"salesforceUser"`
	SalesforcePassword   string `required:"true" split_words:"true"`
	SecretKey            string `required:"true" split_words:"true"`
	BotrunnerUrl         string `split_words:"true"`
	SfcClientId          string `split_words:"true"`
	SfcClientSecret      string `split_words:"true"`
	SfcUsername          string `split_words:"true"`
	SfcPassword          string `split_words:"true"`
	SfcSecurityToken     string `split_words:"true"`
	SfcBaseUrl           string `split_words:"true"`
	SfcChatUrl           string `split_words:"true"`
	SfcLoginUrl          string `split_words:"true"`
	SfcApiVersion        string `split_words:"true" default:"52"`
	SfcOrganizationId    string `split_words:"true"`
	SfcDeploymentId      string `split_words:"true"`
	SfcButtonId          string `split_words:"true"`
	SfcOwnerId           string `split_words:"true"`
	IntegrationsChannel  string `split_words:"true" default:"outgoing_webhook"`
	IntegrationsBotId    string `split_words:"true"`
	IntegrationsBotJWT   string `split_words:"true"`
	IntegrationsBaseUrl  string `split_words:"true"`
}
