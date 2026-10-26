package cataloghttp

import (
	"fmt"
	"net/http"
	"reflect"
)

type ErrorSink func(error)
type CorrelationResolver func(*http.Request) string

type Runtime struct {
	access      AccessVerifier
	report      ErrorSink
	correlation CorrelationResolver
}

func NewRuntime(access AccessVerifier, report ErrorSink,
	correlation CorrelationResolver,
) (*Runtime, error) {
	if nilPort(access) || report == nil || correlation == nil {
		return nil, fmt.Errorf("catalog HTTP dependencies are required")
	}
	return &Runtime{access: access, report: report, correlation: correlation}, nil
}

func (runtime *Runtime) Ready() bool {
	return runtime != nil && !nilPort(runtime.access) && runtime.report != nil &&
		runtime.correlation != nil
}

func (runtime *Runtime) AccessFor(request *http.Request, permission Permission,
	operation, purposeTarget string,
) (Access, error) {
	if !runtime.Ready() || request == nil {
		return Access{}, ErrAuthentication
	}
	access, err := runtime.access.Verify(request.Context(), AccessRequest{
		AuthorizationValues: append([]string(nil), request.Header.Values("Authorization")...),
		TenantIDValues:      append([]string(nil), request.Header.Values("X-Tenant-ID")...),
		TenantSlugValues:    append([]string(nil), request.Header.Values("X-Tenant-Slug")...),
		Permission:          permission, Operation: operation, PurposeTarget: purposeTarget,
	})
	if err != nil {
		return Access{}, err
	}
	if !access.ValidFor(permission, operation) {
		return Access{}, ErrForbidden
	}
	return access, nil
}

func nilPort(value any) bool {
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

func PortReady(value any) bool { return !nilPort(value) }
