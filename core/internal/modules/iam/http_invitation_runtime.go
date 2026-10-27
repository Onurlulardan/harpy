package iam

import (
	"fmt"
	"reflect"
	"time"
)

type InvitationHTTP struct {
	security    *HTTPSecurity
	application InvitationHTTPApplication
	report      HTTPErrorSink
	clock       func() time.Time
}

func (security *HTTPSecurity) NewInvitationHTTP(
	application InvitationHTTPApplication,
	report HTTPErrorSink,
) (*InvitationHTTP, error) {
	return newInvitationHTTP(security, application, report, time.Now)
}

func newInvitationHTTP(
	security *HTTPSecurity,
	application InvitationHTTPApplication,
	report HTTPErrorSink,
	clock func() time.Time,
) (*InvitationHTTP, error) {
	if security == nil || !security.Valid() || nilInvitationHTTP(application) ||
		report == nil || clock == nil {
		return nil, fmt.Errorf("IAM invitation HTTP dependencies are required")
	}
	return &InvitationHTTP{
		security: security, application: application, report: report, clock: clock,
	}, nil
}

func nilInvitationHTTP(value InvitationHTTPApplication) bool {
	if value == nil {
		return true
	}
	seen := reflect.ValueOf(value)
	return seen.Kind() == reflect.Pointer && seen.IsNil()
}
