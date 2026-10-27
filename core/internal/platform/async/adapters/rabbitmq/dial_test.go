package rabbitmq

import (
	"context"
	"crypto/tls"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/configfixture"
)

func TestDialConfigCarriesExactHeartbeatTLSAndRedactedFailure(t *testing.T) {
	settings := workerRabbitSettings(t)
	connection := buildConnectionConfig(context.Background(), settings)
	if connection.Heartbeat != settings.Heartbeat() || connection.Locale != "en_US" ||
		connection.TLSClientConfig == nil ||
		connection.TLSClientConfig.MinVersion != tls.VersionTLS12 ||
		connection.Properties["connection_name"] != "harpy-core-worker" ||
		connection.Dial == nil {
		t.Fatal("RabbitMQ connection reliability config drifted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := Dial(ctx, settings)
	if err == nil || strings.Contains(err.Error(), "integration-secret") {
		t.Fatalf("RabbitMQ dial did not fail safely: %v", err)
	}
}

func workerRabbitSettings(t *testing.T) config.RabbitMQ {
	t.Helper()
	t.Setenv("HARPY_ENVIRONMENT", "test")
	t.Setenv("HARPY_POSTGRES_ENDPOINT_KIND", "direct_postgres")
	t.Setenv(
		"HARPY_POSTGRES_DSN",
		"postgres://harpy_v2_runtime@127.0.0.1:1/harpy?sslmode=disable",
	)
	t.Setenv("HARPY_RABBITMQ_URL", "amqp://worker:integration-secret@127.0.0.1:1/%2f")
	configfixture.SetWorker(t)
	settings, err := config.LoadWorker()
	if err != nil {
		t.Fatal(err)
	}
	return settings.RabbitMQ()
}
