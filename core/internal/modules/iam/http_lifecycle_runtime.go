package iam

import (
	"fmt"
	"reflect"
	"time"
)

type LifecycleHTTP struct {
	security    *HTTPSecurity
	application LifecycleHTTPApplication
	report      HTTPErrorSink
	clock       func() time.Time
}

func (security *HTTPSecurity) NewLifecycleHTTP(
	application LifecycleHTTPApplication, report HTTPErrorSink,
) (*LifecycleHTTP, error) {
	return newLifecycleHTTP(security, application, report, time.Now)
}

func newLifecycleHTTP(security *HTTPSecurity, application LifecycleHTTPApplication,
	report HTTPErrorSink, clock func() time.Time) (*LifecycleHTTP, error) {
	if security == nil || !security.Valid() || nilLifecycleHTTP(application) ||
		report == nil || clock == nil {
		return nil, fmt.Errorf("IAM lifecycle HTTP dependencies are required")
	}
	return &LifecycleHTTP{security: security, application: application,
		report: report, clock: clock}, nil
}

func nilLifecycleHTTP(value LifecycleHTTPApplication) bool {
	if value == nil {
		return true
	}
	seen := reflect.ValueOf(value)
	return seen.Kind() == reflect.Pointer && seen.IsNil()
}
