//go:build integration

package migrator

import (
	"context"
	"testing"
	"time"
)

type builtEntrypointDeployEvidence struct {
	releaseManifest    string
	releaseSignature   string
	trustRoot          string
	bootstrapReceipt   string
	extensionReceipt   string
	cleanupReceipt     string
	authorityManifest  string
	authoritySignature string
	catalog            builtEntrypointArtifactCatalog
}

func prepareBuiltEntrypointDeployEvidence(
	t *testing.T,
	ctx context.Context,
	database builtEntrypointDatabase,
	stage builtEntrypointStage,
) builtEntrypointDeployEvidence {
	t.Helper()
	directory := t.TempDir()
	now := time.Now().UTC().Truncate(time.Second)
	live := readBuiltEntrypointLiveIdentity(t, ctx, database)
	catalog := prepareBuiltEntrypointArtifactCatalog(t, directory, stage)
	signer := prepareBuiltEntrypointSigner(t, directory, now)
	authority, authoritySignature := writeBuiltEntrypointAuthority(
		t, directory, now, catalog, signer,
	)
	catalog.cutoverAuthorityOperation, catalog.cutoverAuthorityOperationSignature =
		writeBuiltEntrypointCutoverAuthorityOperation(
			t, directory, now, catalog, authority, signer,
		)
	release, releaseSignature := writeBuiltEntrypointRelease(
		t, directory, now, catalog, authority, signer,
	)
	bootstrap := writeBuiltEntrypointBootstrapReceipt(
		t, directory, now, live, catalog, authority, release, signer,
	)
	extension := writeBuiltEntrypointExtensionReceipt(
		t, directory, now, live, catalog, authority, release, signer,
	)
	cleanup := writeBuiltEntrypointCleanupReceipt(
		t, directory, now, live, catalog, authority, release, signer,
	)
	return builtEntrypointDeployEvidence{
		releaseManifest: release.path, releaseSignature: releaseSignature.path,
		trustRoot:        signer.root.path,
		bootstrapReceipt: bootstrap.path, extensionReceipt: extension.path,
		cleanupReceipt:    cleanup.path,
		authorityManifest: authority.path, authoritySignature: authoritySignature.path,
		catalog: catalog,
	}
}

func builtEntrypointDeployEnvironment(
	database builtEntrypointDatabase,
	evidence builtEntrypointDeployEvidence,
) []string {
	return []string{
		"HARPY_ENVIRONMENT=test",
		"HARPY_POSTGRES_ENDPOINT_KIND=direct_postgres",
		"HARPY_POSTGRES_DSN=" + database.migratorDSN,
		"HARPY_POSTGRES_MAX_CONNS=4", "HARPY_POSTGRES_MIN_CONNS=0",
		"HARPY_POSTGRES_CONNECT_TIMEOUT=5s", "HARPY_POSTGRES_CLEANUP_TIMEOUT=5s",
		"HARPY_POSTGRES_HEALTH_CHECK_PERIOD=5s", "HARPY_READINESS_TIMEOUT=5s",
		"HARPY_LOG_LEVEL=error",
		"HARPY_RELEASE_MANIFEST_PATH=" + evidence.releaseManifest,
		"HARPY_RELEASE_SIGNATURE_PATH=" + evidence.releaseSignature,
		"HARPY_RELEASE_TRUST_ROOT_PATH=" + evidence.trustRoot,
		"HARPY_BOOTSTRAP_RECEIPT_PATH=" + evidence.bootstrapReceipt,
		"HARPY_EXTENSION_RECEIPT_PATH=" + evidence.extensionReceipt,
		"HARPY_CREDENTIAL_CLEANUP_RECEIPT_PATH=" + evidence.cleanupReceipt,
		"HARPY_CATALOG_PACKAGE_SET_PATH=" + evidence.catalog.catalogPackageSet.path,
		"HARPY_CATALOG_PACKAGE_PUBLIC_KEYRING=" +
			evidence.catalog.catalogPackageKeyring,
		"HARPY_BOOTSTRAP_AUTHORITY_MANIFEST_PATH=" + evidence.authorityManifest,
		"HARPY_BOOTSTRAP_AUTHORITY_SIGNATURE_PATH=" + evidence.authoritySignature,
		"HARPY_MIGRATION_MANIFEST_PATH=" + evidence.catalog.migration.path,
		"HARPY_EXTENSION_MANIFEST_PATH=" + evidence.catalog.extension.path,
		"HARPY_ROLES_MANIFEST_PATH=" + evidence.catalog.roles.path,
		"HARPY_PRIVILEGES_MANIFEST_PATH=" + evidence.catalog.privileges.path,
		"HARPY_OWNERSHIP_MANIFEST_PATH=" + evidence.catalog.ownership.path,
		"HARPY_BOOTSTRAP_OPERATION_PATH=" + evidence.catalog.bootstrapOperation.path,
		"HARPY_EXTENSION_OPERATION_PATH=" + evidence.catalog.extensionOperation.path,
		"HARPY_CREDENTIAL_CLEANUP_OPERATION_PATH=" + evidence.catalog.cleanupOperation.path,
		"HARPY_CUTOVER_AUTHORITY_OPERATION_PATH=" +
			evidence.catalog.cutoverAuthorityOperation.path,
		"HARPY_CUTOVER_AUTHORITY_OPERATION_SIGNATURE_PATH=" +
			evidence.catalog.cutoverAuthorityOperationSignature.path,
	}
}
