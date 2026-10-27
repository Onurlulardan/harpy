//go:build integration

package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/Onurlulardan/harpy/internal/attestation"
	"github.com/jackc/pgx/v5/pgconn"
)

func buildCutoverAuthorityBinary(t *testing.T, ctx context.Context) string {
	t.Helper()
	name := "cutoverauthority"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	target := filepath.Join(t.TempDir(), name)
	command := exec.CommandContext(ctx, "go", "build", "-o", target, "./ops/cutoverauthority")
	command.Dir = cutoverRepositoryRoot(t)
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("build external cutover authority: %v: %s", err, output)
	}
	return target
}

func adoptBuiltCutoverAuthority(
	t *testing.T, ctx context.Context, binary string, database cutoverIntegrationDatabase,
	signer builtCutoverSigner, artifacts builtCutoverAuthorityArtifacts, directory string,
) builtCutoverAuthorityArtifacts {
	t.Helper()
	dsnPath := writeBuiltSecret(t, directory, "root.dsn", database.dsn)
	config := writeBuiltCutoverArtifact(t, directory, "authority-adopt-config.json", map[string]any{
		"contract_version": "external-cutover-authority-config.v1",
		"action":           "adopt_authority", "environment": "test", "deadline_seconds": 120,
		"adopt": map[string]any{
			"root_dsn_file": dsnPath, "trust_root": signer.rootPath,
			"release_manifest":              artifacts.releaseArtifact.path,
			"release_signature":             artifacts.releaseSignature.path,
			"bootstrap_authority_manifest":  artifacts.bootstrapAuthority.path,
			"bootstrap_authority_signature": artifacts.bootstrapSignature.path,
			"authority_operation":           artifacts.operationArtifact.path,
			"authority_operation_signature": artifacts.operationSignature.path,
			"signing_key_file":              signer.signingKeyPath,
			"receipt_path":                  artifacts.receiptArtifact.path,
			"receipt_signature_path":        artifacts.receiptSignature.path,
		},
	})
	first := runBuiltAuthorityCommand(t, ctx, binary, config.path)
	assertBuiltAuthorityReplay(t, first, false)
	stored := readBuiltArtifact(t, artifacts.receiptArtifact.path)
	assertBuiltAdoptedCatalog(t, ctx, database, stored.digest)
	second := runBuiltAuthorityCommand(t, ctx, binary, config.path)
	assertBuiltAuthorityReplay(t, second, true)
	artifacts.receiptArtifact = readBuiltArtifact(t, artifacts.receiptArtifact.path)
	artifacts.receiptSignature = readBuiltArtifact(t, artifacts.receiptSignature.path)
	return artifacts
}

func assertBuiltAdoptedCatalog(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase, digest string,
) {
	t.Helper()
	var tableOwner, functionOwner string
	var securityDefiner, ownerCanSelect bool
	err := database.admin.QueryRow(ctx, `SELECT
pg_catalog.pg_get_userbyid(relation.relowner),
pg_catalog.pg_get_userbyid(procedure_row.proowner),procedure_row.prosecdef,
pg_catalog.has_table_privilege('harpy_cutover_authority_owner',
 'cutover_authority.authority_adoption_inventory','SELECT')
FROM pg_catalog.pg_class AS relation
JOIN pg_catalog.pg_namespace AS namespace ON namespace.oid=relation.relnamespace
CROSS JOIN pg_catalog.pg_proc AS procedure_row
JOIN pg_catalog.pg_namespace AS procedure_namespace
  ON procedure_namespace.oid=procedure_row.pronamespace
WHERE namespace.nspname='cutover_authority'
  AND relation.relname='authority_adoption_inventory'
  AND procedure_namespace.nspname='cutover_authority'
  AND procedure_row.proname='assert_adopted_catalog_v1'`).Scan(
		&tableOwner, &functionOwner, &securityDefiner, &ownerCanSelect,
	)
	if err != nil {
		t.Fatalf("inspect stored authority catalog: %v", err)
	}
	_, err = database.admin.Exec(ctx,
		"SELECT cutover_authority.assert_adopted_catalog_v1($1::text)", digest)
	if err != nil {
		var postgresError *pgconn.PgError
		_ = errors.As(err, &postgresError)
		t.Fatalf("assert stored authority catalog: %v where=%q detail=%q internal=%q table_owner=%s function_owner=%s security_definer=%t owner_select=%t",
			err, postgresError.Where, postgresError.Detail, postgresError.InternalQuery,
			tableOwner, functionOwner, securityDefiner, ownerCanSelect)
	}
}

func runBuiltAuthorityCommand(
	t *testing.T, ctx context.Context, binary, configPath string,
) []byte {
	t.Helper()
	output, err := runBuiltAuthorityProcess(ctx, binary, configPath)
	if err != nil {
		t.Fatalf("run external cutover authority: %v: %s", err, output)
	}
	return output
}

func runBuiltAuthorityProcess(
	ctx context.Context, binary, configPath string,
) ([]byte, error) {
	command := exec.CommandContext(ctx, binary, "--config", configPath)
	return command.CombinedOutput()
}

func assertBuiltAuthorityReplay(t *testing.T, document []byte, expected bool) {
	t.Helper()
	var result struct {
		Action   string `json:"action"`
		Replayed bool   `json:"replayed"`
	}
	if err := json.Unmarshal(document, &result); err != nil ||
		result.Action != "adopt_authority" || result.Replayed != expected {
		t.Fatalf("authority adoption replay result = %s err=%v", document, err)
	}
}

func writeBuiltSecret(t *testing.T, directory, name, value string) string {
	t.Helper()
	path := filepath.Join(directory, name)
	if err := os.WriteFile(path, []byte(value), 0o600); err != nil {
		t.Fatalf("write built secret %s: %v", name, err)
	}
	return path
}

func readBuiltArtifact(t *testing.T, path string) builtCutoverArtifact {
	t.Helper()
	document, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read built artifact %s: %v", path, err)
	}
	digest, err := attestation.DigestHex(document)
	if err != nil {
		t.Fatal(err)
	}
	return builtCutoverArtifact{path: path, document: document, digest: digest}
}
