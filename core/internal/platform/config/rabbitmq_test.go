package config

import (
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestWorkerRabbitMQConfigIsRequiredAndProfileScoped(t *testing.T) {
	if _, err := loadWorker(validEnvironment()); err == nil {
		t.Fatal("worker started without RabbitMQ authority")
	}
	if _, err := loadAPI(append(validAPIEnvironment(),
		"HARPY_RABBITMQ_URL=amqp://worker:secret@rabbit.invalid/%2f",
	)); err == nil || !strings.Contains(err.Error(), "unknown key") {
		t.Fatalf("API accepted worker broker credentials: %v", err)
	}
}

func TestRabbitMQURLRequiresTLSAndDoesNotLeakCredentials(t *testing.T) {
	tests := []struct {
		name        string
		environment string
		value       string
	}{
		{"production plaintext", "production", "amqp://worker:must-not-leak@rabbit.invalid/%2f"},
		{"missing password", "development", "amqp://worker@rabbit.invalid/%2f"},
		{"multiple hosts", "development", "amqp://worker:secret@one.invalid,two.invalid/%2f"},
		{"query override", "development", "amqp://worker:secret@rabbit.invalid/%2f?heartbeat=0"},
		{"missing vhost", "development", "amqp://worker:secret@rabbit.invalid"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entries := withEnvironment(
				validWorkerEnvironment(),
				"HARPY_ENVIRONMENT="+test.environment,
				"HARPY_RABBITMQ_URL="+test.value,
				"HARPY_OTLP_ENDPOINT=https://otel.invalid",
				"HARPY_OTLP_INSECURE=false",
			)
			_, err := loadWorker(entries)
			if err == nil || strings.Contains(err.Error(), "must-not-leak") {
				t.Fatalf("invalid broker URL result: %v", err)
			}
		})
	}
	secret := Secret{value: "amqps://worker:must-not-leak@rabbit.invalid/%2f"}
	if text := fmt.Sprintf("%v %#v", secret, secret); strings.Contains(text, "must-not-leak") {
		t.Fatal("RabbitMQ secret leaked through formatting")
	}
}

func TestRabbitMQReliabilityTimingIsExact(t *testing.T) {
	worker, err := loadWorker(validWorkerEnvironment())
	if err != nil {
		t.Fatal(err)
	}
	settings := worker.RabbitMQ()
	if settings.Heartbeat() != 30*time.Second || settings.DialTimeout() != 5*time.Second ||
		settings.ConfirmTimeout() != 10*time.Second || settings.ReconnectMinimum() != time.Second ||
		settings.ReconnectMaximum() != 30*time.Second {
		t.Fatal("RabbitMQ reliability timing drifted")
	}
}
