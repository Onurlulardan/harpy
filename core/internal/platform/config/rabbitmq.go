package config

import "time"

const (
	rabbitMQHeartbeat        = 30 * time.Second
	rabbitMQDialTimeout      = 5 * time.Second
	rabbitMQConfirmTimeout   = 10 * time.Second
	rabbitMQReconnectMinimum = time.Second
	rabbitMQReconnectMaximum = 30 * time.Second
)

type RabbitMQ struct {
	url Secret
}

func loadRabbitMQ(values environmentValues, environment Environment) (RabbitMQ, error) {
	url, err := loadSecret(values, "HARPY_RABBITMQ_URL", true, secretFileLimit)
	if err != nil {
		return RabbitMQ{}, err
	}
	if err := validateRabbitMQURL("HARPY_RABBITMQ_URL", url, environment); err != nil {
		return RabbitMQ{}, err
	}
	return RabbitMQ{url: url}, nil
}

func (value RabbitMQ) URL() Secret               { return value.url }
func (RabbitMQ) Heartbeat() time.Duration        { return rabbitMQHeartbeat }
func (RabbitMQ) DialTimeout() time.Duration      { return rabbitMQDialTimeout }
func (RabbitMQ) ConfirmTimeout() time.Duration   { return rabbitMQConfirmTimeout }
func (RabbitMQ) ReconnectMinimum() time.Duration { return rabbitMQReconnectMinimum }
func (RabbitMQ) ReconnectMaximum() time.Duration { return rabbitMQReconnectMaximum }
