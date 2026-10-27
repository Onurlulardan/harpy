package platformauthority

import (
	"context"
	"fmt"

	postgresadapter "github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

func NewDeliveryRuntime(
	ctx context.Context,
	dependencies DeliveryDependencies,
) (DeliveryRuntime, error) {
	if ctx == nil || ctx.Err() != nil || dependencies.Pool == nil ||
		dependencies.Clock == nil || dependencies.Finalizer == nil {
		return DeliveryRuntime{}, fmt.Errorf(
			"platform authority delivery dependencies are invalid",
		)
	}
	store, err := postgresadapter.NewStore(
		dependencies.Pool, dependencies.Clock,
	)
	if err != nil {
		return DeliveryRuntime{}, err
	}
	queue, err := application.NewDeliveryQueue(store)
	if err != nil {
		return DeliveryRuntime{}, err
	}
	kind := deliveryRegistryKind(dependencies.Mode)
	trust, err := verifyDeliveryTrust(dependencies, kind)
	if err != nil {
		return DeliveryRuntime{}, err
	}
	var registry ports.Registry
	if dependencies.Mode != DeliveryAirGap {
		registry, err = newDeliveryRegistry(ctx, dependencies, trust)
		if err != nil {
			return DeliveryRuntime{}, err
		}
	}
	processor, err := application.NewDeliveryProcessor(
		store, registry, dependencies.Clock, trust, kind,
		dependencies.Finalizer, dependencies.Mode == DeliveryAirGap,
	)
	if err != nil {
		return DeliveryRuntime{}, err
	}
	return DeliveryRuntime{Queue: queue, Processor: processor}, nil
}

func deliveryRegistryKind(mode DeliveryMode) ports.RegistryKind {
	switch mode {
	case DeliveryDevelopment, DeliveryAirGap:
		return ports.RegistryDevelopment
	case DeliverySaaS:
		return ports.RegistrySaaS
	default:
		return ""
	}
}
