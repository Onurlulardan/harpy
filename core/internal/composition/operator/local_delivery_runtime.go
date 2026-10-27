package operator

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	"github.com/Onurlulardan/harpy/core/internal/orchestration/environmentoperator"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	installationmanifest "github.com/Onurlulardan/harpy/core/internal/platform/installationmanifest"
	"github.com/Onurlulardan/harpy/core/internal/platform/localinstall"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

type deliveryRuntime struct {
	settings config.OperatorInitialize
	manifest installationmanifest.Manifest
	identity environmentoperator.InstallationIdentity
	delivery *tenancy.InstallationDeliveryRuntime
	pool     *platformpostgres.Pool
}

func openDeliveryRuntime(ctx context.Context) (*deliveryRuntime, error) {
	settings, err := config.LoadOperatorInitialize()
	if err != nil {
		return nil, err
	}
	document, err := localinstall.ReadRegular(
		settings.InstallationManifest().Path().Value(), installationmanifest.MaximumBytes)
	if err != nil {
		return nil, err
	}
	manifest, err := installationmanifest.Parse(document)
	clear(document)
	if err != nil {
		return nil, err
	}
	identity := environmentoperator.DeriveInstallationIdentity(manifest)
	if !identity.Valid() {
		return nil, installationv1Error{}
	}
	pool, err := openLocalOperatorPool(ctx, settings.Runtime())
	if err != nil {
		return nil, err
	}
	delivery, err := buildInstallationDelivery(pool)
	if err != nil {
		pool.Close()
		return nil, err
	}
	return &deliveryRuntime{settings: settings, manifest: manifest,
		identity: identity, delivery: delivery, pool: pool}, nil
}

func (runtime *deliveryRuntime) close() { runtime.pool.Close() }

type installationv1Error struct{}

func (installationv1Error) Error() string { return "installation identity is invalid" }
