package operator

import (
	"context"
	"crypto/rand"
	"io"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/orchestration/environmentoperator"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	installationmanifest "github.com/Onurlulardan/harpy/core/internal/platform/installationmanifest"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
)

func composeEnvironmentCoordinator(
	ctx context.Context,
	settings config.OperatorInitialize,
	manifest installationmanifest.Manifest,
	pool *platformpostgres.Pool,
) (*environmentoperator.Coordinator, io.Closer, error) {
	runtime, err := BuildRuntime(pool)
	if err != nil {
		return nil, nil, err
	}
	offering, err := NewEnvironmentOffering(pool)
	if err != nil {
		return nil, nil, err
	}
	authority, authorityCloser, err := BuildEnvironmentAuthority(ctx, settings, manifest, pool)
	if err != nil {
		return nil, authorityCloser, err
	}
	keyring, err := settings.DirectoryKeyring()
	if err != nil {
		return nil, authorityCloser, err
	}
	finalizer, err := NewEnvironmentOperatorFinalizer(
		runtime, keyring, passwordLoader(settings.InitialPassword()),
	)
	if err != nil {
		return nil, authorityCloser, err
	}
	coordinator, err := environmentoperator.New(environmentoperator.Dependencies{
		Policy: runtime.DeploymentPolicy, Onboarding: runtime.Onboarding,
		Migration: runtime.TenantMigration, Offering: offering,
		Authority: authority, Finalizer: finalizer, Entropy: rand.Reader,
		Sleep: sleepContext, LeaseOwner: "core-operator-initialize",
		PollInterval: 250 * time.Millisecond,
	})
	if err != nil {
		return nil, authorityCloser, err
	}
	return coordinator, authorityCloser, nil
}
