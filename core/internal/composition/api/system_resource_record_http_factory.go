package api

import (
	"fmt"
	"reflect"
)

type systemResourceRecordHTTPErrorSink func(error)

type systemResourceRecordHTTPDependencies struct {
	Access      systemResourceRecordHTTPAccessVerifier
	Application systemResourceRecordHTTPApplication
	Report      systemResourceRecordHTTPErrorSink
}

type systemResourceRecordHTTPHandler struct {
	access      systemResourceRecordHTTPAccessVerifier
	application systemResourceRecordHTTPApplication
	report      systemResourceRecordHTTPErrorSink
}

func newSystemResourceRecordHTTPHandler(
	dependencies systemResourceRecordHTTPDependencies,
) (*systemResourceRecordHTTPHandler, error) {
	if nilSystemResourceRecordHTTPDependency(dependencies.Access) ||
		nilSystemResourceRecordHTTPDependency(dependencies.Application) ||
		dependencies.Report == nil {
		return nil, fmt.Errorf("system resource record HTTP dependencies are required")
	}
	return &systemResourceRecordHTTPHandler{
		access: dependencies.Access, application: dependencies.Application,
		report: dependencies.Report,
	}, nil
}

func nilSystemResourceRecordHTTPDependency(value any) bool {
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
