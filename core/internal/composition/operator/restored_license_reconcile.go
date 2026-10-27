package operator

import (
	"context"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing"
	platformops "github.com/Onurlulardan/harpy/core/internal/modules/platformops"
	"github.com/Onurlulardan/harpy/core/internal/platform/buildinfo"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

func reconcileRestoredLicense(
	ctx context.Context,
) (restoredLicenseOutcome, string, error) {
	settings, err := config.LoadOperatorLicenseReconcile()
	if err != nil {
		return restoredLicenseOutcome{}, "configuration", err
	}
	info, err := buildinfo.ReadOperator(settings.Runtime().Environment())
	if err != nil {
		return restoredLicenseOutcome{}, "build_contract", err
	}
	pool, err := openOperatorPool(
		ctx, settings.Runtime(), platformpostgres.ProfileOperatorLicenseReconcile,
	)
	if err != nil {
		return restoredLicenseOutcome{}, "database", err
	}
	defer pool.Close()
	database, err := bindRestoredLicenseDatabase(pool)
	if err != nil {
		return restoredLicenseOutcome{}, "composition", err
	}
	databaseNow, err := billing.ReadLocalLicenseDatabaseClock(ctx, database.clock)
	if err != nil {
		return restoredLicenseOutcome{}, "database", err
	}
	evidence, err := loadRestoredLicenseEvidence(settings, databaseNow)
	if err != nil {
		return restoredLicenseOutcome{}, "evidence", err
	}
	replay, found, err := readRestoredLicenseReplay(ctx, database.billing, evidence)
	if err != nil {
		return restoredLicenseOutcome{}, "evidence", err
	}
	if found {
		return replay, "reconcile", nil
	}
	binding, err := billing.LoadLocalLicenseRestoreBinding(
		ctx, database.billing, evidence.EnvironmentID,
	)
	if err != nil {
		return restoredLicenseOutcome{}, "binding", err
	}
	if evidence.EnvironmentID != binding.Binding.EnvironmentID {
		return restoredLicenseOutcome{}, "evidence",
			fmt.Errorf("restored license receipt does not match database")
	}
	trust, err := billing.ReadLocalLicenseTrustBundle(
		ctx, settings.LicenseTrustBundle().Path().Value(), binding.DatabaseNow,
	)
	if err != nil {
		return restoredLicenseOutcome{}, "evidence", err
	}
	document, err := issueRestoredLicenseSnapshot(
		ctx, settings, binding, evidence.OperationID, info.Version().Version,
	)
	if err != nil {
		return restoredLicenseOutcome{}, "source", err
	}
	defer clear(document)
	snapshot, err := billing.VerifyLocalLicenseSnapshot(
		ctx, document, binding.Binding, binding.DatabaseNow, trust,
	)
	if err != nil || snapshot.Sequence() <= binding.Sequence ||
		snapshot.Issuer() != binding.Issuer {
		return restoredLicenseOutcome{}, "verification",
			fmt.Errorf("restored license snapshot is invalid: %w", err)
	}
	grant, err := platformops.MintRestoredLicenseGrant(
		ctx, database.grants, platformops.RestoredLicenseGrantInput{
			OperationID: evidence.OperationID, EnvironmentID: evidence.EnvironmentID,
			RestoreReceiptID:       evidence.RestoreReceiptID,
			RestoreReceiptSHA256:   evidence.RestoreReceiptSHA256,
			LicenseID:              binding.Binding.LicenseID,
			LicenseeOrganizationID: binding.Binding.LicenseeOrganizationID,
			DeliveryKind:           string(binding.Binding.DeliveryKind), Issuer: binding.Issuer,
			PreviousSequence: binding.Sequence, PreviousSHA256: binding.SHA256,
			NextSequence: snapshot.Sequence(), NextSHA256: snapshot.Digest(),
			AuthorityRevision: binding.AuthorityRevision,
		},
	)
	if err != nil {
		return restoredLicenseOutcome{}, "capability", err
	}
	coordinator, err := billing.NewLocalRestoredLicenseCoordinator(database.billing, trust)
	if err != nil {
		return restoredLicenseOutcome{}, "composition", err
	}
	var receipt billing.LocalRestoredLicenseReceipt
	err = platformops.RunRestoredLicenseGrant(
		grant.Grant, grant.DatabaseNow,
		func(scope platformops.RestoredLicenseScope) error {
			var restoreErr error
			receipt, restoreErr = coordinator.Restore(
				ctx, scope, document, grant.DatabaseNow,
			)
			return restoreErr
		},
	)
	if err != nil {
		return restoredLicenseOutcome{}, "reconcile", err
	}
	return restoredLicenseOutcome{
		OperationID: receipt.OperationID, EnvironmentID: evidence.EnvironmentID,
		PreviousSequence:  receipt.PreviousSequence,
		AcceptedSequence:  receipt.AcceptedSequence,
		AuthorityRevision: receipt.AuthorityRevision, Replayed: receipt.Replay,
	}, "reconcile", nil
}
