// Code generated by mockery v2.15.0. DO NOT EDIT.

package mocks

import (
	http "net/http"

	ddtrace "gopkg.in/DataDog/dd-trace-go.v1/ddtrace"

	mock "github.com/stretchr/testify/mock"

	proxy "yalochat.com/salesforce-integration/base/clients/proxy"
)

// ProxyInterface is an autogenerated mock type for the ProxyInterface type
type ProxyInterface struct {
	mock.Mock
}

// SendHTTPRequest provides a mock function with given fields: mainSpan, request
func (_m *ProxyInterface) SendHTTPRequest(mainSpan ddtrace.Span, request *proxy.Request) (*http.Response, error) {
	ret := _m.Called(mainSpan, request)

	var r0 *http.Response
	if rf, ok := ret.Get(0).(func(ddtrace.Span, *proxy.Request) *http.Response); ok {
		r0 = rf(mainSpan, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*http.Response)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(ddtrace.Span, *proxy.Request) error); ok {
		r1 = rf(mainSpan, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewProxyInterface interface {
	mock.TestingT
	Cleanup(func())
}

// NewProxyInterface creates a new instance of ProxyInterface. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewProxyInterface(t mockConstructorTestingTNewProxyInterface) *ProxyInterface {
	mock := &ProxyInterface{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}