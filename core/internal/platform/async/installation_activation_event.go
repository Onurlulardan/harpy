package async

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	installationpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres/installationactivation"
)

var ErrInstallationActivationEventInvalid = errors.New(
	"installation activation event is invalid",
)

var installationEventDigestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

var installationEventSchemaPattern = regexp.MustCompile(`^tenant_[0-9a-f]{32}$`)

type InstallationActivationRow interface{ Scan(...any) error }

type InstallationActivationExecutor struct {
	QueryRow func(context.Context, string, ...any) InstallationActivationRow
}

type InstallationActivationEventInput struct {
	OperationID, EnvironmentID, TenantID, ActorID [16]byte
	SchemaName, ManifestSHA256                    string
}

func RecordInstallationActivationEvent(
	ctx context.Context,
	executor InstallationActivationExecutor,
	input InstallationActivationEventInput,
) error {
	if ctx == nil || ctx.Err() != nil || executor.QueryRow == nil ||
		input.OperationID == [16]byte{} || input.EnvironmentID == [16]byte{} ||
		input.TenantID == [16]byte{} || input.ActorID == [16]byte{} ||
		!installationEventSchemaPattern.MatchString(input.SchemaName) ||
		!installationEventDigestPattern.MatchString(input.ManifestSHA256) {
		return ErrInstallationActivationEventInvalid
	}
	binding, err := installationpostgres.Bind(installationpostgres.Database{
		QueryRow: func(
			ctx context.Context,
			statement string,
			arguments ...any,
		) installationpostgres.Row {
			return executor.QueryRow(ctx, statement, arguments...)
		},
	})
	if err != nil {
		return fmt.Errorf("bind installation activation event owner: %w", err)
	}
	_, err = binding.Record(ctx,
		input.OperationID, input.EnvironmentID, input.TenantID,
		input.ActorID, input.SchemaName, input.ManifestSHA256,
	)
	if err != nil {
		return fmt.Errorf("record installation activation event: %w", err)
	}
	return nil
}
