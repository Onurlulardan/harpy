package operator

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

type developmentPublicationClock struct{}

func (developmentPublicationClock) Now() time.Time {
	return time.Now().UTC().Truncate(time.Second)
}

func runLocalAuthorityPublication(ctx context.Context) error {
	if err := platformauthority.PublishDevelopmentRecoveryTrust(
		"/run/dev-authority",
		"/run/recovery-signing/current",
	); err != nil {
		return err
	}
	settings, err := config.LoadOperatorInitialize()
	if err != nil {
		return err
	}
	manifest, err := readInstallationManifest(settings.InstallationManifest())
	if err != nil {
		return err
	}
	pool, err := openLocalOperatorPool(ctx, settings.Runtime())
	if err != nil {
		return err
	}
	defer pool.Close()
	_, err = platformauthority.BindDevelopmentInstallation(ctx,
		platformauthority.DevelopmentBindingDependencies{
			Pool: pool, Clock: developmentPublicationClock{},
			EnvironmentID:     manifest.EnvironmentIDString(),
			RecoveryTrustPath: "/run/dev-authority/current/recovery-trust-bundle.v1.json",
			RecoverySeedPath:  "/run/recovery-signing/current/recovery-control-signer.ed25519",
			AckTrustPath:      "/run/dev-authority/current/protected-authority-ack-trust-bundle.v1.json",
			ReleaseTrustPath:  "/run/dev-authority/current/protected-authority-release-trust-root.json",
			JournalPath:       "/run/dev-authority/protected-authority-heads.journal",
			AckSeedPath:       "/run/secrets/local-authority-ack-signing-seed",
		})
	if err != nil {
		return err
	}
	executor, err := platformpostgres.BindIAMOperatorInitialize(pool, func(
		database platformpostgres.IAMOperatorInitializeDatabase,
	) (iam.InstallationActivationExecutor, error) {
		return iam.InstallationActivationExecutor{QueryRow: func(
			ctx context.Context, statement string, arguments ...any,
		) iam.InstallationActivationRow {
			return database.QueryRow(ctx, statement, arguments...)
		}}, nil
	})
	if err != nil {
		return err
	}
	return iam.ReconcileInstallationPlatformOwner(
		ctx, executor, manifest.EnvironmentID(),
	)
}
