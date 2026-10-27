//go:build integration

package postgres

import (
	"context"
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/Onurlulardan/harpy/internal/attestation"
	"github.com/jackc/pgx/v5"
)

func seedTenantRequirementAuthority(
	t *testing.T, ctx context.Context, connection *pgx.Conn,
) {
	t.Helper()
	document, digest := tenantRequirementCatalogEvidence(t)
	_, err := connection.Exec(ctx, `
INSERT INTO platform.tenant_migration_requirement_catalogs (
    migration_run_id, contract_version, release_id, environment_id,
    database_origin, migration_manifest_sha256, catalog_sha256,
    requirement_count, target_count, canonical_document
) VALUES ($1,'tenant-migration-requirement-catalog.v2',$2,$3,'fresh',$4,$5,3,2,$6);
INSERT INTO platform.tenant_migration_requirements (
    migration_run_id, ordinal, owner, logical_version, database_origin,
    selected_path, selected_sha256
) VALUES
($1,1,'tenancy',20260714011000,'fresh',
 'core/internal/modules/tenancy/migrations/tenant/fresh/20260714011000_create_tenancy_schema_identity.fresh.sql',
 '3639feb1de870e0d71490f5530391fb952db381dadabd94574c53c7f113bf471'),
($1,2,'tenancy',20260714011100,'fresh',
 'core/internal/modules/tenancy/migrations/tenant/fresh/20260714011100_create_tenancy_schema_identity_guard.fresh.sql',
 '193a74b5dd839f9e6e2e9f8c24bc16fe0e2e190ab1fc0e78b034417ad47d4250'),
($1,3,'tenancy',20260714011200,'fresh',
 'core/internal/modules/tenancy/migrations/tenant/fresh/20260714011200_guard_tenancy_schema_identity.fresh.sql',
 '7c40e9e01ebdf6e9033f27f26eb04f82a93775db29f46d9f8760c6b9c352e479');
INSERT INTO platform.tenant_migration_requirement_targets (
    migration_run_id, ordinal, tenant_id, schema_name
) VALUES
($1,1,$7,'tenant_alpha'),
($1,2,$8,'tenant_beta');`,
		pgx.QueryExecModeSimpleProtocol, integrationReleaseRun, integrationReleaseID,
		integrationEnvironment, digestMarker('6'), digest, document,
		integrationTenantID, integrationTenantBID,
	)
	if err != nil {
		t.Fatalf("seed tenant requirement authority: %v", err)
	}
}

func tenantRequirementCatalogEvidence(t *testing.T) ([]byte, string) {
	t.Helper()
	requirements := []any{
		tenantRequirementEvidence(1, 20260714011000,
			"20260714011000_create_tenancy_schema_identity.fresh.sql",
			"3639feb1de870e0d71490f5530391fb952db381dadabd94574c53c7f113bf471"),
		tenantRequirementEvidence(2, 20260714011100,
			"20260714011100_create_tenancy_schema_identity_guard.fresh.sql",
			"193a74b5dd839f9e6e2e9f8c24bc16fe0e2e190ab1fc0e78b034417ad47d4250"),
		tenantRequirementEvidence(3, 20260714011200,
			"20260714011200_guard_tenancy_schema_identity.fresh.sql",
			"7c40e9e01ebdf6e9033f27f26eb04f82a93775db29f46d9f8760c6b9c352e479"),
	}
	targets := []any{
		map[string]any{"ordinal": 1, "tenant_id": integrationTenantID,
			"schema_name": "tenant_alpha"},
		map[string]any{"ordinal": 2, "tenant_id": integrationTenantBID,
			"schema_name": "tenant_beta"},
	}
	document, err := attestation.MarshalCanonicalTenantRequirementCatalog(map[string]any{
		"contract_version": "tenant-migration-requirement-catalog.v2",
		"migration_run_id": integrationReleaseRun, "release_id": integrationReleaseID,
		"environment_id": integrationEnvironment, "database_origin": "fresh",
		"migration_manifest_sha256": digestMarker('6'),
		"requirements":              requirements, "targets": targets,
	})
	if err != nil {
		t.Fatalf("canonicalize tenant requirement catalog: %v", err)
	}
	return document, integrationDocumentSHA256(document)
}

func tenantRequirementEvidence(ordinal int, version int64, name, digest string) map[string]any {
	return map[string]any{
		"ordinal": ordinal, "owner": "tenancy", "logical_version": version,
		"database_origin": "fresh",
		"selected_path":   "core/internal/modules/tenancy/migrations/tenant/fresh/" + name,
		"selected_sha256": digest,
	}
}

func integrationDocumentSHA256(document []byte) string {
	return fmt.Sprintf("%x", sha256.Sum256(document))
}
