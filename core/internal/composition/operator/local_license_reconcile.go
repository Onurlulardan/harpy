package operator

import (
	"context"
	"fmt"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	installationmanifest "github.com/Onurlulardan/harpy/core/internal/platform/installationmanifest"
	"github.com/Onurlulardan/harpy/core/internal/platform/localinstall"
)

type databaseClock struct{ now time.Time }

func reconcileLocalDevelopmentLicense(ctx context.Context) error {
	settings, err := config.LoadOperatorInitialize()
	if err != nil {
		return err
	}
	document, err := localinstall.ReadRegular(
		settings.InstallationManifest().Path().Value(), installationmanifest.MaximumBytes)
	if err != nil {
		return err
	}
	manifest, err := installationmanifest.Parse(document)
	clear(document)
	if err != nil {
		return err
	}
	environmentID := billing.UUID(manifest.EnvironmentID())
	if !billing.IsUUIDv7(environmentID) {
		return fmt.Errorf("local environment identity is not UUIDv7")
	}
	pool, err := openLocalOperatorPool(ctx, settings.Runtime())
	if err != nil {
		return err
	}
	defer pool.Close()
	database, err := bindBillingDatabase(pool)
	if err != nil {
		return err
	}
	clock, err := database.clock(ctx)
	if err != nil {
		return err
	}
	trust, err := billing.ReadLocalLicenseTrustBundle(ctx,
		settings.LicenseTrustBundle().Path().Value(), clock.now.UTC())
	if err != nil {
		return err
	}
	snapshot, err := reconcileSnapshot(ctx, settings, database, trust,
		localinstall.DeriveIdentity(environmentID), environmentID)
	if err != nil {
		return err
	}
	return publishOffering(ctx, database, trust, snapshot, clock.now.UTC(),
		localinstall.DeriveIdentity(environmentID))
}

func reconcileSnapshot(
	ctx context.Context,
	settings config.OperatorInitialize,
	database *billingDatabase,
	trust billing.LocalLicenseTrustBundle,
	identity localinstall.Identity,
	environmentID billing.UUID,
) (billing.LocalLicenseVerifiedSnapshot, error) {
	store, err := database.licenseStore()
	if err != nil {
		return billing.LocalLicenseVerifiedSnapshot{}, err
	}
	source, err := billing.NewLocalLicenseSnapshotFile(
		settings.LicenseSource().SnapshotFile().Path().Value())
	if err != nil {
		return billing.LocalLicenseVerifiedSnapshot{}, err
	}
	reconciler, err := billing.NewLocalLicenseReconciler(store, source, trust,
		[]billing.LocalLicenseComponentVersion{{Component: "core-operator", Version: "local-v2"}})
	if err != nil {
		return billing.LocalLicenseVerifiedSnapshot{}, err
	}
	result, err := reconciler.Reconcile(ctx, billing.LocalLicenseRefreshCommand{
		OperationID: identity.RefreshOperationID,
		Binding: billing.LocalLicenseBinding{
			LicenseID:              identity.LicenseID,
			LicenseeOrganizationID: identity.LicenseeOrganizationID,
			EnvironmentID:          environmentID,
			DeliveryKind:           billing.LocalLicenseDeliveryDevelopment,
		},
	})
	return result.Snapshot, err
}
