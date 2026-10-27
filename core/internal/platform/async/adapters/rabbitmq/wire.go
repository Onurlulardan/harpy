package rabbitmq

import (
	"context"

	amqp "github.com/rabbitmq/amqp091-go"
)

type confirmation interface {
	WaitContext(context.Context) (bool, error)
}

type publishWire interface {
	Publish(context.Context, string, string, bool, bool, amqp.Publishing) (confirmation, error)
	Returns() <-chan amqp.Return
	Closed() <-chan *amqp.Error
	Close() error
}

type amqpWire struct {
	channel *amqp.Channel
	returns chan amqp.Return
	closed  chan *amqp.Error
}

func openWire(connection *amqp.Connection) (*amqpWire, error) {
	if connection == nil || connection.IsClosed() {
		return nil, newFailure(FailureClosed, "00000000-0000-0000-0000-000000000000", 0)
	}
	channel, err := connection.Channel()
	if err != nil {
		return nil, err
	}
	if err := declarePublisherExchange(channel); err != nil {
		_ = channel.Close()
		return nil, err
	}
	if err := channel.Confirm(false); err != nil {
		_ = channel.Close()
		return nil, err
	}
	return &amqpWire{
		channel: channel,
		returns: channel.NotifyReturn(make(chan amqp.Return, 1)),
		closed:  channel.NotifyClose(make(chan *amqp.Error, 1)),
	}, nil
}

func (wire *amqpWire) Publish(
	ctx context.Context, exchange string, routingKey string,
	mandatory bool, immediate bool, message amqp.Publishing,
) (confirmation, error) {
	return wire.channel.PublishWithDeferredConfirmWithContext(
		ctx, exchange, routingKey, mandatory, immediate, message,
	)
}

func (wire *amqpWire) Returns() <-chan amqp.Return { return wire.returns }
func (wire *amqpWire) Closed() <-chan *amqp.Error  { return wire.closed }
func (wire *amqpWire) Close() error                { return wire.channel.Close() }
