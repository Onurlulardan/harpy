package runtimev1

import (
	_ "embed"
	"fmt"
)

// PRESERVE_NON_MIGRATION: embedded document names the Goose native history table.
//
//go:embed fresh-migration-manifest.json
var freshMigrationManifest []byte

func ReadEmbeddedFreshMigrationManifest() ([]byte, error) {
	if len(freshMigrationManifest) == 0 {
		return nil, fmt.Errorf("fresh migration manifest is not embedded")
	}
	return append([]byte(nil), freshMigrationManifest...), nil
}
