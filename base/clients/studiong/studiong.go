package studiong

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"net/http"
	"yalochat.com/salesforce-integration/base/clients/proxy"
	"yalochat.com/salesforce-integration/base/constants"
	"yalochat.com/salesforce-integration/base/events"
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
	// datadog tracing
	span := tracer.StartSpan("send_to")
	span.SetTag(ext.AnalyticsEvent, true)
	span.SetTag(events.UserID, userID)
	defer span.Finish()
	uri := fmt.Sprintf("v1/triggers/%s/webhook", state)
	span.SetTag(ext.ResourceName, fmt.Sprintf("%s %s", http.MethodPost, uri))
	payload := StudioNGRequest{UserID: userID}
	span.SetTag(events.Payload, fmt.Sprintf("%#v", payload))

	if state == "" || userID == "" {
		err := errors.New("userID and state are required")
		span.SetTag(ext.Error, err)
		return err
	}

	requestBytes, _ := json.Marshal(payload)
	header := make(map[string]string)
	header["Content-Type"] = "application/json"
	header["Authorization"] = fmt.Sprintf("Bearer %s", cs.Token)

	newRequest := proxy.Request{
		Body:      requestBytes,
		Method:    http.MethodPost,
		URI:       uri,
		HeaderMap: header,
	}

	proxiedResponse, proxyError := cs.Proxy.SendHTTPRequest(span, &newRequest)
	if proxyError != nil {
		errorMessage := fmt.Sprintf("%s : %s", constants.ForwardError, proxyError.Error())
		logrus.Error(errorMessage)
		span.SetTag(ext.Error, proxyError)
		return errors.New(errorMessage)
	}

	if proxiedResponse.StatusCode != http.StatusOK {
		err := helpers.ErrorResponseMap(proxiedResponse.Body, constants.StatusError, proxiedResponse.StatusCode)
		span.SetTag(ext.Error, err)
		return err
	}

	logrus.WithFields(logrus.Fields{
		"state":  state,
		"userID": userID,
	}).Info("Send message to send-to")

	return nil
}
