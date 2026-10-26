package operator

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/billing"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority"
	"github.com/Onurlulardan/harpy/core/internal/platform/localinstall"
)

func prepare(release, installation, authority string, now time.Time) error {
	manifest, err := localinstall.PrepareDevelopmentInstallation(release, installation)
	if err != nil {
		return err
	}
	defer clear(manifest)
	if err := billing.ReconcileDevelopmentAuthority(authority, manifest, now); err != nil {
		return err
	}
	return platformauthority.ReconcileDevelopmentRegistryArtifacts(
		authority,
		"/run/secrets/local-authority-ack-signing-seed",
		"/run/secrets/local-authority-release-signing-seed",
		now,
	)
}
