package audithttp

import (
	"fmt"
	"reflect"
)

type Dependencies struct {
	Access      AccessVerifier
	Application Application
	Report      ErrorSink
	Correlation CorrelationResolver
}

type Handler struct{ dependencies Dependencies }

func NewHandler(dependencies Dependencies) (*Handler, error) {
	if nilPort(dependencies.Access) || nilPort(dependencies.Application) ||
		dependencies.Report == nil || dependencies.Correlation == nil {
		return nil, fmt.Errorf("audit HTTP dependencies are invalid")
	}
	return &Handler{dependencies: dependencies}, nil
}

func (handler *Handler) Ready() bool {
	return handler != nil && !nilPort(handler.dependencies.Access) &&
		!nilPort(handler.dependencies.Application) && handler.dependencies.Report != nil &&
		handler.dependencies.Correlation != nil
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
