// Code generated by mockery. DO NOT EDIT.

package filesystem

import (
	schema "github.com/desertwitch/gover/internal/schema"
	mock "github.com/stretchr/testify/mock"
)

// mockDiskStatProvider is an autogenerated mock type for the diskStatProvider type
type mockDiskStatProvider struct {
	mock.Mock
}

type mockDiskStatProvider_Expecter struct {
	mock *mock.Mock
}

func (_m *mockDiskStatProvider) EXPECT() *mockDiskStatProvider_Expecter {
	return &mockDiskStatProvider_Expecter{mock: &_m.Mock}
}

// GetDiskUsage provides a mock function with given fields: storage
func (_m *mockDiskStatProvider) GetDiskUsage(storage schema.Storage) (DiskStats, error) {
	ret := _m.Called(storage)

	if len(ret) == 0 {
		panic("no return value specified for GetDiskUsage")
	}

	var r0 DiskStats
	var r1 error
	if rf, ok := ret.Get(0).(func(schema.Storage) (DiskStats, error)); ok {
		return rf(storage)
	}
	if rf, ok := ret.Get(0).(func(schema.Storage) DiskStats); ok {
		r0 = rf(storage)
	} else {
		r0 = ret.Get(0).(DiskStats)
	}

	if rf, ok := ret.Get(1).(func(schema.Storage) error); ok {
		r1 = rf(storage)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockDiskStatProvider_GetDiskUsage_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'GetDiskUsage'
type mockDiskStatProvider_GetDiskUsage_Call struct {
	*mock.Call
}

// GetDiskUsage is a helper method to define mock.On call
//   - storage schema.Storage
func (_e *mockDiskStatProvider_Expecter) GetDiskUsage(storage interface{}) *mockDiskStatProvider_GetDiskUsage_Call {
	return &mockDiskStatProvider_GetDiskUsage_Call{Call: _e.mock.On("GetDiskUsage", storage)}
}

func (_c *mockDiskStatProvider_GetDiskUsage_Call) Run(run func(storage schema.Storage)) *mockDiskStatProvider_GetDiskUsage_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(schema.Storage))
	})
	return _c
}

func (_c *mockDiskStatProvider_GetDiskUsage_Call) Return(_a0 DiskStats, _a1 error) *mockDiskStatProvider_GetDiskUsage_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockDiskStatProvider_GetDiskUsage_Call) RunAndReturn(run func(schema.Storage) (DiskStats, error)) *mockDiskStatProvider_GetDiskUsage_Call {
	_c.Call.Return(run)
	return _c
}

// HasEnoughFreeSpace provides a mock function with given fields: s, minFree, fileSize
func (_m *mockDiskStatProvider) HasEnoughFreeSpace(s schema.Storage, minFree uint64, fileSize uint64) (bool, error) {
	ret := _m.Called(s, minFree, fileSize)

	if len(ret) == 0 {
		panic("no return value specified for HasEnoughFreeSpace")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(schema.Storage, uint64, uint64) (bool, error)); ok {
		return rf(s, minFree, fileSize)
	}
	if rf, ok := ret.Get(0).(func(schema.Storage, uint64, uint64) bool); ok {
		r0 = rf(s, minFree, fileSize)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(schema.Storage, uint64, uint64) error); ok {
		r1 = rf(s, minFree, fileSize)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// mockDiskStatProvider_HasEnoughFreeSpace_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'HasEnoughFreeSpace'
type mockDiskStatProvider_HasEnoughFreeSpace_Call struct {
	*mock.Call
}

// HasEnoughFreeSpace is a helper method to define mock.On call
//   - s schema.Storage
//   - minFree uint64
//   - fileSize uint64
func (_e *mockDiskStatProvider_Expecter) HasEnoughFreeSpace(s interface{}, minFree interface{}, fileSize interface{}) *mockDiskStatProvider_HasEnoughFreeSpace_Call {
	return &mockDiskStatProvider_HasEnoughFreeSpace_Call{Call: _e.mock.On("HasEnoughFreeSpace", s, minFree, fileSize)}
}

func (_c *mockDiskStatProvider_HasEnoughFreeSpace_Call) Run(run func(s schema.Storage, minFree uint64, fileSize uint64)) *mockDiskStatProvider_HasEnoughFreeSpace_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(schema.Storage), args[1].(uint64), args[2].(uint64))
	})
	return _c
}

func (_c *mockDiskStatProvider_HasEnoughFreeSpace_Call) Return(_a0 bool, _a1 error) *mockDiskStatProvider_HasEnoughFreeSpace_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *mockDiskStatProvider_HasEnoughFreeSpace_Call) RunAndReturn(run func(schema.Storage, uint64, uint64) (bool, error)) *mockDiskStatProvider_HasEnoughFreeSpace_Call {
	_c.Call.Return(run)
	return _c
}

// newMockDiskStatProvider creates a new instance of mockDiskStatProvider. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockDiskStatProvider(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockDiskStatProvider {
	mock := &mockDiskStatProvider{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
