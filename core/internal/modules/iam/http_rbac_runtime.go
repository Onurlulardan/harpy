package iam

import (
	"fmt"
	"reflect"
	"time"
)

type RBACHTTP struct {
	security    *HTTPSecurity
	application RBACHTTPApplication
	report      HTTPErrorSink
	clock       func() time.Time
}

func (security *HTTPSecurity) NewRBACHTTP(
	application RBACHTTPApplication,
	report HTTPErrorSink,
) (*RBACHTTP, error) {
	return newRBACHTTP(security, application, report, time.Now)
}

func newRBACHTTP(
	security *HTTPSecurity,
	application RBACHTTPApplication,
	report HTTPErrorSink,
	clock func() time.Time,
) (*RBACHTTP, error) {
	if security == nil || !security.Valid() || nilRBACHTTP(application) ||
		report == nil || clock == nil {
		return nil, fmt.Errorf("IAM RBAC HTTP dependencies are required")
	}
	return &RBACHTTP{
		security: security, application: application, report: report, clock: clock,
	}, nil
}

func nilRBACHTTP(value RBACHTTPApplication) bool {
	if value == nil {
		return true
	}
	seen := reflect.ValueOf(value)
	return seen.Kind() == reflect.Pointer && seen.IsNil()
}
