package proxy

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	httptrace "gopkg.in/DataDog/dd-trace-go.v1/contrib/net/http"
)

var NewRequest = http.NewRequest

const (
	FormUrlencodeHeader = "application/x-www-form-urlencoded"
)

// Proxy define a third service that will receive our messages to be sent
type Proxy struct {
	BaseURL string
	Client  *http.Client
}

func NewProxy(baseUrl string) *Proxy {
	return &Proxy{
		Client: httptrace.WrapClient(&http.Client{
			Timeout: time.Second * 10,
		}),

		BaseURL: baseUrl,
	}
}

// Request groups all the values needed to make an HTTP request
type Request struct {
	URI        string
	Method     string
	Header     http.Header
	HeaderMap  map[string]string
	Body       []byte
	DataEncode url.Values
}

// Interface Define actions on Proxy struct
type ProxyInterface interface {
	SendHTTPRequest(request *Request) (*http.Response, error)
}

// SendHTTPRequest Sends the HTTP `request` to the `${BaseURL}${uri}` path
func (proxy *Proxy) SendHTTPRequest(request *Request) (*http.Response, error) {
	if proxy.BaseURL == "" {
		return nil, fmt.Errorf("The BaseURL of the instance is empty")
	}

	if request.URI == "" || request.Method == "" {
		return nil, fmt.Errorf("URI and Method are required in a request")
	}

	absolutePath := fmt.Sprintf("%s%s", proxy.BaseURL, request.URI)
	logrus.WithFields(logrus.Fields{
		"baseUrl": proxy.BaseURL,
		"path":    request.URI,
		"method":  request.Method,
	}).Info("Sending request")

	var body io.Reader
	if request.DataEncode != nil {
		body = strings.NewReader(request.DataEncode.Encode())
	} else {
		body = bytes.NewReader(request.Body)
	}

	newRequest, err := NewRequest(request.Method, absolutePath, body)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"Error making a request": err.Error(),
		}).Error("Error making a request")

		return nil, fmt.Errorf("Error making a request" + err.Error())
	}

	if request.Header != nil {
		newRequest.Header = request.Header
	}

	for key, value := range request.HeaderMap {
		newRequest.Header.Add(key, value)
	}

	response, err := proxy.Client.Do(newRequest)

	if err != nil {
		logrus.WithFields(logrus.Fields{
			"Error proxying a request": err,
		}).Error("Error proxying a request")

		return nil, fmt.Errorf("Error proxying a request" + err.Error())
	}

	return response, nil
}
