//go:build integration

package postgres

import (
	"context"
	"fmt"
	"testing"
	"time"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	brokerrabbit "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/rabbitmq"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const pipelineRabbitImage = "rabbitmq:4.3.2-alpine@sha256:" +
	"835cbc6fabceedef7a3b0ce3195a17adccb753a3ba645bdd61c5557b6e4bf580"

type pipelineRabbitFixture struct {
	url string
}

func startPipelineRabbit(t *testing.T, ctx context.Context) pipelineRabbitFixture {
	t.Helper()
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: pipelineRabbitImage,
			Env: map[string]string{
				"RABBITMQ_DEFAULT_USER":  "harpy_worker",
				"RABBITMQ_DEFAULT_PASS":  "integration-secret",
				"RABBITMQ_DEFAULT_VHOST": "harpy_confirm_crash",
			},
			ExposedPorts: []string{"5672/tcp"},
			WaitingFor: wait.ForListeningPort("5672/tcp").
				WithStartupTimeout(2 * time.Minute),
		},
		Started: true,
	})
	if err != nil {
		t.Fatalf("start exact RabbitMQ 4.3.2 image: %v", err)
	}
	t.Cleanup(func() {
		cleanup, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()
		if err := container.Terminate(cleanup); err != nil {
			t.Errorf("terminate RabbitMQ: %v", err)
		}
	})
	host, err := container.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}
	port, err := container.MappedPort(ctx, "5672/tcp")
	if err != nil {
		t.Fatal(err)
	}
	return pipelineRabbitFixture{url: fmt.Sprintf(
		"amqp://harpy_worker:integration-secret@%s:%s/harpy_confirm_crash",
		host, port.Port(),
	)}
}

func dialPipelineRabbit(
	t *testing.T,
	ctx context.Context,
	settings config.RabbitMQ,
) *amqp.Connection {
	t.Helper()
	deadline := time.Now().Add(90 * time.Second)
	for {
		connection, err := brokerrabbit.Dial(ctx, settings)
		if err == nil {
			return connection
		}
		if time.Now().After(deadline) {
			t.Fatalf("RabbitMQ did not accept pipeline connection: %v", err)
		}
		select {
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case <-time.After(100 * time.Millisecond):
		}
	}
}

func openPipelinePublisher(
	t *testing.T,
	ctx context.Context,
	settings config.RabbitMQ,
	consumers eventsv1.ConsumerCatalog,
) (*brokerrabbit.Publisher, *amqp.Connection) {
	t.Helper()
	connection := dialPipelineRabbit(t, ctx, settings)
	if err := brokerrabbit.DeclareTopology(connection, consumers); err != nil {
		_ = connection.Close()
		t.Fatalf("declare pipeline topology: %v", err)
	}
	publisher, err := brokerrabbit.OpenPublisher(connection)
	if err != nil {
		_ = connection.Close()
		t.Fatalf("open pipeline publisher: %v", err)
	}
	return publisher, connection
}
