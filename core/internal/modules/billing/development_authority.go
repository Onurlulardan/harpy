package billing

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	installationv1 "github.com/Onurlulardan/harpy/contracts/installation/v1"
	"github.com/Onurlulardan/harpy/core/internal/platform/localinstall"
	"github.com/google/uuid"
)

const (
	developmentAuthorityCurrent = "current"
	developmentRootSeedFile     = "root.seed"
	developmentSignerSeedFile   = "snapshot-signer.seed"
	developmentTrustBundleFile  = "license-trust.json"
	developmentSnapshotFile     = "license-snapshot.json"
	developmentRootKeyID        = "local-license-root-1"
	developmentSignerKeyID      = "local-license-signer-1"
	developmentLicenseIssuer    = "harpy-local-dev"
)

func ReconcileDevelopmentAuthority(
	root string,
	manifestDocument []byte,
	now time.Time,
) error {
	info, err := os.Lstat(root)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("local authority directory is invalid")
	}
	manifest, err := installationv1.Parse(manifestDocument)
	if err != nil {
		return fmt.Errorf("local installation manifest is invalid: %w", err)
	}
	current := filepath.Join(root, developmentAuthorityCurrent)
	if _, err = os.Lstat(current); err == nil {
		return verifyDevelopmentAuthority(current, manifest, now)
	} else if !os.IsNotExist(err) {
		return err
	}
	staged := filepath.Join(root, ".staged-"+uuid.Must(uuid.NewV7()).String())
	if err = os.Mkdir(staged, 0o700); err != nil {
		return err
	}
	published := false
	defer func() {
		if !published {
			_ = os.RemoveAll(staged)
		}
	}()
	set, err := buildDevelopmentAuthorityDocuments(manifest, now)
	if err != nil {
		return err
	}
	defer set.destroy()
	for name, file := range map[string]struct {
		content []byte
		mode    os.FileMode
	}{
		developmentRootSeedFile:    {set.rootSeed, 0o600},
		developmentSignerSeedFile:  {set.signerSeed, 0o600},
		developmentTrustBundleFile: {set.trust, 0o400},
		developmentSnapshotFile:    {set.snapshot, 0o400},
	} {
		if err = localinstall.WriteExclusive(
			filepath.Join(staged, name), file.content, file.mode,
		); err != nil {
			return err
		}
	}
	if err = verifyDevelopmentAuthority(staged, manifest, now); err != nil {
		return err
	}
	if err = os.Rename(staged, current); err != nil {
		return fmt.Errorf("publish local development authority: %w", err)
	}
	published = true
	return nil
}

func readDevelopmentAuthorityFile(root, name string, maximum int64) ([]byte, error) {
	content, err := localinstall.ReadRegular(filepath.Join(root, name), maximum)
	if err != nil {
		return nil, errors.New("local development authority is invalid")
	}
	return content, nil
}
