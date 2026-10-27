package platformops

import (
	"fmt"
	"os"

	recoverysigningv1 "github.com/Onurlulardan/harpy/contracts/recoverysigning/v1"
	plansigner "github.com/Onurlulardan/harpy/core/internal/modules/platformops/adapters/recoveryplansigner"
	platformopsapp "github.com/Onurlulardan/harpy/core/internal/modules/platformops/application"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

func newRuntimeRecoveryService(
	settings config.API,
	store *runtimeStore,
	clock runtimeClock,
) (*platformopsapp.RecoveryService, error) {
	mode, err := settings.DeploymentMode()
	if err != nil {
		return nil, fmt.Errorf("read PlatformOps deployment mode: %w", err)
	}
	if mode == config.DeploymentSaaS {
		service := platformopsapp.NewRecoveryService(store, store, store, store, clock)
		if service == nil {
			return nil, fmt.Errorf("build SaaS PlatformOps recovery service")
		}
		return service, nil
	}
	authority := settings.Authority()
	bundle, err := os.ReadFile(authority.RecoveryTrustBundle().Path().Value())
	if err != nil || len(bundle) == 0 || len(bundle) > recoverysigningv1.MaximumBundleBytes {
		return nil, fmt.Errorf("read PlatformOps recovery trust bundle")
	}
	seed, err := os.ReadFile(authority.RecoveryControlKey().Path().Value())
	if err != nil {
		return nil, fmt.Errorf("read PlatformOps recovery control key")
	}
	signer, err := plansigner.New(bundle, seed)
	clear(seed)
	if err != nil {
		return nil, fmt.Errorf("build PlatformOps restore plan signer: %w", err)
	}
	service := platformopsapp.NewRecoveryServiceWithSigner(
		store, store, store, store, signer, clock)
	if service == nil {
		return nil, fmt.Errorf("build PlatformOps recovery service")
	}
	return service, nil
}
