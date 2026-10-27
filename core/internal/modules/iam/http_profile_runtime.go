package iam

import (
	"fmt"
	"reflect"
	"time"
)

type ProfileHTTP struct {
	security    *HTTPSecurity
	application ProfileHTTPApplication
	report      HTTPErrorSink
	clock       func() time.Time
}

func (security *HTTPSecurity) NewProfileHTTP(
	application ProfileHTTPApplication,
	report HTTPErrorSink,
) (*ProfileHTTP, error) {
	return newProfileHTTP(security, application, report, time.Now)
}

func newProfileHTTP(
	security *HTTPSecurity,
	application ProfileHTTPApplication,
	report HTTPErrorSink,
	clock func() time.Time,
) (*ProfileHTTP, error) {
	if security == nil || !security.Valid() || nilProfileHTTP(application) ||
		report == nil || clock == nil {
		return nil, fmt.Errorf("IAM profile HTTP dependencies are required")
	}
	return &ProfileHTTP{
		security: security, application: application, report: report, clock: clock,
	}, nil
}

func nilProfileHTTP(value ProfileHTTPApplication) bool {
	if value == nil {
		return true
	}
	seen := reflect.ValueOf(value)
	return seen.Kind() == reflect.Pointer && seen.IsNil()
}
