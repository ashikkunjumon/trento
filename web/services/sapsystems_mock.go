// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package services

import (
	mock "github.com/stretchr/testify/mock"
	sapsystem "github.com/trento-project/trento/internal/sapsystem"
)

// MockSAPSystemsService is an autogenerated mock type for the SAPSystemsService type
type MockSAPSystemsService struct {
	mock.Mock
}

// GetAttachedDatabasesById provides a mock function with given fields: id
func (_m *MockSAPSystemsService) GetAttachedDatabasesById(id string) (sapsystem.SAPSystemsList, error) {
	ret := _m.Called(id)

	var r0 sapsystem.SAPSystemsList
	if rf, ok := ret.Get(0).(func(string) sapsystem.SAPSystemsList); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(sapsystem.SAPSystemsList)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSAPSystems provides a mock function with given fields:
func (_m *MockSAPSystemsService) GetSAPSystems() (sapsystem.SAPSystemsList, error) {
	ret := _m.Called()

	var r0 sapsystem.SAPSystemsList
	if rf, ok := ret.Get(0).(func() sapsystem.SAPSystemsList); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(sapsystem.SAPSystemsList)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSAPSystemsById provides a mock function with given fields: id
func (_m *MockSAPSystemsService) GetSAPSystemsById(id string) (sapsystem.SAPSystemsList, error) {
	ret := _m.Called(id)

	var r0 sapsystem.SAPSystemsList
	if rf, ok := ret.Get(0).(func(string) sapsystem.SAPSystemsList); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(sapsystem.SAPSystemsList)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSAPSystemsByType provides a mock function with given fields: systemType
func (_m *MockSAPSystemsService) GetSAPSystemsByType(systemType int) (sapsystem.SAPSystemsList, error) {
	ret := _m.Called(systemType)

	var r0 sapsystem.SAPSystemsList
	if rf, ok := ret.Get(0).(func(int) sapsystem.SAPSystemsList); ok {
		r0 = rf(systemType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(sapsystem.SAPSystemsList)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(int) error); ok {
		r1 = rf(systemType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
