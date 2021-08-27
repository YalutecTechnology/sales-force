package login

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"yalochat.com/salesforce-integration/base/clients/proxy"
	"yalochat.com/salesforce-integration/base/constants"
	"yalochat.com/salesforce-integration/base/helpers"
)

func TestGetToken(t *testing.T) {
	accessTokenResponse := `{"access_token":"00D3g0000003VOm!ARQAQJzlD0cKgBAzx.ot_qHdZnNhebgM.Ijk7an_LdZzN_JUqHasD1GjpeHow5i0TcHmYjtj4cEEL5rMwE7F7mGR9S5eIsi1","instance_url":"https://na110.salesforce.com","id":"https://login.salesforce.com/id/00D3g0000003VOmEAM/0053g000000usWaAAI","token_type":"Bearer","issued_at":"1626975076132","signature":"2UtMAk/S2Xr0HNe73DYSG3UpUzYDP8khlPPlzVpNmco="}`

	t.Run("Get Token Successful", func(t *testing.T) {
		mock := &proxy.Mock{}
		salesforceClient := &SfcLoginClient{Proxy: mock}
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte(accessTokenResponse))),
		}, nil)
		tokenPayload := TokenPayload{
			GrantType:    "password",
			ClientId:     "clientId",
			ClientSecret: "clientSecret",
			Username:     "username",
			Password:     "password",
		}
		token, err := salesforceClient.GetToken(tokenPayload)

		if err != nil {
			t.Fatalf("Expected nil error, but retrieved this %#v", err)
		}

		if token == "" {
			t.Fatalf(`Expected token, but retrieved ""`)
		}
	})

	t.Run("Should fail by invalid payload", func(t *testing.T) {
		expectedError := helpers.InvalidPayload
		salesforceClient := &SfcLoginClient{
			Proxy: &proxy.Proxy{},
		}
		tokenPayload := TokenPayload{
			GrantType: "password",
		}
		token, err := salesforceClient.GetToken(tokenPayload)

		if token != "" {
			t.Fatalf(`Expected "", but retrieved %s`, token)
		}

		if !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("Error message should contain %s, but this was found <%s>", expectedError, err.Error())
		}
	})

	t.Run("Should fail by unmarshall error received", func(t *testing.T) {
		mock := &proxy.Mock{}
		expectedError := constants.UnmarshallError
		salesforceClient := &SfcLoginClient{
			Proxy: proxy.NewProxy("http://salesforce.com"),
		}
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusOK,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte("{Invalid Payload:}"))),
		}, nil)

		tokenPayload := TokenPayload{
			GrantType:    "password",
			ClientId:     "clientId",
			ClientSecret: "clientSecret",
			Username:     "username",
			Password:     "password",
		}
		token, err := salesforceClient.GetToken(tokenPayload)

		if token != "" {
			t.Fatalf(`Expected "", but retrieved %s`, token)
		}

		if !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("Error message should contain %s, but this was found <%s>", expectedError, err.Error())
		}
	})

	t.Run("Should fail by proxy error received", func(t *testing.T) {
		mock := &proxy.Mock{}
		expectedError := constants.ForwardError
		salesforceClient := &SfcLoginClient{
			Proxy: mock,
		}
		mock.On("SendHTTPRequest").Return(&http.Response{}, fmt.Errorf("Error proxying a request"))

		tokenPayload := TokenPayload{
			GrantType:    "password",
			ClientId:     "clientId",
			ClientSecret: "clientSecret",
			Username:     "username",
			Password:     "password",
		}
		token, err := salesforceClient.GetToken(tokenPayload)

		if token != "" {
			t.Fatalf(`Expected "", but retrieved %s`, token)
		}

		if !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("Error message should contain %s, but this was found <%s>", expectedError, err.Error())
		}
	})

	t.Run("Should fail by status error received", func(t *testing.T) {
		mock := &proxy.Mock{}
		expectedError := constants.StatusError
		salesforceClient := &SfcLoginClient{Proxy: mock}
		mock.On("SendHTTPRequest").Return(&http.Response{
			StatusCode: http.StatusInternalServerError,
			Body:       ioutil.NopCloser(bytes.NewReader([]byte("{}"))),
		}, nil)

		tokenPayload := TokenPayload{
			GrantType:    "password",
			ClientId:     "clientId",
			ClientSecret: "clientSecret",
			Username:     "username",
			Password:     "password",
		}
		token, err := salesforceClient.GetToken(tokenPayload)

		if token != "" {
			t.Fatalf(`Expected "", but retrieved %s`, token)
		}

		if !strings.Contains(err.Error(), expectedError) {
			t.Fatalf("Error message should contain %s, but this was found <%s>", expectedError, err.Error())
		}
	})

}
