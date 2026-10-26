package iam

import (
	"fmt"
	"reflect"
	"time"
)

type IdentityHTTP struct {
	security    *HTTPSecurity
	application IdentityHTTPApplication
	report      HTTPErrorSink
	clock       func() time.Time
}

func (security *HTTPSecurity) NewIdentityHTTP(
	application IdentityHTTPApplication,
	report HTTPErrorSink,
) (*IdentityHTTP, error) {
	return newIdentityHTTP(security, application, report, time.Now)
}

func newIdentityHTTP(
	security *HTTPSecurity,
	application IdentityHTTPApplication,
	report HTTPErrorSink,
	clock func() time.Time,
) (*IdentityHTTP, error) {
	if security == nil || !security.Valid() || nilIdentityHTTP(application) ||
		report == nil || clock == nil {
		return nil, fmt.Errorf("IAM system-resource HTTP dependencies are required")
	}
	return &IdentityHTTP{
		security: security, application: application,
		report: report, clock: clock,
	}, nil
}

func nilIdentityHTTP(value IdentityHTTPApplication) bool {
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
