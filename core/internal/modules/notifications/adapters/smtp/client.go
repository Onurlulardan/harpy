package smtp

import (
	"errors"
	"reflect"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

var ErrConfiguration = errors.New("notification SMTP adapter is not configured")

type Client struct {
	exchanger exchanger
	now       func() time.Time
}

func NewClient(exchange exchanger, now func() time.Time) (*Client, error) {
	if missing(exchange) || now == nil {
		return nil, ErrConfiguration
	}
	return &Client{exchanger: exchange, now: now}, nil
}

func NewProductionClient() (*Client, error) {
	return NewClient(newNetworkExchanger(), time.Now)
}

func missing(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	return reflected.Kind() == reflect.Pointer && reflected.IsNil()
}

var _ ports.ProviderSender = (*Client)(nil)
