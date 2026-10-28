//go:build integration

package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func readIntegrationEvidenceIdentity(t *testing.T, output string) existingEvidenceIdentity {
	t.Helper()
	authorityDocument, err := os.ReadFile(filepath.Join(output, "bootstrap-authority.json"))
	if err != nil {
		t.Fatal(err)
	}
	var authority runtimev1.BootstrapAuthorityManifest
	if err := attestation.DecodeStrict(authorityDocument, &authority); err != nil {
		t.Fatal(err)
	}
	authorityDigest, err := attestation.DigestHex(authorityDocument)
	if err != nil {
		t.Fatal(err)
	}
	bootstrapDocument, err := os.ReadFile(filepath.Join(output, "bootstrap-receipt.json"))
	if err != nil {
		t.Fatal(err)
	}
	var bootstrap runtimev1.BootstrapReceipt
	if err := attestation.DecodeStrict(bootstrapDocument, &bootstrap); err != nil {
		t.Fatal(err)
	}
	return existingEvidenceIdentity{
		environmentID: authority.EnvironmentID, releaseID: authority.ReleaseID,
		authorityDigest: authorityDigest,
		systemID:        bootstrap.Payload.DatabaseSystemIdentifier,
		databaseOID:     bootstrap.Payload.DatabaseOID,
	}
}

func mutateIntegrationEnvironmentRelease(t *testing.T, ctx context.Context, dsn string) {
	t.Helper()
	config, err := pgx.ParseConfig(dsn)
	if err != nil {
		t.Fatal(err)
	}
	config.Database = databaseName
	connection, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close(context.Background())
	if _, err := connection.Exec(ctx,
		"UPDATE platform.environment_identity SET release_id=$1", uuid.NewString(),
	); err != nil {
		t.Fatal(err)
	}
}
