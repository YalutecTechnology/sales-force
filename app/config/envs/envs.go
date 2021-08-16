package envs

// Envs represents the list of well known env vars used by the app
type Envs struct {
	AppName                 string `default:"salesforce-integration" split_words:"true"`
	Host                    string `required:"true" split_words:"true" default:"localhost"`
	Port                    string `required:"true" split_words:"true" default:"8080"`
	SentryDSN               string `default:"" split_words:"true"`
	Environment             string `default:"dev" split_words:"true"`
	MainContextTimeOut      int16  `default:"10" split_words:"true"`
	RedisAddress            string `split_words:"true" default:"localhost:6379"`
	RedisMaster             string `split_words:"true"`
	RedisSentinelAddress    string `split_words:"true"`
	YaloUsername            string `required:"true" split_words:"true" default:"yaloUser"`
	YaloPassword            string `required:"true" split_words:"true" default:"yaloPassword"`
	SalesforceUsername      string `required:"true" split_words:"true" default:"salesforceUser"`
	SalesforcePassword      string `required:"true" split_words:"true" default:"salesforcePassword"`
	SecretKey               string `required:"true" split_words:"true" default:"salesforceUser"`
	SfscClientId            string `split_words:"true"`
	SfscClientSecret        string `split_words:"true"`
	SfscApiUsername         string `split_words:"true"`
	SfscPassword            string `split_words:"true"`
	SfscBaseUrl             string `split_words:"true"`
	SfscLoginUrl            string `split_words:"true"`
	SfscApiVersion          int    `split_words:"true" default:"52"`
	SfscOrganizationId      string `split_words:"true"`
	SfscDeploymentId        string `split_words:"true"`
	SfscButtonId            string `split_words:"true"`
	SfscOwnerId             string `split_words:"true"`
	SfscContactId           string `split_words:"true"`
	IntegrationsMydomainUrl string `split_words:"true"`
	IntegrationsBotId       string `split_words:"true"`
	IntegrationsBotSlug     string `split_words:"true"`
	IntegrationsJWT         string `split_words:"true"`
	IntegrationsBaseUrl     string `split_words:"true"`
	IntegrationsChannel     string `split_words:"true" default:"outgoing"`
}
