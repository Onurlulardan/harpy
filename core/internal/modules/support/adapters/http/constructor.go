package supporthttp

import (
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strings"
)

type BookingConfig struct {
	URL, Provider string
}

type Dependencies struct {
	Access       AccessVerifier
	Applications Applications
	RateKeys     RateKeyResolver
	Booking      BookingConfig
	Report       ErrorSink
	Correlation  CorrelationResolver
}

type Handler struct {
	access       AccessVerifier
	applications Applications
	rateKeys     RateKeyResolver
	booking      BookingConfig
	report       ErrorSink
	correlation  CorrelationResolver
}

func NewHandler(dependencies Dependencies) (*Handler, error) {
	if nilPort(dependencies.Access) || !dependencies.Applications.ready() ||
		nilPort(dependencies.RateKeys) || dependencies.Report == nil ||
		dependencies.Correlation == nil || !validBooking(dependencies.Booking) {
		return nil, fmt.Errorf("support HTTP dependencies are required")
	}
	return &Handler{access: dependencies.Access, applications: dependencies.Applications,
		rateKeys: dependencies.RateKeys, booking: dependencies.Booking,
		report: dependencies.Report, correlation: dependencies.Correlation}, nil
}

func (handler *Handler) Ready() bool {
	return handler != nil && !nilPort(handler.access) && handler.applications.ready() &&
		!nilPort(handler.rateKeys) && handler.report != nil && handler.correlation != nil &&
		validBooking(handler.booking)
}

func validBooking(config BookingConfig) bool {
	config.URL, config.Provider = strings.TrimSpace(config.URL), strings.TrimSpace(config.Provider)
	if config.URL == "" {
		return config.Provider == ""
	}
	parsed, err := url.Parse(config.URL)
	return err == nil && parsed.Scheme == "https" && parsed.Host != "" &&
		len(config.URL) <= 2048 && len(config.Provider) >= 1 && len(config.Provider) <= 50
}

func nilPort(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}

type Route struct {
	Method, Pattern string
	Handler         http.Handler
}
