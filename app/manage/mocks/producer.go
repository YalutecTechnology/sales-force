// Code generated by mockery v2.15.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"

// Producer is an autogenerated mock type for the Producer type
type Producer struct {
	mock.Mock
}

// SendMessage provides a mock function with given fields: msg
func (_m *Producer) SendMessage(msg interface{}) error {
	ret := _m.Called(msg)

	var r0 error
	if rf, ok := ret.Get(0).(func(interface{}) error); ok {
		r0 = rf(msg)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewProducer interface {
	mock.TestingT
	Cleanup(func())
}

// NewProducer creates a new instance of Producer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewProducer(t mockConstructorTestingTNewProducer) *Producer {
	mock := &Producer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
