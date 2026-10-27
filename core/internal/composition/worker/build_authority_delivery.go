package worker

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

func buildAuthorityDeliveryWorkload(
	ctx context.Context,
	pool *platformpostgres.Pool,
	settings config.Worker,
) (managedAsyncWorkload, error) {
	if ctx == nil || ctx.Err() != nil || pool == nil {
		return nil, fmt.Errorf("platform authority delivery dependencies are invalid")
	}
	clock := authorityWorkerClock{}
	recoveryFinalizer, err := iam.NewRecoveryAnchorAuthorityFinalizer(pool)
	if err != nil {
		return nil, err
	}
	dependencies, err := authorityDeliveryDependencies(
		settings.Authority(),
		pool,
		clock,
		authorityOwnerFinalizer{recovery: recoveryFinalizer},
	)
	if err != nil {
		return nil, err
	}
	runtime, err := platformauthority.NewDeliveryRuntime(ctx, dependencies)
	if err != nil {
		return nil, err
	}
	ownerID, err := newAuthorityDeliveryOwnerID()
	if err != nil {
		return nil, err
	}
	return newAuthorityDeliveryWorkload(authorityDeliveryRunner{
		queue: runtime.Queue, processor: runtime.Processor, ownerID: ownerID,
	})
}

func authorityDeliveryDependencies(
	settings config.WorkerAuthority,
	pool *platformpostgres.Pool,
	clock platformauthority.DeliveryClock,
	finalizer platformauthority.DeliveryFinalizer,
) (platformauthority.DeliveryDependencies, error) {
	mode, err := authorityDeliveryMode(settings.Mode())
	if err != nil {
		return platformauthority.DeliveryDependencies{}, err
	}
	dependencies := platformauthority.DeliveryDependencies{
		Pool: pool, Clock: clock, Finalizer: finalizer, Mode: mode,
		AckTrustPath:     settings.AckTrustBundle().Path().Value(),
		ReleaseTrustPath: settings.ReleaseTrustRoot().Path().Value(),
	}
	switch mode {
	case platformauthority.DeliveryDevelopment:
		dependencies.RecoveryTrustPath =
			settings.RecoveryTrustBundle().Path().Value()
		dependencies.LocalJournalPath = settings.LocalJournal().Path().Value()
		dependencies.LocalAckSeedPath =
			settings.LocalAckSigningKey().Path().Value()
	case platformauthority.DeliverySaaS:
		dependencies.RegistryURL = settings.RegistryURL().Value()
		identity, configured := settings.WorkloadIdentity()
		dependencies.HasWorkloadIdentity = configured
		if configured {
			dependencies.WorkloadIdentityPath = identity.Path().Value()
		}
	}
	return dependencies, nil
}

func authorityDeliveryMode(
	mode config.WorkerAuthorityMode,
) (platformauthority.DeliveryMode, error) {
	switch mode {
	case config.WorkerAuthorityDevelopment:
		return platformauthority.DeliveryDevelopment, nil
	case config.WorkerAuthorityAirGap:
		return platformauthority.DeliveryAirGap, nil
	case config.WorkerAuthoritySaaS:
		return platformauthority.DeliverySaaS, nil
	default:
		return "", fmt.Errorf("platform authority delivery mode is invalid")
	}
}
