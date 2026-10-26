//go:build integration

package deadletterdelivery

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/configfixture"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const rabbitImage = "rabbitmq:4.3.2-alpine@sha256:835cbc6fabceedef7a3b0ce3195a17adccb753a3ba645bdd61c5557b6e4bf580"

type managedRabbitFixture struct {
	container testcontainers.Container
	settings  config.RabbitMQ
	url       string
}

func startManagedRabbit(t *testing.T) *managedRabbitFixture {
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
			WaitingFor: wait.ForListeningPort("5672/tcp").
				WithStartupTimeout(2 * time.Minute),
		},
		Started: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanup, cleanupCancel := context.WithTimeout(context.Background(), time.Minute)
		defer cleanupCancel()
		if err := container.Terminate(cleanup); err != nil {
			t.Errorf("terminate RabbitMQ: %v", err)
		}
	})
	fixture := &managedRabbitFixture{container: container}
	fixture.refresh(t)
	return fixture
}

func (fixture *managedRabbitFixture) refresh(t *testing.T) {
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
	t.Setenv("HARPY_ENVIRONMENT", "test")
	t.Setenv("HARPY_POSTGRES_ENDPOINT_KIND", "direct_postgres")
	t.Setenv("HARPY_POSTGRES_DSN", "postgres://runtime@127.0.0.1:1/db?sslmode=disable")
	t.Setenv("HARPY_RABBITMQ_URL", fixture.url)
	configfixture.SetWorker(t)
	worker, err := config.LoadWorker()
	if err != nil {
		t.Fatal(err)
	}
	fixture.settings = worker.RabbitMQ()
}
