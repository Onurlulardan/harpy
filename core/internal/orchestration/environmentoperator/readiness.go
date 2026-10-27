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
			return ErrResumable
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
			return ErrResumable
		}
		remaining -= wait
	}
}
