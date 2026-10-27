package localinstall

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	developmentManifestFile = "manifest.json"
)

func PrepareDevelopmentInstallation(
	release string,
	installation string,
) ([]byte, error) {
	info, err := os.Lstat(installation)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("local state directory is invalid")
	}
	environmentID, err := readDevelopmentEnvironmentID(release)
	if err != nil {
		return nil, err
	}
	manifest, err := reconcileDevelopmentManifest(
		filepath.Join(installation, developmentManifestFile), environmentID,
	)
	if err != nil {
		return nil, err
	}
	return manifest.CanonicalBytes(), nil
}
