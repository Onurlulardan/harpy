package iam

import (
	"fmt"
	"reflect"
	"time"
)

type AuthHTTP struct {
	security    *HTTPSecurity
	application AuthHTTPApplication
	report      HTTPErrorSink
	clock       func() time.Time
}

func (security *HTTPSecurity) NewAuthHTTP(
	application AuthHTTPApplication,
	report HTTPErrorSink,
) (*AuthHTTP, error) {
	return newAuthHTTP(security, application, report, time.Now)
}

func newAuthHTTP(
	security *HTTPSecurity,
	application AuthHTTPApplication,
	report HTTPErrorSink,
	clock func() time.Time,
) (*AuthHTTP, error) {
	if security == nil || !security.Valid() || nilAuthHTTP(application) ||
		report == nil || clock == nil {
		return nil, fmt.Errorf("IAM auth HTTP dependencies are required")
	}
	return &AuthHTTP{
		security: security, application: application, report: report, clock: clock,
	}, nil
}

func nilAuthHTTP(value AuthHTTPApplication) bool {
	if value == nil {
		return true
	}
	seen := reflect.ValueOf(value)
	switch seen.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Pointer, reflect.Slice:
		return seen.IsNil()
	default:
		return false
	}
}
