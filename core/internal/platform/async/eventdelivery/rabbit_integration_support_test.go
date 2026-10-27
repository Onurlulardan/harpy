//go:build integration

package eventdelivery

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"strconv"
	"testing"
	"time"

	brokerrabbit "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/rabbitmq"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/configfixture"
	"github.com/moby/moby/api/types/container"
	mobynetwork "github.com/moby/moby/api/types/network"
	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const eventRabbitImage = "rabbitmq:4.3.2-alpine@sha256:835cbc6fabceedef7a3b0ce3195a17adccb753a3ba645bdd61c5557b6e4bf580"

type eventRabbitFixture struct {
	container testcontainers.Container
	settings  config.RabbitMQ
}

func startEventRabbit(t *testing.T) eventRabbitFixture {
	t.Helper()
	hostPort := reserveEventRabbitPort(t)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: eventRabbitImage,
			Env: map[string]string{
				"RABBITMQ_DEFAULT_USER":  "harpy_worker",
				"RABBITMQ_DEFAULT_PASS":  "integration-secret",
				"RABBITMQ_DEFAULT_VHOST": "harpy_eventdelivery",
			},
			ExposedPorts: []string{"5672/tcp"},
			HostConfigModifier: func(config *container.HostConfig) {
				if config.PortBindings == nil {
					config.PortBindings = mobynetwork.PortMap{}
				}
				port := mobynetwork.MustParsePort("5672/tcp")
				config.PortBindings[port] = []mobynetwork.PortBinding{{
					HostIP: netip.MustParseAddr("127.0.0.1"), HostPort: hostPort,
				}}
			},
			WaitingFor: wait.ForListeningPort("5672/tcp").
				WithStartupTimeout(2 * time.Minute),
		},
		Started: true,
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		cleanup, stop := context.WithTimeout(context.Background(), time.Minute)
		defer stop()
		if err := container.Terminate(cleanup); err != nil {
			t.Errorf("terminate RabbitMQ: %v", err)
		}
	})
	fixture := eventRabbitFixture{container: container}
	fixture.refresh(t)
	return fixture
}

func reserveEventRabbitPort(t *testing.T) string {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	return strconv.Itoa(port)
}

func (fixture *eventRabbitFixture) refresh(t *testing.T) {
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
	url := fmt.Sprintf(
		"amqp://harpy_worker:integration-secret@%s:%s/harpy_eventdelivery",
		host, port.Port(),
	)
	fixture.settings = loadEventRabbitSettings(t, url)
}

func loadEventRabbitSettings(t *testing.T, url string) config.RabbitMQ {
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

func dialEventRabbit(t *testing.T, settings config.RabbitMQ) *amqp.Connection {
	t.Helper()
	deadline := time.Now().Add(90 * time.Second)
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		connection, err := brokerrabbit.Dial(ctx, settings)
		cancel()
		if err == nil {
			t.Cleanup(func() { _ = connection.Close() })
			return connection
		}
		if time.Now().After(deadline) {
			t.Fatalf("RabbitMQ did not recover: %v", err)
		}
		time.Sleep(250 * time.Millisecond)
	}
}
