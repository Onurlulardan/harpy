package migrationmanifest

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"

	"github.com/Onurlulardan/harpy/tools/internal/storagecatalog"
)

func LoadStorageInventory(repositoryRoot string) (StorageInventory, error) {
	if !filepath.IsAbs(repositoryRoot) {
		return StorageInventory{}, fmt.Errorf("storage inventory root must be absolute")
	}
	manifest, err := storagecatalog.Load(repositoryRoot)
	if err != nil {
		return StorageInventory{}, fmt.Errorf("load storage ownership contract: %w", err)
	}
	claims, err := migrationClaims(manifest)
	if err != nil {
		return StorageInventory{}, err
	}
	streams, err := loadMigrationStreams(repositoryRoot, claims)
	if err != nil {
		return StorageInventory{}, err
	}
	return StorageInventory{claims: claims, streams: streams}, nil
}

func migrationClaims(manifest storagecatalog.Manifest) ([]StorageClaim, error) {
	seen := map[string]bool{}
	var claims []StorageClaim
	for _, owner := range manifest.Owners {
		for _, directory := range owner.MigrationDirs {
			if !fs.ValidPath(directory.Path) {
				return nil, fmt.Errorf(
					"storage migration directory %q is invalid", directory.Path,
				)
			}
			if directory.Scope != "platform" && directory.Scope != "tenant" {
				return nil, fmt.Errorf(
					"storage migration directory %q has invalid scope %q",
					directory.Path, directory.Scope,
				)
			}
			if seen[directory.Path] {
				return nil, fmt.Errorf(
					"storage migration directory %q is claimed more than once",
					directory.Path,
				)
			}
			seen[directory.Path] = true
			claims = append(claims, StorageClaim{
				Owner: owner.Name, Directory: directory.Path, Scope: directory.Scope,
			})
		}
	}
	if len(claims) == 0 {
		return nil, fmt.Errorf("storage migration claims are empty")
	}
	sort.Slice(claims, func(left, right int) bool {
		return claims[left].Directory < claims[right].Directory
	})
	return claims, nil
}
