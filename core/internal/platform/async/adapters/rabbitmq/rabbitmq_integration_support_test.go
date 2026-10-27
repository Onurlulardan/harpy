//go:build integration

package rabbitmq

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/configfixture"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const rabbitImage = "rabbitmq:4.3.2-alpine@sha256:835cbc6fabceedef7a3b0ce3195a17adccb753a3ba645bdd61c5557b6e4bf580"

type rabbitFixture struct {
	container testcontainers.Container
	url       string
	settings  config.RabbitMQ
}

func startRabbitFixture(t *testing.T) rabbitFixture {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: rabbitImage,
			Env: map[string]string{
				"RABBITMQ_DEFAULT_USER":  "harpy_worker",
				"RABBITMQ_DEFAULT_PASS":  "integration-secret",
				"RABBITMQ_DEFAULT_VHOST": "harpy_integration",
			},
			ExposedPorts: []string{"5672/tcp"},
			WaitingFor:   wait.ForListeningPort("5672/tcp").WithStartupTimeout(2 * time.Minute),
		},
		Started: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), time.Minute)
		defer cleanupCancel()
		if err := container.Terminate(cleanupCtx); err != nil {
			t.Errorf("terminate RabbitMQ: %v", err)
		}
	})
	fixture := rabbitFixture{container: container}
	fixture.refresh(t)
	return fixture
}

func (fixture *rabbitFixture) refresh(t *testing.T) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	host, err := fixture.container.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}
	port, err := fixture.container.MappedPort(ctx, "5672/tcp")
	if err != nil {
		t.Fatal(err)
	}
	fixture.url = fmt.Sprintf(
		"amqp://harpy_worker:integration-secret@%s:%s/harpy_integration",
		host, port.Port(),
	)
	fixture.settings = loadRabbitSettings(t, fixture.url)
}

func loadRabbitSettings(t *testing.T, url string) config.RabbitMQ {
	t.Helper()
	t.Setenv("HARPY_ENVIRONMENT", "test")
	t.Setenv("HARPY_POSTGRES_ENDPOINT_KIND", "direct_postgres")
	t.Setenv(
		"HARPY_POSTGRES_DSN",
		"postgres://harpy_v2_runtime@127.0.0.1:1/harpy?sslmode=disable",
	)
	t.Setenv("HARPY_RABBITMQ_URL", url)
	configfixture.SetWorker(t)
	worker, err := config.LoadWorker()
	if err != nil {
		t.Fatal(err)
	}
	return worker.RabbitMQ()
}

func dialRabbit(t *testing.T, settings config.RabbitMQ) *amqp.Connection {
	t.Helper()
	deadline := time.Now().Add(90 * time.Second)
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		connection, err := Dial(ctx, settings)
		cancel()
		if err == nil {
			t.Cleanup(func() { _ = connection.Close() })
			return connection
		}
		if time.Now().After(deadline) {
			debugContext, debugCancel := context.WithTimeout(context.Background(), 5*time.Second)
			diagnosticConnection, diagnostic := amqp.DialConfig(
				settings.URL().Value(), buildConnectionConfig(debugContext, settings),
			)
			if diagnosticConnection != nil {
				_ = diagnosticConnection.Close()
			}
			debugCancel()
			detail := "unavailable"
			if diagnostic != nil {
				detail = strings.ReplaceAll(diagnostic.Error(), "integration-secret", "[REDACTED]")
			}
			t.Fatalf("RabbitMQ did not accept connections: %v (%s)", err, detail)
		}
		time.Sleep(250 * time.Millisecond)
	}
}

func getDelivery(t *testing.T, channel *amqp.Channel, queue string) amqp.Delivery {
	t.Helper()
	deadline := time.Now().Add(15 * time.Second)
	for {
		delivery, ok, err := channel.Get(queue, false)
		if err != nil {
			t.Fatal(err)
		}
		if ok {
			return delivery
		}
		if time.Now().After(deadline) {
			t.Fatalf("queue %s did not receive a message", queue)
		}
		time.Sleep(50 * time.Millisecond)
	}
}
