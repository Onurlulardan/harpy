package iam

import (
	"fmt"
	"reflect"
	"time"
)

type PortalUserHTTP struct {
	security    *HTTPSecurity
	application PortalUserHTTPApplication
	report      HTTPErrorSink
	clock       func() time.Time
}

func (security *HTTPSecurity) NewPortalUserHTTP(
	application PortalUserHTTPApplication, report HTTPErrorSink,
) (*PortalUserHTTP, error) {
	if security == nil || !security.Valid() || nilPortalUserApplication(application) ||
		report == nil {
		return nil, fmt.Errorf("IAM portal user HTTP dependencies are required")
	}
	return &PortalUserHTTP{security: security, application: application,
		report: report, clock: time.Now}, nil
}

func nilPortalUserApplication(value PortalUserHTTPApplication) bool {
	if value == nil {
		return true
	}
	seen := reflect.ValueOf(value)
	return seen.Kind() == reflect.Pointer && seen.IsNil()
}
