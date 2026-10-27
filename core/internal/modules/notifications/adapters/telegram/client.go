package telegram

import (
	"crypto/tls"
	"errors"
	"net/http"
	"reflect"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

var ErrConfiguration = errors.New("notification Telegram adapter is not configured")

type httpDoer interface {
	Do(*http.Request) (*http.Response, error)
}

type Client struct {
	http httpDoer
	now  func() time.Time
}

func NewClient(httpClient httpDoer, now func() time.Time) (*Client, error) {
	if missing(httpClient) || now == nil {
		return nil, ErrConfiguration
	}
	return &Client{http: httpClient, now: now}, nil
}

func NewProductionClient() (*Client, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{MinVersion: tls.VersionTLS12}
	client := &http.Client{Transport: transport, Timeout: 30 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return errors.New("Telegram redirects are disabled")
		}}
	return NewClient(client, time.Now)
}

func missing(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	return reflected.Kind() == reflect.Pointer && reflected.IsNil()
}

var _ ports.ProviderSender = (*Client)(nil)
