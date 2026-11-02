package main

import (
	"fmt"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

func schemaMigrationManifest(string) ([]byte, error) {
	document, err := runtimev1.ReadEmbeddedFreshMigrationManifest()
	if err != nil {
		return nil, fmt.Errorf("read embedded fresh migration manifest: %w", err)
	}
	return document, nil
}
