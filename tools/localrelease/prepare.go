package main

import (
	"context"
	"fmt"
	"time"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

func runPrepare(ctx context.Context, arguments []string) error {
	options, err := parsePrepareOptions(arguments)
	if err != nil {
		return err
	}
	contract, err := readEmbeddedBuildContract()
	if err != nil {
		return err
	}
	return executePrepare(ctx, options, contract)
}

func executePrepare(
	ctx context.Context, options prepareOptions, contract runtimev1.BinarySchemaContract,
) error {
	adminDSN, err := readSecret(options.adminDSNFile, "admin DSN")
	if err != nil {
		return err
	}
	passwords, err := readApplicationPasswords(options)
	if err != nil {
		return err
	}
	packageAuthority, err := loadCatalogPackageAuthority(options)
	if err != nil {
		return err
	}
	defer clear(packageAuthority.privateKey)
	migration, err := loadPreparedMigration(options.migrationManifest, contract)
	if err != nil {
		return err
	}
	exists, err := existingOwnedOutput(options.outputDirectory)
	if err != nil {
		return err
	}
	if exists && !options.replaceOwned {
		return reuseExistingPrepare(
			ctx, options, contract, migration, adminDSN, passwords,
			packageAuthority,
		)
	}
	if exists && reuseExistingPrepare(
		ctx, options, contract, migration, adminDSN, passwords,
		packageAuthority,
	) == nil {
		return nil
	}
	// A reused release that fails verification must be rebuilt in full.
	// refreshSignedRelease replaces only the trust root and release signature,
	// leaving authority and receipts on the previous key.
	agentDebugLog("A", "prepare.go:executePrepare", "full release build", map[string]string{
		"existingOutput": fmt.Sprintf("%t", exists),
	})
	operationContext, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()
	catalog, err := buildReleaseCatalog(migration)
	if err != nil {
		return err
	}
	live, err := preparePostgres(
		operationContext, adminDSN, passwords, contract,
	)
	if err != nil {
		return err
	}
	now := time.Now().UTC().Truncate(time.Second)
	ids := applyStoredReleaseID(
		newReleaseIDsForEnvironment(live.environmentID), live.appliedReleaseID,
	)
	agentDebugLog("H", "prepare.go:executePrepare", "release identity", map[string]string{
		"appliedRelease": fmt.Sprintf("%t", live.appliedReleaseID != ""),
	})
	catalog.catalogPackageSet, err = buildCatalogPackageSet(
		ids.release, packageAuthority,
	)
	if err != nil {
		return err
	}
	signer, err := newSigningAuthority(now)
	if err != nil {
		return err
	}
	authority, err := buildBootstrapAuthority(now, ids, catalog, signer)
	if err != nil {
		return err
	}
	cutoverOperation, operation, err := buildCutoverOperation(
		now, ids, live, catalog, authority, signer,
	)
	if err != nil {
		return err
	}
	release, _, err := buildReleaseManifest(
		now, ids, catalog, authority, cutoverOperation, signer,
	)
	if err != nil {
		return err
	}
	evidence, err := buildAllReceipts(
		now, ids, live, catalog, authority, cutoverOperation, operation, release, signer,
	)
	if err != nil {
		return err
	}
	if err := verifyGeneratedEvidence(
		now, catalog, signer, packageAuthority, evidence,
	); err != nil {
		return err
	}
	if err := publishGeneratedEvidence(
		options.outputDirectory, catalog, signer, evidence, exists,
	); err != nil {
		return err
	}
	fmt.Printf("local release: %s\nrelease_id: %s\nmigration_sha256: %s\nextension_sha256: %s\n",
		options.outputDirectory, ids.release, catalog.migration.digest, catalog.extension.digest)
	return nil
}
