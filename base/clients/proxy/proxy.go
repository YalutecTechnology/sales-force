package proxy

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/hashicorp/go-retryablehttp"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	"yalochat.com/salesforce-integration/base/events"

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

func NewProxy(baseUrl string, timeout int, maxRetries int, minRetryWait int, maxRetryWait int) *Proxy {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = maxRetries
	retryClient.RetryWaitMin = time.Duration(minRetryWait) * time.Second
	retryClient.RetryWaitMax = time.Duration(maxRetryWait) * time.Second
	retryClient.HTTPClient.Timeout = time.Second * time.Duration(timeout)

	return &Proxy{
		Client: httptrace.WrapClient(retryClient.StandardClient()),

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

// ProxyInterface Interface Define actions on Proxy struct
type ProxyInterface interface {
	SendHTTPRequest(mainSpan tracer.Span, request *Request) (*http.Response, error)
}

// SendHTTPRequest Sends the HTTP `request` to the `${BaseURL}${uri}` path
func (proxy *Proxy) SendHTTPRequest(mainSpan tracer.Span, request *Request) (*http.Response, error) {
	// datadog tracing
	spanContext := events.GetSpanContextFromSpan(mainSpan)
	span := tracer.StartSpan("proxy.SendHTTPRequest", tracer.ChildOf(spanContext))
	span.SetTag(ext.ResourceName, fmt.Sprintf("%s %s", request.Method, request.URI))
	span.SetTag(ext.SpanTypeHTTP, true)
	span.SetTag(ext.AnalyticsEvent, true)
	defer span.Finish()

	if proxy.BaseURL == "" {
		err := fmt.Errorf("the BaseURL of the instance is empty")
		span.SetTag(ext.Error, err)
		return nil, err
	}

	if request.URI == "" || request.Method == "" {
		err := fmt.Errorf("URI and Method are required in a request")
		span.SetTag(ext.Error, err)
		return nil, err
	}

	absolutePath := fmt.Sprintf("%s%s", proxy.BaseURL, request.URI)
	logrus.WithFields(logrus.Fields{
		"baseUrl": proxy.BaseURL,
		"path":    request.URI,
		"method":  request.Method,
	}).Info("Sending request")

	var body io.Reader
	if request.DataEncode != nil {
		span.SetTag(events.Payload, request.DataEncode)
		body = strings.NewReader(request.DataEncode.Encode())
	} else {
		span.SetTag(events.Payload, string(request.Body))
		body = bytes.NewReader(request.Body)
	}

	newRequest, err := NewRequest(request.Method, absolutePath, body)
	if err != nil {
		logrus.WithError(err).Error("Error making a request")
		span.SetTag(ext.Error, err)
		return nil, fmt.Errorf("Error making a request" + err.Error())
	}
	span.SetTag(ext.HTTPMethod, newRequest.Method)
	span.SetTag(ext.HTTPURL, absolutePath)

	if request.Header != nil {
		newRequest.Header = request.Header
	}

	for key, value := range request.HeaderMap {
		newRequest.Header.Add(key, value)
	}
	span.SetTag("headers", newRequest.Header)

	newRequest.Close = true
	response, err := proxy.Client.Do(newRequest)

	if err != nil {
		logrus.WithError(err).Error("Error proxying a request")
		span.SetTag(ext.Error, err)
		return nil, fmt.Errorf("Error proxying a request" + err.Error())
	}

	span.SetTag(ext.HTTPCode, response.StatusCode)
	if response.StatusCode < http.StatusOK || response.StatusCode > 299 {
		logrus.Info(fmt.Sprintf("Response with status %d", response.StatusCode))
		span.SetTag(ext.Error, errors.New(fmt.Sprintf("Response with status %d", response.StatusCode)))
	}
	return response, nil
}
