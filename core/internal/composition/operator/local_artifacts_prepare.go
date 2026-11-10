package operator

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority"
	"github.com/Onurlulardan/harpy/core/internal/platform/localinstall"
)

func prepare(release, installation, authority string, now time.Time) error {
	manifest, err := localinstall.PrepareDevelopmentInstallation(release, installation)
	if err != nil {
		return err
	}
	defer clear(manifest)
	current := filepath.Join(authority, "current")
	if err = os.MkdirAll(current, 0o700); err != nil {
		return err
	}
	info, err := os.Lstat(current)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("local authority directory is invalid")
	}
	return platformauthority.ReconcileDevelopmentRegistryArtifacts(
		authority,
		"/run/secrets/local-authority-ack-signing-seed",
		"/run/secrets/local-authority-release-signing-seed",
		now,
	)
}
