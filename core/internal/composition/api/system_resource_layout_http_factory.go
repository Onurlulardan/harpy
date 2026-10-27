package api

import (
	"fmt"
	"reflect"
)

type systemResourceLayoutHTTPErrorSink func(error)

type systemResourceLayoutHTTPDependencies struct {
	Access      systemResourceLayoutHTTPAccessVerifier
	Application systemResourceLayoutHTTPApplication
	Report      systemResourceLayoutHTTPErrorSink
}

type systemResourceLayoutHTTPHandler struct {
	access      systemResourceLayoutHTTPAccessVerifier
	application systemResourceLayoutHTTPApplication
	report      systemResourceLayoutHTTPErrorSink
}

func newSystemResourceLayoutHTTPHandler(
	dependencies systemResourceLayoutHTTPDependencies,
) (*systemResourceLayoutHTTPHandler, error) {
	if nilSystemResourceLayoutHTTPDependency(dependencies.Access) ||
		nilSystemResourceLayoutHTTPDependency(dependencies.Application) || dependencies.Report == nil {
		return nil, fmt.Errorf("system resource layout HTTP dependencies are required")
	}
	return &systemResourceLayoutHTTPHandler{
		access: dependencies.Access, application: dependencies.Application,
		report: dependencies.Report,
	}, nil
}

func nilSystemResourceLayoutHTTPDependency(value any) bool {
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
