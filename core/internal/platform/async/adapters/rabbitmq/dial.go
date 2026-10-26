package rabbitmq

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	amqp "github.com/rabbitmq/amqp091-go"
)

func Dial(ctx context.Context, settings config.RabbitMQ) (*amqp.Connection, error) {
	if ctx == nil || settings.URL().Value() == "" {
		return nil, fmt.Errorf("RabbitMQ dial settings are incomplete")
	}
	connection, err := amqp.DialConfig(settings.URL().Value(), buildConnectionConfig(ctx, settings))
	if err != nil {
		return nil, fmt.Errorf("RabbitMQ dial failed")
	}
	return connection, nil
}

func buildConnectionConfig(ctx context.Context, settings config.RabbitMQ) amqp.Config {
	properties := amqp.NewConnectionProperties()
	properties["connection_name"] = "harpy-core-worker"
	properties["product"] = "Harpy"
	dialer := net.Dialer{
		Timeout:   settings.DialTimeout(),
		KeepAlive: settings.Heartbeat(),
	}
	return amqp.Config{
		Heartbeat:       settings.Heartbeat(),
		Locale:          "en_US",
		Properties:      properties,
		TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
		Dial: func(network string, address string) (net.Conn, error) {
			return dialer.DialContext(ctx, network, address)
		},
	}
}
