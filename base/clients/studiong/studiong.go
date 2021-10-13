package studiong

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"yalochat.com/salesforce-integration/base/clients/proxy"
	"yalochat.com/salesforce-integration/base/constants"
	"yalochat.com/salesforce-integration/base/helpers"
)

type StudioNG struct {
	Proxy proxy.ProxyInterface
	Token string
}

type StudioNGInterface interface {
	SendTo(state, userID string) error
}

func NewStudioNGClient(url, token string) *StudioNG {
	return &StudioNG{
		Proxy: proxy.NewProxy(url, 30),
		Token: token,
	}
}

type StudioNGRequest struct {
	UserID string `json:"userId"`
}

// SendTo attempt to forward a request to the given proxy. Some
// business logic to filter is made here.
func (cs *StudioNG) SendTo(state, userID string) error {
	if state == "" || userID == "" {
		return errors.New("userID and state are required")
	}

	requestBytes, _ := json.Marshal(StudioNGRequest{UserID: userID})
	header := make(map[string]string)
	header["Content-Type"] = "application/json"
	header["Authorization"] = fmt.Sprintf("Bearer %s", cs.Token)

	newRequest := proxy.Request{
		Body:      requestBytes,
		Method:    "POST",
		URI:       fmt.Sprintf("v1/triggers/%s/webhook", state),
		HeaderMap: header,
	}

	proxiedResponse, proxyError := cs.Proxy.SendHTTPRequest(&newRequest)
	if proxyError != nil {
		errorMessage := fmt.Sprintf("%s : %s", constants.ForwardError, proxyError.Error())
		logrus.Error(errorMessage)
		return errors.New(errorMessage)
	}

	if proxiedResponse.StatusCode != http.StatusOK {
		return helpers.ErrorResponseMap(proxiedResponse.Body, constants.StatusError, proxiedResponse.StatusCode)
	}

	logrus.WithFields(logrus.Fields{
		"state":  state,
		"userID": userID,
	}).Info("Send message to send-to")

	return nil
}
