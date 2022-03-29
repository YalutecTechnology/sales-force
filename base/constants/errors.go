package constants

const (
	ForwardError               = "Error forwarding the request through the Proxy"
	UnmarshallError            = "Error unmarshalling the response from salesForce"
	StatusError                = "Error call with status"
	RequestError               = "Error making request"
	QueryParamError            = "Error getting query param"
	ResponseError              = "Error getting response, it was empty or format not handled correctly"
	ErrInterconnectionNotFound = applicationErrors("not found interconnection")
)

type applicationErrors string

func (e applicationErrors) Error() string { return string(e) }
