package documentshttp

import (
	"fmt"
	"net/http"
	"reflect"
)

type ErrorSink func(error)
type CorrelationResolver func(*http.Request) string

type Dependencies struct {
	Access      AccessVerifier
	Application Application
	Report      ErrorSink
	Correlation CorrelationResolver
}

type Handler struct {
	access      AccessVerifier
	application Application
	report      ErrorSink
	correlation CorrelationResolver
}

func NewHandler(dependencies Dependencies) (*Handler, error) {
	if httpNil(dependencies.Access) || httpNil(dependencies.Application) ||
		dependencies.Report == nil || dependencies.Correlation == nil {
		return nil, fmt.Errorf("Documents HTTP dependencies are required")
	}
	return &Handler{access: dependencies.Access, application: dependencies.Application,
		report: dependencies.Report, correlation: dependencies.Correlation}, nil
}

func (handler *Handler) Ready() bool {
	return handler != nil && !httpNil(handler.access) && !httpNil(handler.application) &&
		handler.report != nil && handler.correlation != nil
}

func httpNil(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}
