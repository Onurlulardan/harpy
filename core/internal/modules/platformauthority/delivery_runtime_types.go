package platformauthority

import (
	"context"

	postgresadapter "github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

type DeliveryMode string

const (
	DeliveryAirGap      DeliveryMode = "airgap"
	DeliveryDevelopment DeliveryMode = "development"
	DeliverySaaS        DeliveryMode = "saas"
)

type DeliveryPool = postgresadapter.Pool
type DeliveryClock = ports.Clock
type DeliveryClaim = ports.DeliveryClaim
type DeliveryOperation = ports.OperationRecord
type DeliveryFinalizer = ports.OperationFinalizer

var (
	ErrDeliveryAwaitingImport       = application.ErrDeliveryAwaitingImport
	ErrDeliveryPermanent            = application.ErrDeliveryPermanent
	ErrDeliveryFinalizerUnavailable = ports.ErrFinalizerUnavailable
)

type DeliveryQueue interface {
	ClaimDelivery(context.Context, string) (DeliveryClaim, bool, error)
	RetryDelivery(context.Context, DeliveryClaim, string, bool) error
	UpdateDeliveryImport(context.Context, DeliveryClaim) error
	CompleteDelivery(context.Context, DeliveryClaim) error
}

type DeliveryProcessor interface {
	Process(context.Context, string) error
}

type DeliveryDependencies struct {
	Pool                 *DeliveryPool
	Clock                DeliveryClock
	Finalizer            DeliveryFinalizer
	Mode                 DeliveryMode
	AckTrustPath         string
	ReleaseTrustPath     string
	RecoveryTrustPath    string
	LocalJournalPath     string
	LocalAckSeedPath     string
	RegistryURL          string
	WorkloadIdentityPath string
	HasWorkloadIdentity  bool
}

type DeliveryRuntime struct {
	Queue     DeliveryQueue
	Processor DeliveryProcessor
}
