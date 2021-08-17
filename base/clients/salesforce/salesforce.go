package salesforce

import (
	"yalochat.com/salesforce-integration/base/clients/proxy"
)

type SalesforceClient struct {
	ApiVersion  string
	AccessToken string
	Proxy       proxy.ProxyInterface
}

type SaleforceInterface interface {
	CreateCase() error
}
