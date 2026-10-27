package notificationhttp

import (
	"fmt"
	"net/http"
	"reflect"
)

type ErrorSink func(error)
type CorrelationResolver func(*http.Request) string

type HandlerDependencies struct {
	Access      AccessVerifier
	Application Application
	Realtime    RealtimeTransport
	Report      ErrorSink
	Correlation CorrelationResolver
}

type Handler struct {
	access      AccessVerifier
	application Application
	realtime    RealtimeTransport
	report      ErrorSink
	correlation CorrelationResolver
}

func NewHandler(dependencies HandlerDependencies) (*Handler, error) {
	if nilValue(dependencies.Access) || nilValue(dependencies.Application) ||
		nilValue(dependencies.Realtime) || dependencies.Report == nil ||
		dependencies.Correlation == nil {
		return nil, fmt.Errorf("notification HTTP dependencies are required")
	}
	return &Handler{access: dependencies.Access, application: dependencies.Application,
		realtime: dependencies.Realtime, report: dependencies.Report,
		correlation: dependencies.Correlation}, nil
}

func (handler *Handler) Ready() bool {
	return handler != nil && !nilValue(handler.access) && !nilValue(handler.application) &&
		!nilValue(handler.realtime) && handler.report != nil && handler.correlation != nil
}

func nilValue(value any) bool {
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
