package api

import (
	"crypto/ed25519"
	"fmt"
	"os"
	"runtime"
	"time"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	platformauthority "github.com/Onurlulardan/harpy/core/internal/modules/platformauthority"
	recoveryanchors "github.com/Onurlulardan/harpy/core/internal/orchestration/recoveryanchors"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func buildPlatformOpsRecoveryAnchors(
	settings config.API,
	pool *platformpostgres.Pool,
	coordinator *platformtx.Coordinator,
) (*recoveryanchors.Service, error) {
	mode, err := settings.DeploymentMode()
	if err != nil {
		return nil, err
	}
	if mode != config.DeploymentOnPrem {
		return recoveryanchors.NewDisabledService(), nil
	}
	authorityConfig := settings.Authority()
	trust, err := readPlatformOpsAuthorityFile(
		authorityConfig.RecoveryTrustBundle().Path().Value(),
		32*1024, false)
	if err != nil {
		return nil, err
	}
	seed, err := readPlatformOpsAuthorityFile(
		authorityConfig.RecoveryControlKey().Path().Value(), ed25519.SeedSize, true)
	if err != nil {
		clear(trust)
		return nil, err
	}
	planner, err := platformauthority.NewOnPremisePlanner(
		platformauthority.OnPremisePlannerDependencies{Pool: pool,
			Clock: platformOpsAuthorityClock{}, RecoveryTrustBundle: trust,
			RecoveryControlSeed: seed})
	clear(trust)
	clear(seed)
	if err != nil {
		return nil, fmt.Errorf("build recovery anchor authority: %w", err)
	}
	authority, err := recoveryanchors.NewAuthority(planner)
	if err != nil {
		return nil, err
	}
	iamProfile, err := settings.IAMHTTPProfile()
	if err != nil {
		return nil, err
	}
	anchorRuntime, err := iammodule.NewRecoveryAnchorRuntime(pool, coordinator,
		authority, iamProfile)
	if err != nil {
		return nil, fmt.Errorf("build IAM recovery anchor runtime: %w", err)
	}
	result, err := recoveryanchors.NewService(anchorRuntime)
	if err != nil {
		return nil, err
	}
	return result, nil
}

type platformOpsAuthorityClock struct{}

func (platformOpsAuthorityClock) Now() time.Time {
	return time.Now().UTC().Truncate(time.Microsecond)
}

func readPlatformOpsAuthorityFile(path string, maximum int64, sensitive bool) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 ||
		info.Size() < 1 || info.Size() > maximum ||
		(runtime.GOOS != "windows" && sensitive && info.Mode().Perm()&0o077 != 0) {
		return nil, fmt.Errorf("recovery authority artifact is invalid")
	}
	document, err := os.ReadFile(path)
	if err != nil || int64(len(document)) != info.Size() {
		clear(document)
		return nil, fmt.Errorf("read recovery authority artifact")
	}
	return document, nil
}
