package application

import (
	"context"
	"errors"
	"reflect"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

var (
	ErrDeliveryWorker = errors.New("notification delivery worker is not configured")
	ErrOutcomeUnknown = errors.New("notification provider outcome is unknown")
)

type DeliveryWorker struct {
	store    ports.DeliveryRuntimeStore
	targets  ports.DeliveryTargetResolver
	opener   ports.SecretOpener
	registry ports.ProviderRegistry
	clock    ports.Clock
	ids      ports.IDSource
}

func NewDeliveryWorker(store ports.DeliveryRuntimeStore,
	targets ports.DeliveryTargetResolver, opener ports.SecretOpener,
	registry ports.ProviderRegistry, clock ports.Clock, ids ports.IDSource,
) (*DeliveryWorker, error) {
	dependencies := []any{store, targets, opener, registry, clock, ids}
	for _, dependency := range dependencies {
		if missingWorkerDependency(dependency) {
			return nil, ErrDeliveryWorker
		}
	}
	return &DeliveryWorker{store: store, targets: targets, opener: opener,
		registry: registry, clock: clock, ids: ids}, nil
}

func missingWorkerDependency(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map,
		reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}

type DeliveryProcessResult struct {
	Outcome ports.ProviderOutcome
	Code    string
}

func (worker *DeliveryWorker) Execute(ctx context.Context, tenantID,
	deliveryID domain.ID, owner string,
) (DeliveryProcessResult, error) {
	if worker == nil || ctx == nil || !tenantID.Valid() || !deliveryID.Valid() ||
		owner == "" {
		return DeliveryProcessResult{}, ErrDeliveryWorker
	}
	now := worker.clock.Now().UTC()
	work, err := worker.store.AcquireDelivery(ctx, tenantID, deliveryID, owner, now)
	if err != nil {
		return DeliveryProcessResult{}, err
	}
	if work.Delivery.State == domain.DeliverySuppressed {
		return DeliveryProcessResult{Outcome: ports.ProviderPermanent,
			Code: "PREFERENCE_DISABLED"}, nil
	}
	if work.Delivery.State != domain.DeliverySending {
		return DeliveryProcessResult{}, ErrDeliveryWorker
	}
	return worker.execute(ctx, tenantID, work, now)
}
