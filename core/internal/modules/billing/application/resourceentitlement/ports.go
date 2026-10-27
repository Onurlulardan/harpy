package resourceentitlement

import (
	"context"
	"time"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/entitlement"
)

type Store interface {
	Inspect(context.Context, billingapi.MeteredResourceMutationRequest) (Inspection, error)
	ResolveFeature(context.Context, entitlement.Key) (bool, error)
	Record(context.Context, MutationRecord) error
}

type ReconciliationStore interface {
	InspectReconciliation(
		context.Context,
		billingapi.MeteredResourceReconcileRequest,
	) (ReconciliationInspection, error)
	ResolveFeature(context.Context, entitlement.Key) (bool, error)
	RecordReconciliation(context.Context, ReconciliationRecord) error
}

type ResourceStore interface {
	Store
	ReconciliationStore
}

type Usage interface {
	InspectGauge(context.Context, entitlement.Key) (GaugeSnapshot, error)
	AdjustGauge(context.Context, GaugeAdjustmentRequest) (GaugeAdjustment, error)
}

type Clock func() time.Time
