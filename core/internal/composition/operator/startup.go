package operator

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	"github.com/Onurlulardan/harpy/core/internal/orchestration/environmentoperator"
	"github.com/Onurlulardan/harpy/core/internal/platform/buildinfo"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	installationmanifest "github.com/Onurlulardan/harpy/core/internal/platform/installationmanifest"
	"github.com/Onurlulardan/harpy/core/internal/platform/migration"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/processcode"
	recoverysigning "github.com/Onurlulardan/harpy/core/internal/platform/recoverysigning"
	"github.com/Onurlulardan/harpy/core/internal/platform/schemacontract"
)

func initializeEnvironment(ctx context.Context) (
	environmentoperator.InitializeResult,
	string,
	error,
) {
	settings, err := config.LoadOperatorInitialize()
	if err != nil {
		return environmentoperator.InitializeResult{}, "configuration", err
	}
	if _, err = buildinfo.ReadOperator(settings.Runtime().Environment()); err != nil {
		return environmentoperator.InitializeResult{}, "build_contract", err
	}
	registry, err := migration.LoadReleaseRegistry()
	if err != nil {
		return environmentoperator.InitializeResult{}, "build_contract", err
	}
	embedded, err := schemacontract.LoadEmbedded(registry)
	if err != nil {
		return environmentoperator.InitializeResult{}, "build_contract", err
	}
	manifest, err := readInstallationManifest(settings.InstallationManifest())
	if err != nil {
		return environmentoperator.InitializeResult{}, "manifest", err
	}
	if string(settings.DeploymentMode()) != string(manifest.DeploymentMode()) {
		return environmentoperator.InitializeResult{}, "mode",
			environmentoperator.ErrDependencyMismatch
	}
	recovery, err := reconcileRecoverySigning(ctx, settings, manifest)
	if err != nil {
		return environmentoperator.InitializeResult{}, "recovery_signing", err
	}
	databaseContext, cancel := context.WithTimeout(ctx,
		settings.Runtime().Postgres().ConnectTimeout()+settings.Runtime().ReadinessTimeout())
	pool, err := platformpostgres.Open(databaseContext, settings.Runtime().Postgres(),
		platformpostgres.ProfileOperatorInitialize, embedded)
	cancel()
	if err != nil {
		return environmentoperator.InitializeResult{}, "database", err
	}
	defer pool.Close()
	coordinator, closer, err := composeEnvironmentCoordinator(ctx, settings, manifest, pool)
	if closer != nil {
		defer closer.Close()
	}
	if err != nil {
		return environmentoperator.InitializeResult{}, "composition", err
	}
	delivery, err := buildInstallationDelivery(pool)
	if err != nil {
		return environmentoperator.InitializeResult{}, "composition", err
	}
	authoritySynchronizer, err := tenancy.BindTenantAuthoritySynchronizer(pool)
	if err != nil {
		return environmentoperator.InitializeResult{}, "composition", err
	}
	result, err := coordinator.Initialize(ctx, environmentoperator.InitializeRequest{
		Manifest: manifest,
		RecoverySigning: environmentoperator.RecoverySigningResult{
			Generation: recovery.Generation, TrustBundleSHA256: recovery.TrustBundleSHA256,
			Reused: recovery.Reused,
		},
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "initialize error: %v\n", err)
		return result, "initialize", err
	}
	if _, err = authoritySynchronizer.Synchronize(
		ctx, result.OperationID, manifest.SHA256(),
	); err != nil {
		return result, "tenant_authority", err
	}
	if err = stageInstallationDelivery(
		ctx, delivery, manifest, result, settings.InitialPassword(),
	); err != nil {
		return result, "credential_delivery", err
	}
	return result, "initialize", nil
}

func reconcileRecoverySigning(
	ctx context.Context,
	settings config.OperatorInitialize,
	manifest installationmanifest.Manifest,
) (recoverysigning.Result, error) {
	if settings.DeploymentMode() == config.DeploymentSaaS {
		return recoverysigning.Result{}, nil
	}
	return recoverysigning.Reconcile(ctx,
		settings.RecoverySigningDirectory().Path().Value(), manifest.EnvironmentIDString())
}

func classifyRecoveryError(err error) processcode.Code {
	if errors.Is(err, recoverysigning.ErrFenced) {
		return processcode.OperatorTerminal
	}
	if errors.Is(err, recoverysigning.ErrInvalid) {
		return processcode.InvalidStartup
	}
	return processcode.OperatorResumable
}

func sleepContext(ctx context.Context, duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
