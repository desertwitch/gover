// Code generated by mockery. DO NOT EDIT.

package io

import (
	mock "github.com/stretchr/testify/mock"
	unix "golang.org/x/sys/unix"
)

// mockUnixProvider is an autogenerated mock type for the unixProvider type
type mockUnixProvider struct {
	mock.Mock
}

type mockUnixProvider_Expecter struct {
	mock *mock.Mock
}

func (_m *mockUnixProvider) EXPECT() *mockUnixProvider_Expecter {
	return &mockUnixProvider_Expecter{mock: &_m.Mock}
}

// Chmod provides a mock function with given fields: path, mode
func (_m *mockUnixProvider) Chmod(path string, mode uint32) error {
	ret := _m.Called(path, mode)

	if len(ret) == 0 {
		panic("no return value specified for Chmod")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, uint32) error); ok {
		r0 = rf(path, mode)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockUnixProvider_Chmod_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Chmod'
type mockUnixProvider_Chmod_Call struct {
	*mock.Call
}

// Chmod is a helper method to define mock.On call
//   - path string
//   - mode uint32
func (_e *mockUnixProvider_Expecter) Chmod(path interface{}, mode interface{}) *mockUnixProvider_Chmod_Call {
	return &mockUnixProvider_Chmod_Call{Call: _e.mock.On("Chmod", path, mode)}
}

func (_c *mockUnixProvider_Chmod_Call) Run(run func(path string, mode uint32)) *mockUnixProvider_Chmod_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(uint32))
	})
	return _c
}

func (_c *mockUnixProvider_Chmod_Call) Return(_a0 error) *mockUnixProvider_Chmod_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockUnixProvider_Chmod_Call) RunAndReturn(run func(string, uint32) error) *mockUnixProvider_Chmod_Call {
	_c.Call.Return(run)
	return _c
}

// Chown provides a mock function with given fields: path, uid, gid
func (_m *mockUnixProvider) Chown(path string, uid int, gid int) error {
	ret := _m.Called(path, uid, gid)

	if len(ret) == 0 {
		panic("no return value specified for Chown")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, int, int) error); ok {
		r0 = rf(path, uid, gid)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockUnixProvider_Chown_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Chown'
type mockUnixProvider_Chown_Call struct {
	*mock.Call
}

// Chown is a helper method to define mock.On call
//   - path string
//   - uid int
//   - gid int
func (_e *mockUnixProvider_Expecter) Chown(path interface{}, uid interface{}, gid interface{}) *mockUnixProvider_Chown_Call {
	return &mockUnixProvider_Chown_Call{Call: _e.mock.On("Chown", path, uid, gid)}
}

func (_c *mockUnixProvider_Chown_Call) Run(run func(path string, uid int, gid int)) *mockUnixProvider_Chown_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(int), args[2].(int))
	})
	return _c
}

func (_c *mockUnixProvider_Chown_Call) Return(_a0 error) *mockUnixProvider_Chown_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockUnixProvider_Chown_Call) RunAndReturn(run func(string, int, int) error) *mockUnixProvider_Chown_Call {
	_c.Call.Return(run)
	return _c
}

// Lchown provides a mock function with given fields: path, uid, gid
func (_m *mockUnixProvider) Lchown(path string, uid int, gid int) error {
	ret := _m.Called(path, uid, gid)

	if len(ret) == 0 {
		panic("no return value specified for Lchown")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, int, int) error); ok {
		r0 = rf(path, uid, gid)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockUnixProvider_Lchown_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Lchown'
type mockUnixProvider_Lchown_Call struct {
	*mock.Call
}

// Lchown is a helper method to define mock.On call
//   - path string
//   - uid int
//   - gid int
func (_e *mockUnixProvider_Expecter) Lchown(path interface{}, uid interface{}, gid interface{}) *mockUnixProvider_Lchown_Call {
	return &mockUnixProvider_Lchown_Call{Call: _e.mock.On("Lchown", path, uid, gid)}
}

func (_c *mockUnixProvider_Lchown_Call) Run(run func(path string, uid int, gid int)) *mockUnixProvider_Lchown_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(int), args[2].(int))
	})
	return _c
}

func (_c *mockUnixProvider_Lchown_Call) Return(_a0 error) *mockUnixProvider_Lchown_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockUnixProvider_Lchown_Call) RunAndReturn(run func(string, int, int) error) *mockUnixProvider_Lchown_Call {
	_c.Call.Return(run)
	return _c
}

// Link provides a mock function with given fields: oldpath, newpath
func (_m *mockUnixProvider) Link(oldpath string, newpath string) error {
	ret := _m.Called(oldpath, newpath)

	if len(ret) == 0 {
		panic("no return value specified for Link")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(oldpath, newpath)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockUnixProvider_Link_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Link'
type mockUnixProvider_Link_Call struct {
	*mock.Call
}

// Link is a helper method to define mock.On call
//   - oldpath string
//   - newpath string
func (_e *mockUnixProvider_Expecter) Link(oldpath interface{}, newpath interface{}) *mockUnixProvider_Link_Call {
	return &mockUnixProvider_Link_Call{Call: _e.mock.On("Link", oldpath, newpath)}
}

func (_c *mockUnixProvider_Link_Call) Run(run func(oldpath string, newpath string)) *mockUnixProvider_Link_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string))
	})
	return _c
}

func (_c *mockUnixProvider_Link_Call) Return(_a0 error) *mockUnixProvider_Link_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockUnixProvider_Link_Call) RunAndReturn(run func(string, string) error) *mockUnixProvider_Link_Call {
	_c.Call.Return(run)
	return _c
}

// Mkdir provides a mock function with given fields: path, mode
func (_m *mockUnixProvider) Mkdir(path string, mode uint32) error {
	ret := _m.Called(path, mode)

	if len(ret) == 0 {
		panic("no return value specified for Mkdir")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, uint32) error); ok {
		r0 = rf(path, mode)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockUnixProvider_Mkdir_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Mkdir'
type mockUnixProvider_Mkdir_Call struct {
	*mock.Call
}

// Mkdir is a helper method to define mock.On call
//   - path string
//   - mode uint32
func (_e *mockUnixProvider_Expecter) Mkdir(path interface{}, mode interface{}) *mockUnixProvider_Mkdir_Call {
	return &mockUnixProvider_Mkdir_Call{Call: _e.mock.On("Mkdir", path, mode)}
}

func (_c *mockUnixProvider_Mkdir_Call) Run(run func(path string, mode uint32)) *mockUnixProvider_Mkdir_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(uint32))
	})
	return _c
}

func (_c *mockUnixProvider_Mkdir_Call) Return(_a0 error) *mockUnixProvider_Mkdir_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockUnixProvider_Mkdir_Call) RunAndReturn(run func(string, uint32) error) *mockUnixProvider_Mkdir_Call {
	_c.Call.Return(run)
	return _c
}

// Symlink provides a mock function with given fields: oldpath, newpath
func (_m *mockUnixProvider) Symlink(oldpath string, newpath string) error {
	ret := _m.Called(oldpath, newpath)

	if len(ret) == 0 {
		panic("no return value specified for Symlink")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(oldpath, newpath)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockUnixProvider_Symlink_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'Symlink'
type mockUnixProvider_Symlink_Call struct {
	*mock.Call
}

// Symlink is a helper method to define mock.On call
//   - oldpath string
//   - newpath string
func (_e *mockUnixProvider_Expecter) Symlink(oldpath interface{}, newpath interface{}) *mockUnixProvider_Symlink_Call {
	return &mockUnixProvider_Symlink_Call{Call: _e.mock.On("Symlink", oldpath, newpath)}
}

func (_c *mockUnixProvider_Symlink_Call) Run(run func(oldpath string, newpath string)) *mockUnixProvider_Symlink_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].(string))
	})
	return _c
}

func (_c *mockUnixProvider_Symlink_Call) Return(_a0 error) *mockUnixProvider_Symlink_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockUnixProvider_Symlink_Call) RunAndReturn(run func(string, string) error) *mockUnixProvider_Symlink_Call {
	_c.Call.Return(run)
	return _c
}

// UtimesNano provides a mock function with given fields: path, times
func (_m *mockUnixProvider) UtimesNano(path string, times []unix.Timespec) error {
	ret := _m.Called(path, times)

	if len(ret) == 0 {
		panic("no return value specified for UtimesNano")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(string, []unix.Timespec) error); ok {
		r0 = rf(path, times)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// mockUnixProvider_UtimesNano_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'UtimesNano'
type mockUnixProvider_UtimesNano_Call struct {
	*mock.Call
}

// UtimesNano is a helper method to define mock.On call
//   - path string
//   - times []unix.Timespec
func (_e *mockUnixProvider_Expecter) UtimesNano(path interface{}, times interface{}) *mockUnixProvider_UtimesNano_Call {
	return &mockUnixProvider_UtimesNano_Call{Call: _e.mock.On("UtimesNano", path, times)}
}

func (_c *mockUnixProvider_UtimesNano_Call) Run(run func(path string, times []unix.Timespec)) *mockUnixProvider_UtimesNano_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(string), args[1].([]unix.Timespec))
	})
	return _c
}

func (_c *mockUnixProvider_UtimesNano_Call) Return(_a0 error) *mockUnixProvider_UtimesNano_Call {
	_c.Call.Return(_a0)
	return _c
}

func (_c *mockUnixProvider_UtimesNano_Call) RunAndReturn(run func(string, []unix.Timespec) error) *mockUnixProvider_UtimesNano_Call {
	_c.Call.Return(run)
	return _c
}

// newMockUnixProvider creates a new instance of mockUnixProvider. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func newMockUnixProvider(t interface {
	mock.TestingT
	Cleanup(func())
}) *mockUnixProvider {
	mock := &mockUnixProvider{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
