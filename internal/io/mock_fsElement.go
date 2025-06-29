// Code generated by mockery. DO NOT EDIT.

package io

import (
	schema "github.com/desertwitch/gover/internal/schema"
	mock "github.com/stretchr/testify/mock"
)

// mockFsElement is an autogenerated mock type for the fsElement type
type mockFsElement struct {
	mock.Mock
}

type mockFsElement_Expecter struct {
	mock *mock.Mock
}

func (_m *mockFsElement) EXPECT() *mockFsElement_Expecter {
	return &mockFsElement_Expecter{mock: &_m.Mock}
}

// GetDestPath provides a mock function with no fields
func (_m *mockFsElement) GetDestPath() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetDestPath")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// mockFsElement_GetDestPath_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetDestPath'
type mockFsElement_GetDestPath_Call struct {
	*mock.Call
}

// GetDestPath is a helper method to define mock.On call
func (_e *mockFsElement_Expecter) GetDestPath() *mockFsElement_GetDestPath_Call {
	return &mockFsElement_GetDestPath_Call{Call: _e.mock.On("GetDestPath")}
}

func (_c *mockFsElement_GetDestPath_Call) Run(run func()) *mockFsElement_GetDestPath_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *mockFsElement_GetDestPath_Call) Return(_a0 string) *mockFsElement_GetDestPath_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockFsElement_GetDestPath_Call) RunAndReturn(run func() string) *mockFsElement_GetDestPath_Call {
	_c.Call.Return(run)
	return _c
}

// GetMetadata provides a mock function with no fields
func (_m *mockFsElement) GetMetadata() *schema.Metadata {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetMetadata")
	}

	var r0 *schema.Metadata
	if rf, ok := ret.Get(0).(func() *schema.Metadata); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*schema.Metadata)
		}
	}

	return r0
}

// mockFsElement_GetMetadata_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetMetadata'
type mockFsElement_GetMetadata_Call struct {
	*mock.Call
}

// GetMetadata is a helper method to define mock.On call
func (_e *mockFsElement_Expecter) GetMetadata() *mockFsElement_GetMetadata_Call {
	return &mockFsElement_GetMetadata_Call{Call: _e.mock.On("GetMetadata")}
}

func (_c *mockFsElement_GetMetadata_Call) Run(run func()) *mockFsElement_GetMetadata_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *mockFsElement_GetMetadata_Call) Return(_a0 *schema.Metadata) *mockFsElement_GetMetadata_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockFsElement_GetMetadata_Call) RunAndReturn(run func() *schema.Metadata) *mockFsElement_GetMetadata_Call {
	_c.Call.Return(run)
	return _c
}

// GetSourcePath provides a mock function with no fields
func (_m *mockFsElement) GetSourcePath() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetSourcePath")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// mockFsElement_GetSourcePath_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetSourcePath'
type mockFsElement_GetSourcePath_Call struct {
	*mock.Call
}

// GetSourcePath is a helper method to define mock.On call
func (_e *mockFsElement_Expecter) GetSourcePath() *mockFsElement_GetSourcePath_Call {
	return &mockFsElement_GetSourcePath_Call{Call: _e.mock.On("GetSourcePath")}
}

func (_c *mockFsElement_GetSourcePath_Call) Run(run func()) *mockFsElement_GetSourcePath_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *mockFsElement_GetSourcePath_Call) Return(_a0 string) *mockFsElement_GetSourcePath_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockFsElement_GetSourcePath_Call) RunAndReturn(run func() string) *mockFsElement_GetSourcePath_Call {
	_c.Call.Return(run)
	return _c
}

// newMockFsElement creates a new instance of mockFsElement. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockFsElement(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockFsElement {
	mock := &mockFsElement{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
