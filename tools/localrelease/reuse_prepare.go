package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

const reusePrepareValidityFloor = 2 * time.Minute

func reuseEvidenceClock(now time.Time) time.Time {
	return now.Add(reusePrepareValidityFloor)
}

func existingOwnedOutput(path string) (bool, error) {
	if _, err := os.Lstat(path); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("inspect release output: %w", err)
	}
	if err := verifyOwnedDirectory(path, filepath.Dir(path)); err != nil {
		return false, err
	}
	return true, nil
}

func reuseExistingPrepare(
	ctx context.Context, options prepareOptions, contract runtimev1.BinarySchemaContract,
	migration artifact, adminDSN string, passwords applicationPasswords,
	packageAuthority catalogPackageAuthority,
) error {
	catalog, err := loadExistingReleaseCatalog(options.outputDirectory, migration, contract)
	if err != nil {
		return fmt.Errorf("validate existing release catalog: %w", err)
	}
	now := reuseEvidenceClock(time.Now().UTC().Truncate(time.Second))
	identity, err := loadExistingEvidence(
		options.outputDirectory, now, catalog, packageAuthority,
	)
	if err != nil {
		return err
	}
	operationContext, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()
	if err := verifyReuseDatabase(
		operationContext, adminDSN, passwords, identity,
		migration.digest == foundationMigrationDigest,
	); err != nil {
		return fmt.Errorf("validate existing release database binding: %w", err)
	}
	fmt.Printf("local release reused: %s\nrelease_id: %s\n",
		options.outputDirectory, identity.releaseID)
	return nil
}
