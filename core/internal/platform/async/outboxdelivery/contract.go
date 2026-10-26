package outboxdelivery

import (
	"context"
	"errors"
	"time"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/admission"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	brokerrabbit "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/rabbitmq"
	"github.com/google/uuid"
)

const (
	authorityPageSize = 128
	claimBatchSize    = 1
	recoveryBatchSize = 50
	idlePollInterval  = 250 * time.Millisecond
)

var (
	ErrDependencies = errors.New("outbox delivery dependencies are incomplete")
	ErrAlreadyStart = errors.New("outbox delivery workload already started")
)

type Broker interface {
	Publish(context.Context, brokerrabbit.PublishRequest) (brokerrabbit.PublishReceipt, error)
	Close() error
}

type Connector interface {
	Open(context.Context) (Broker, error)
}

type Settings struct {
	ConfirmTimeout   time.Duration
	ReconnectMinimum time.Duration
	ReconnectMaximum time.Duration
}

func (settings Settings) valid() bool {
	return settings.ConfirmTimeout > 0 && settings.ReconnectMinimum > 0 &&
		settings.ReconnectMaximum >= settings.ReconnectMinimum
}

type Runner struct {
	store       asyncpostgres.PublisherStore
	catalog     eventsv1.ProducerCatalog
	connector   Connector
	settings    Settings
	admission   *admission.Gate
	publisherID uuid.UUID
	jitter      jitterSource
	lifecycle   runnerLifecycle
}
