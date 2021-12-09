package login

import (
	"errors"
	"fmt"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"net/http"
	"net/url"
	"yalochat.com/salesforce-integration/base/events"

	"github.com/sirupsen/logrus"
	"yalochat.com/salesforce-integration/base/clients/proxy"
	"yalochat.com/salesforce-integration/base/constants"
	"yalochat.com/salesforce-integration/base/helpers"
)

type SfcLoginClient struct {
	Proxy proxy.ProxyInterface
}

type TokenPayload struct {
	GrantType    string `json:"grant_type" validate:"required"`
	ClientId     string `json:"client_id" validate:"required"`
	ClientSecret string `json:"client_secret" validate:"required"`
	Username     string `json:"username" validate:"required"`
	Password     string `json:"password" validate:"required"`
}

type SfcLoginInterface interface {
	GetToken(TokenPayload) (string, error)
}

// GetToken Get Access Token for Salesforce Requests
func (c *SfcLoginClient) GetToken(tokenPayload TokenPayload) (string, error) {
	// datadog tracing
	span := tracer.StartSpan("get_token")
	span.SetTag(ext.AnalyticsEvent, true)
	span.SetTag(events.Payload, fmt.Sprintf("%#v", tokenPayload))
	defer span.Finish()
	uri := "/services/oauth2/token"
	span.SetTag(ext.ResourceName, fmt.Sprintf("%s %s", http.MethodPost, uri))

	var errorMessage string
	if tokenPayload.GrantType == "" {
		tokenPayload.GrantType = "password"
	}

	// This log should hide the secrets before sending to production
	/*logrus.WithFields(logrus.Fields{
		"payload": tokenPayload,
	}).Info("Payload received")*/

	//validating token Payload struct
	if err := helpers.Govalidator().Struct(tokenPayload); err != nil {
		errorMessage = fmt.Sprintf("%s : %s", helpers.InvalidPayload, err.Error())
		logrus.Error(errorMessage)
		span.SetTag(ext.Error, err)
		return "", errors.New(errorMessage)
	}

	//building request to send through proxy
	dataEncode := url.Values{}
	dataEncode.Set("grant_type", tokenPayload.GrantType)
	dataEncode.Set("client_id", tokenPayload.ClientId)
	dataEncode.Set("client_secret", tokenPayload.ClientSecret)
	dataEncode.Set("username", tokenPayload.Username)
	dataEncode.Set("password", tokenPayload.Password)

	header := make(map[string]string)
	header["Content-Type"] = proxy.FormUrlencodeHeader

	newRequest := proxy.Request{
		DataEncode: dataEncode,
		Method:     http.MethodPost,
		URI:        uri,
		HeaderMap:  header,
	}

	proxiedResponse, proxyError := c.Proxy.SendHTTPRequest(span, &newRequest)
	if proxyError != nil {
		errorMessage = fmt.Sprintf("%s : %s", constants.ForwardError, proxyError.Error())
		logrus.Error(errorMessage)
		span.SetTag(ext.Error, proxyError)
		return "", errors.New(errorMessage)
	}

	responseMap := map[string]interface{}{}
	readAndUnmarshalError := helpers.ReadAndUnmarshal(proxiedResponse.Body, &responseMap)

	if readAndUnmarshalError != nil {
		errorMessage = fmt.Sprintf("%s : %s", constants.UnmarshallError, readAndUnmarshalError.Error())
		logrus.Error(errorMessage)
		span.SetTag(ext.Error, readAndUnmarshalError)
		return "", errors.New(errorMessage)
	}

	if proxiedResponse.StatusCode != 200 {
		errorMessage = fmt.Sprintf("%s : %d", constants.StatusError, proxiedResponse.StatusCode)
		logrus.WithFields(logrus.Fields{
			"response": responseMap,
		}).Error(errorMessage)
		err := errors.New(errorMessage)
		span.SetTag(ext.Error, err)
		return "", err
	}

	//check this one if this is a response success
	/*logrus.WithFields(logrus.Fields{
		"response": responseMap,
	}).Info("Get accessToken sucessfully")*/
	logrus.Info("Get accessToken successfully")

	if _, ok := responseMap["access_token"]; ok {
		return responseMap["access_token"].(string), nil
	}
	err := errors.New("could not get accessToken in response")
	span.SetTag(ext.Error, err)
	return "", err
}
