package environmentoperator

import (
	"context"

	installationv1 "github.com/Onurlulardan/harpy/contracts/installation/v1"
)

// waitClaimReady waits for the durable, database-clock-qualified executor
// receipt. Technical service readiness is not sufficient to open onboarding.
func (coordinator *Coordinator) waitClaimReady(
	ctx context.Context,
	request InitializeRequest,
) error {
	if request.Manifest.DeploymentMode() != installationv1.DeploymentOnPremise &&
		request.Manifest.DeploymentMode() != installationv1.DeploymentSaaS {
		return ErrInvalid
	}
	remaining := maximumMigrationWait
	for {
		ready, found, err := coordinator.migration.CurrentReady(ctx)
		if err != nil {
			return Resumable("tenant_migration.claim_ready", "claim_ready_failed", "", err)
		}
		if found {
			if !ready.IsValid() ||
				ready.EnvironmentID != request.Manifest.EnvironmentIDString() ||
				ready.DeploymentMode != string(request.Manifest.DeploymentMode()) {
				return ErrDependencyMismatch
			}
			return nil
		}
		if remaining <= 0 {
			return ErrDependencyMismatch
		}
		wait := min(coordinator.poll, remaining)
		if err := coordinator.sleep(ctx, wait); err != nil {
			return Resumable("tenant_migration.claim_ready_wait", "claim_ready_wait_interrupted", "", err)
		}
		remaining -= wait
	}
}
