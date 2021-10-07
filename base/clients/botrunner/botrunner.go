package botrunner

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"

	"yalochat.com/salesforce-integration/base/clients/proxy"
	"yalochat.com/salesforce-integration/base/helpers"
)

const (
	missingError    = "Missing fields error"
	forwardError    = "Error forwarding the request through the Proxy"
	unmarshallError = "Error unmarshalling the response from the upstream server"
	statusError     = "Error call with status"
)

type BotRunner struct {
	Proxy proxy.ProxyInterface
	Token string
}

type BotRunnerInterface interface {
	SendTo(object map[string]interface{}) (bool, error)
}

func NewBotrunnerClient(url, token string) *BotRunner {
	return &BotRunner{
		Proxy: proxy.NewProxy(url, 30),
		Token: token,
	}
}

func GetRequestToSendTo(botSlug, userId, state, message string) map[string]interface{} {
	requestBody := make(map[string]interface{})
	requestBody["userId"] = userId
	requestBody["botSlug"] = botSlug
	requestBody["state"] = state
	requestBody["message"] = message
	//requestBody["clientBot"] = UserBot
	//requestBody["conversationId"] = conversationId
	//requestBody["clientId"] = userID
	//requestBody["extraInfo"] = extraInfo
	return requestBody
}

// SendTo attempt to forward a request to the given proxy. Some
// business logic to filter is made here.
func (c *BotRunner) SendTo(object map[string]interface{}) (bool, error) {
	if _, ok := object["state"]; !ok {
		logrus.WithFields(logrus.Fields{
			"object": object,
		}).Warn("Invalid state received")

		return false, errors.New(fmt.Sprintf("%s: %s", missingError, "Invalid state received"))
	}

	if _, ok := object["userId"]; !ok {
		logrus.WithFields(logrus.Fields{
			"object": object,
		}).Warn("Invalid userId received")

		return false, errors.New(fmt.Sprintf("%s: %s", missingError, "Invalid userId received"))
	}

	if _, ok := object["message"]; !ok {
		logrus.WithFields(logrus.Fields{
			"object": object,
		}).Warn("Invalid message received")

		return false, errors.New(fmt.Sprintf("%s: %s", missingError, "Invalid message received"))
	}

	if _, ok := object["botSlug"]; !ok {
		logrus.WithFields(logrus.Fields{
			"object": object,
		}).Warn("Invalid botSlug received")

		return false, errors.New(fmt.Sprintf("%s: %s", missingError, "Invalid botSlug received"))
	}

	botSlug := object["botSlug"]
	delete(object, "botSlug")

	//building request to send through proxy
	requestBytes, _ := json.Marshal(object)

	header := make(map[string]string)
	header["Content-Type"] = "application/json"

	var url string
	if c.Token != "" {
		url = fmt.Sprintf("/%s/send-to?jwt=%s", botSlug, c.Token)
	} else {
		url = fmt.Sprintf("/%s/send-to", botSlug)
	}

	newRequest := proxy.Request{
		Body:      requestBytes,
		Method:    "POST",
		URI:       url,
		HeaderMap: header,
	}

	proxiedResponse, proxyError := c.Proxy.SendHTTPRequest(&newRequest)
	if proxyError != nil {
		logrus.WithFields(logrus.Fields{
			"error": proxyError,
		}).Error(forwardError)

		return false, errors.New(fmt.Sprintf("%s: %s", forwardError, proxyError.Error()))
	}

	resultJSON := map[string]interface{}{}
	proxiedError := helpers.ReadAndUnmarshal(proxiedResponse.Body, &resultJSON)

	if proxiedError != nil {
		logrus.WithFields(logrus.Fields{
			"error": proxiedError,
		}).Error(unmarshallError)

		return false, errors.New(fmt.Sprintf("%s: %s", unmarshallError, proxiedError.Error()))
	}

	//do something with resultJSON

	if proxiedResponse.StatusCode != 200 {
		logrus.WithFields(logrus.Fields{
			"error": proxiedResponse.StatusCode,
		}).Error(statusError)

		return false, errors.New(fmt.Sprintf("%s: %d", statusError, proxiedResponse.StatusCode))
	}

	bytesBody, _ := helpers.MarshalJSON(resultJSON)
	if string(bytesBody) == "{}" || string(bytesBody) == "" {
		logrus.WithFields(logrus.Fields{
			"botSlug": botSlug,
			"payload": object,
		}).Info("Send message to send-to")
		return true, nil
	}

	//check this one if this is a response success
	logrus.WithFields(logrus.Fields{
		"botSlug": botSlug,
		"payload": object,
	}).Info("Send message to send-to")
	return true, nil
}
