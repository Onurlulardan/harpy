package iam

import (
	"fmt"
	"reflect"
	"time"
)

type PermissionRequestHTTP struct {
	security    *HTTPSecurity
	application PermissionRequestHTTPApplication
	report      HTTPErrorSink
	clock       func() time.Time
}

func (security *HTTPSecurity) NewPermissionRequestHTTP(
	application PermissionRequestHTTPApplication,
	report HTTPErrorSink,
) (*PermissionRequestHTTP, error) {
	return newPermissionRequestHTTP(security, application, report, time.Now)
}

func newPermissionRequestHTTP(
	security *HTTPSecurity,
	application PermissionRequestHTTPApplication,
	report HTTPErrorSink,
	clock func() time.Time,
) (*PermissionRequestHTTP, error) {
	if security == nil || !security.Valid() || nilPermissionRequestHTTP(application) ||
		report == nil || clock == nil {
		return nil, fmt.Errorf("IAM permission request HTTP dependencies are required")
	}
	return &PermissionRequestHTTP{
		security: security, application: application, report: report, clock: clock,
	}, nil
}

func nilPermissionRequestHTTP(value PermissionRequestHTTPApplication) bool {
	if value == nil {
		return true
	}
	seen := reflect.ValueOf(value)
	return seen.Kind() == reflect.Pointer && seen.IsNil()
}
