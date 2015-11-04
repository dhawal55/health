package mocks

import "github.com/stretchr/testify/mock"

type HealthChecker struct {
	mock.Mock
}

func (_m *HealthChecker) IsHealthy() (bool, error, []string) {
	ret := _m.Called()

	var r0 bool
	if rf, ok := ret.Get(0).(func() bool); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	var r2 []string
	if rf, ok := ret.Get(2).(func() []string); ok {
		r2 = rf()
	} else {
		if ret.Get(2) != nil {
			r2 = ret.Get(2).([]string)
		}
	}

	return r0, r1, r2
}
