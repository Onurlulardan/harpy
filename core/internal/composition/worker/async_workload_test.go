package worker

import (
	"context"
	"errors"
	"testing"

	brokerrabbit "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/rabbitmq"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/outboxdelivery"
)

type probeConnector struct {
	broker outboxdelivery.Broker
	err    error
}

func (connector probeConnector) Open(context.Context) (outboxdelivery.Broker, error) {
	return connector.broker, connector.err
}

type probeBroker struct{ closeErr error }

func (probeBroker) Publish(
	context.Context,
	brokerrabbit.PublishRequest,
) (brokerrabbit.PublishReceipt, error) {
	return brokerrabbit.PublishReceipt{}, errors.New("unexpected publish")
}

func (broker probeBroker) Close() error { return broker.closeErr }

func TestRabbitMQPreflightRequiresOpenAndCleanClose(t *testing.T) {
	if err := probeRabbitMQ(
		context.Background(), probeConnector{err: errors.New("offline")},
	); err == nil {
		t.Fatal("offline broker passed preflight")
	}
	if err := probeRabbitMQ(
		context.Background(), probeConnector{broker: probeBroker{closeErr: errors.New("close")}},
	); err == nil {
		t.Fatal("unclean preflight session passed")
	}
	if err := probeRabbitMQ(
		context.Background(), probeConnector{broker: probeBroker{}},
	); err != nil {
		t.Fatal(err)
	}
}
