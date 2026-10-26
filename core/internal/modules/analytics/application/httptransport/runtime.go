package httptransport

import (
	"fmt"
	"net/http"
	"reflect"
)

type ErrorSink func(error)
type CorrelationResolver func(*http.Request) string
type TrustedClientIPResolver func(*http.Request) (string, error)
type FailureResolver func(error) (int, string, string)

type Runtime struct {
	access      AccessVerifier
	report      ErrorSink
	correlation CorrelationResolver
	clientIP    TrustedClientIPResolver
	failure     FailureResolver
}

func NewRuntime(
	access AccessVerifier, report ErrorSink, correlation CorrelationResolver,
	clientIP TrustedClientIPResolver, failure FailureResolver,
) (*Runtime, error) {
	if PortMissing(access) || report == nil || correlation == nil || clientIP == nil || failure == nil {
		return nil, fmt.Errorf("analytics HTTP runtime dependencies are required")
	}
	return &Runtime{access: access, report: report, correlation: correlation,
		clientIP: clientIP, failure: failure}, nil
}

func (runtime *Runtime) AccessFor(
	request *http.Request, permission Permission, operation string,
) (Access, error) {
	if runtime == nil || PortMissing(runtime.access) || request == nil {
		return Access{}, ErrAuthentication
	}
	access, err := runtime.access.Verify(request.Context(), AccessRequest{
		AuthorizationValues: append([]string(nil), request.Header.Values("Authorization")...),
		TenantIDValues:      append([]string(nil), request.Header.Values("X-Tenant-ID")...),
		TenantSlugValues:    append([]string(nil), request.Header.Values("X-Tenant-Slug")...),
		Permission:          permission, Operation: operation})
	if err != nil {
		return Access{}, err
	}
	if !access.ValidFor(permission, operation) {
		return Access{}, ErrForbidden
	}
	return access, nil
}

func (runtime *Runtime) ClientAddress(request *http.Request) (string, error) {
	if runtime == nil || runtime.clientIP == nil {
		return "", ErrUnavailable
	}
	return runtime.clientIP(request)
}

func (runtime *Runtime) Correlation(request *http.Request) string {
	if runtime == nil || runtime.correlation == nil {
		return ""
	}
	return runtime.correlation(request)
}

func (runtime *Runtime) Report(err error) {
	if runtime != nil && runtime.report != nil && err != nil {
		runtime.report(err)
	}
}

func PortMissing(value any) bool {
	if value == nil {
		return true
	}
	seen := reflect.ValueOf(value)
	switch seen.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return seen.IsNil()
	default:
		return false
	}
}

type Route struct {
	Method, Pattern string
	Handler         http.Handler
}
