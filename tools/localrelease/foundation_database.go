package main

import (
	"fmt"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

func verifyNewReleaseStage(
	contract runtimev1.BinarySchemaContract,
	environmentID string,
) error {
	if contract.MigrationManifestSHA256 == foundationMigrationDigest && environmentID != "" {
		return fmt.Errorf("bootstrap baseline already exists; its evidence cannot be replaced")
	}
	return nil
}
