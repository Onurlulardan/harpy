package migrationmanifest

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"

	"github.com/Onurlulardan/harpy/internal/attestation/migrationsupersession"
)

func resolveSupersessionCatalog(
	repositoryRoot string,
	sourceDirectory string,
) (string, fs.FileInfo, bool, error) {
	candidates := []string{
		path.Join(sourceDirectory, migrationsupersession.CatalogFileName),
		path.Join(path.Dir(sourceDirectory), migrationsupersession.CatalogFileName),
	}
	var selected string
	var selectedInfo fs.FileInfo
	for _, candidate := range candidates {
		absolute := filepath.Join(repositoryRoot, filepath.FromSlash(candidate))
		info, err := os.Lstat(absolute)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return "", nil, false, fmt.Errorf(
				"inspect migration supersession catalog: %w", err,
			)
		}
		if selected != "" {
			return "", nil, false, fmt.Errorf(
				"migration supersession catalog location is ambiguous",
			)
		}
		selected, selectedInfo = candidate, info
	}
	return selected, selectedInfo, selected != "", nil
}
