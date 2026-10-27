package billing

import (
	"context"
	"errors"
	"fmt"

	installationpostgres "github.com/Onurlulardan/harpy/core/internal/modules/billing/adapters/postgres/installationactivation"
	"github.com/google/uuid"
)

var ErrInstallationActivationInvalid = errors.New("billing installation activation is invalid")

func AssignInstallationBaseOffering(
	ctx context.Context,
	executor InstallationActivationExecutor,
	input InstallationBaseOfferingInput,
) error {
	if ctx == nil || ctx.Err() != nil || executor.QueryRow == nil ||
		!validInstallationBaseOffering(input) {
		return ErrInstallationActivationInvalid
	}
	binding, err := bindInstallationActivation(executor)
	if err != nil {
		return err
	}
	valid, err := binding.Assert(ctx, uuid.UUID(input.OperationID),
		uuid.UUID(input.TenantID), input.SchemaName)
	if err != nil || !valid {
		return fmt.Errorf("assert billing installation activation: %w", err)
	}
	authorityID := billingInstallationID(input.OperationID, "tenant-authority")
	assignmentID := billingInstallationID(input.OperationID, "base-offering-assignment")
	var sourceSequence any
	var sourceSHA256 any
	if input.Source.Present {
		sourceSequence, sourceSHA256 = input.Source.Sequence, input.Source.SHA256
	}
	_, err = binding.Assign(ctx, installationpostgres.Assignment{
		OperationID: uuid.UUID(input.OperationID), EnvironmentID: uuid.UUID(input.EnvironmentID),
		TenantID: uuid.UUID(input.TenantID), AuthorityID: uuid.UUID(authorityID),
		AssignmentID: uuid.UUID(assignmentID), DeploymentMode: input.DeploymentMode,
		BindingPurpose:    input.BindingPurpose,
		OfferingVersionID: uuid.UUID(input.OfferingVersionID),
		OfferingRevision:  input.OfferingRevision, OfferingSHA256: input.OfferingSHA256,
		OfferingEffectiveAt: input.OfferingEffectiveAt, SourceSequence: sourceSequence,
		SourceSHA256: sourceSHA256, ManifestSHA256: input.ManifestSHA256,
	})
	if err != nil {
		return fmt.Errorf("assign billing installation base offering: %w", err)
	}
	return nil
}

func OpenInstallationUsage(
	ctx context.Context,
	executor InstallationActivationExecutor,
	input InstallationUsageInput,
) error {
	if ctx == nil || ctx.Err() != nil || executor.QueryRow == nil ||
		zeroInstallationUUID(input.OperationID) || zeroInstallationUUID(input.TenantID) ||
		!installationSHA256Pattern.MatchString(input.ManifestSHA256) ||
		len(input.SchemaName) < 3 || len(input.SchemaName) > 63 {
		return ErrInstallationActivationInvalid
	}
	binding, err := bindInstallationActivation(executor)
	if err != nil {
		return err
	}
	valid, err := binding.Assert(ctx, uuid.UUID(input.OperationID),
		uuid.UUID(input.TenantID), input.SchemaName)
	if err != nil || !valid {
		return fmt.Errorf("assert billing usage activation: %w", err)
	}
	meterCount, replay, err := binding.Open(ctx, installationpostgres.Opening{
		OperationID: uuid.UUID(input.OperationID), TenantID: uuid.UUID(input.TenantID),
		ManifestSHA256: input.ManifestSHA256,
	})
	if err != nil || meterCount < 1 {
		return fmt.Errorf("open billing installation usage: meters=%d replay=%t: %w",
			meterCount, replay, err)
	}
	return nil
}

func bindInstallationActivation(
	executor InstallationActivationExecutor,
) (installationpostgres.Binding, error) {
	return installationpostgres.Bind(installationpostgres.Database{QueryRow: func(
		ctx context.Context, statement string, arguments ...any,
	) installationpostgres.Row {
		return executor.QueryRow(ctx, statement, arguments...)
	}})
}
