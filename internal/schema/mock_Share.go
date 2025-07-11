// Code generated by mockery. DO NOT EDIT.

package schema

import mock "github.com/stretchr/testify/mock"

// MockShare is an autogenerated mock type for the Share type
type MockShare struct {
	mock.Mock
}

type MockShare_Expecter struct {
	mock *mock.Mock
}

func (_m *MockShare) EXPECT() *MockShare_Expecter {
	return &MockShare_Expecter{mock: &_m.Mock}
}

// GetAllocator provides a mock function with no fields
func (_m *MockShare) GetAllocator() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetAllocator")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockShare_GetAllocator_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetAllocator'
type MockShare_GetAllocator_Call struct {
	*mock.Call
}

// GetAllocator is a helper method to define mock.On call
func (_e *MockShare_Expecter) GetAllocator() *MockShare_GetAllocator_Call {
	return &MockShare_GetAllocator_Call{Call: _e.mock.On("GetAllocator")}
}

func (_c *MockShare_GetAllocator_Call) Run(run func()) *MockShare_GetAllocator_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockShare_GetAllocator_Call) Return(_a0 string) *MockShare_GetAllocator_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockShare_GetAllocator_Call) RunAndReturn(run func() string) *MockShare_GetAllocator_Call {
	_c.Call.Return(run)
	return _c
}

// GetCachePool provides a mock function with no fields
func (_m *MockShare) GetCachePool() Pool {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetCachePool")
	}

	var r0 Pool
	if rf, ok := ret.Get(0).(func() Pool); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(Pool)
		}
	}

	return r0
}

// MockShare_GetCachePool_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetCachePool'
type MockShare_GetCachePool_Call struct {
	*mock.Call
}

// GetCachePool is a helper method to define mock.On call
func (_e *MockShare_Expecter) GetCachePool() *MockShare_GetCachePool_Call {
	return &MockShare_GetCachePool_Call{Call: _e.mock.On("GetCachePool")}
}

func (_c *MockShare_GetCachePool_Call) Run(run func()) *MockShare_GetCachePool_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockShare_GetCachePool_Call) Return(_a0 Pool) *MockShare_GetCachePool_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockShare_GetCachePool_Call) RunAndReturn(run func() Pool) *MockShare_GetCachePool_Call {
	_c.Call.Return(run)
	return _c
}

// GetCachePool2 provides a mock function with no fields
func (_m *MockShare) GetCachePool2() Pool {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetCachePool2")
	}

	var r0 Pool
	if rf, ok := ret.Get(0).(func() Pool); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(Pool)
		}
	}

	return r0
}

// MockShare_GetCachePool2_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetCachePool2'
type MockShare_GetCachePool2_Call struct {
	*mock.Call
}

// GetCachePool2 is a helper method to define mock.On call
func (_e *MockShare_Expecter) GetCachePool2() *MockShare_GetCachePool2_Call {
	return &MockShare_GetCachePool2_Call{Call: _e.mock.On("GetCachePool2")}
}

func (_c *MockShare_GetCachePool2_Call) Run(run func()) *MockShare_GetCachePool2_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockShare_GetCachePool2_Call) Return(_a0 Pool) *MockShare_GetCachePool2_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockShare_GetCachePool2_Call) RunAndReturn(run func() Pool) *MockShare_GetCachePool2_Call {
	_c.Call.Return(run)
	return _c
}

// GetDisableCOW provides a mock function with no fields
func (_m *MockShare) GetDisableCOW() bool {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetDisableCOW")
	}

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	return r0
}

// MockShare_GetDisableCOW_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetDisableCOW'
type MockShare_GetDisableCOW_Call struct {
	*mock.Call
}

// GetDisableCOW is a helper method to define mock.On call
func (_e *MockShare_Expecter) GetDisableCOW() *MockShare_GetDisableCOW_Call {
	return &MockShare_GetDisableCOW_Call{Call: _e.mock.On("GetDisableCOW")}
}

func (_c *MockShare_GetDisableCOW_Call) Run(run func()) *MockShare_GetDisableCOW_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockShare_GetDisableCOW_Call) Return(_a0 bool) *MockShare_GetDisableCOW_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockShare_GetDisableCOW_Call) RunAndReturn(run func() bool) *MockShare_GetDisableCOW_Call {
	_c.Call.Return(run)
	return _c
}

// GetIncludedDisks provides a mock function with no fields
func (_m *MockShare) GetIncludedDisks() map[string]Disk {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetIncludedDisks")
	}

	var r0 map[string]Disk
	if rf, ok := ret.Get(0).(func() map[string]Disk); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]Disk)
		}
	}

	return r0
}

// MockShare_GetIncludedDisks_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetIncludedDisks'
type MockShare_GetIncludedDisks_Call struct {
	*mock.Call
}

// GetIncludedDisks is a helper method to define mock.On call
func (_e *MockShare_Expecter) GetIncludedDisks() *MockShare_GetIncludedDisks_Call {
	return &MockShare_GetIncludedDisks_Call{Call: _e.mock.On("GetIncludedDisks")}
}

func (_c *MockShare_GetIncludedDisks_Call) Run(run func()) *MockShare_GetIncludedDisks_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockShare_GetIncludedDisks_Call) Return(_a0 map[string]Disk) *MockShare_GetIncludedDisks_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockShare_GetIncludedDisks_Call) RunAndReturn(run func() map[string]Disk) *MockShare_GetIncludedDisks_Call {
	_c.Call.Return(run)
	return _c
}

// GetName provides a mock function with no fields
func (_m *MockShare) GetName() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetName")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockShare_GetName_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetName'
type MockShare_GetName_Call struct {
	*mock.Call
}

// GetName is a helper method to define mock.On call
func (_e *MockShare_Expecter) GetName() *MockShare_GetName_Call {
	return &MockShare_GetName_Call{Call: _e.mock.On("GetName")}
}

func (_c *MockShare_GetName_Call) Run(run func()) *MockShare_GetName_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockShare_GetName_Call) Return(_a0 string) *MockShare_GetName_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockShare_GetName_Call) RunAndReturn(run func() string) *MockShare_GetName_Call {
	_c.Call.Return(run)
	return _c
}

// GetSpaceFloor provides a mock function with no fields
func (_m *MockShare) GetSpaceFloor() uint64 {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetSpaceFloor")
	}

	var r0 uint64
	if rf, ok := ret.Get(0).(func() uint64); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(uint64)
	}

	return r0
}

// MockShare_GetSpaceFloor_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetSpaceFloor'
type MockShare_GetSpaceFloor_Call struct {
	*mock.Call
}

// GetSpaceFloor is a helper method to define mock.On call
func (_e *MockShare_Expecter) GetSpaceFloor() *MockShare_GetSpaceFloor_Call {
	return &MockShare_GetSpaceFloor_Call{Call: _e.mock.On("GetSpaceFloor")}
}

func (_c *MockShare_GetSpaceFloor_Call) Run(run func()) *MockShare_GetSpaceFloor_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockShare_GetSpaceFloor_Call) Return(_a0 uint64) *MockShare_GetSpaceFloor_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockShare_GetSpaceFloor_Call) RunAndReturn(run func() uint64) *MockShare_GetSpaceFloor_Call {
	_c.Call.Return(run)
	return _c
}

// GetSplitLevel provides a mock function with no fields
func (_m *MockShare) GetSplitLevel() int {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetSplitLevel")
	}

	var r0 int
	if rf, ok := ret.Get(0).(func() int); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// MockShare_GetSplitLevel_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetSplitLevel'
type MockShare_GetSplitLevel_Call struct {
	*mock.Call
}

// GetSplitLevel is a helper method to define mock.On call
func (_e *MockShare_Expecter) GetSplitLevel() *MockShare_GetSplitLevel_Call {
	return &MockShare_GetSplitLevel_Call{Call: _e.mock.On("GetSplitLevel")}
}

func (_c *MockShare_GetSplitLevel_Call) Run(run func()) *MockShare_GetSplitLevel_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockShare_GetSplitLevel_Call) Return(_a0 int) *MockShare_GetSplitLevel_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockShare_GetSplitLevel_Call) RunAndReturn(run func() int) *MockShare_GetSplitLevel_Call {
	_c.Call.Return(run)
	return _c
}

// GetUseCache provides a mock function with no fields
func (_m *MockShare) GetUseCache() string {
	ret := _m.Called()

	if len(ret) == 0 {
		panic("no return value specified for GetUseCache")
	}

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// MockShare_GetUseCache_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetUseCache'
type MockShare_GetUseCache_Call struct {
	*mock.Call
}

// GetUseCache is a helper method to define mock.On call
func (_e *MockShare_Expecter) GetUseCache() *MockShare_GetUseCache_Call {
	return &MockShare_GetUseCache_Call{Call: _e.mock.On("GetUseCache")}
}

func (_c *MockShare_GetUseCache_Call) Run(run func()) *MockShare_GetUseCache_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run()
	})
	return _c
}

func (_c *MockShare_GetUseCache_Call) Return(_a0 string) *MockShare_GetUseCache_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *MockShare_GetUseCache_Call) RunAndReturn(run func() string) *MockShare_GetUseCache_Call {
	_c.Call.Return(run)
	return _c
}

// NewMockShare creates a new instance of MockShare. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewMockShare(t interface {
	mock.TestingT
	Cleanup(func())
}) *MockShare {
	mock := &MockShare{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
