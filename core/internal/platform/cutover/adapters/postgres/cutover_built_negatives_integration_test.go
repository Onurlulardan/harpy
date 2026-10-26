//go:build integration

package postgres

import (
	"context"
	"net/url"
	"strings"
	"testing"
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
)

func TestBuiltCutoverAdoptionRejectsMismatchedAuthorityEvidence(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	database := startCutoverIntegrationDatabase(t, ctx)
	binary := buildCutoverMigratorBinary(t, ctx)
	wrongDatabaseDSN := prepareWrongBuiltCutoverDatabase(t, ctx, database)
	tests := []struct {
		name      string
		tamper    bool
		override  string
		role      string
		wrongSeal bool
	}{
		{name: "wrong artifact", tamper: true},
		{name: "wrong source identity", override: "HARPY_CUTOVER_SOURCE_DSN", role: "harpy_cutover_decoder"},
		{name: "wrong target identity", override: "HARPY_CUTOVER_TARGET_DSN", role: "harpy_adopter"},
		{name: "wrong session seal", wrongSeal: true},
	}
	for index, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			directory := t.TempDir()
			now := time.Now().UTC()
			signer := prepareBuiltCutoverSigner(t, directory, now)
			identity := builtCutoverIdentityFor(index+20, "completed")
			artifacts := buildBuiltCutoverLeaseArtifacts(
				t, ctx, database, signer, directory, identity, now,
			)
			seedBuiltCutoverOperation(t, ctx, database, artifacts)
			leaseBuiltCutoverRoles(t, ctx, database, artifacts.leaseExpiresAt(t))
			t.Cleanup(func() { cleanupBuiltAdoptionRoles(t, ctx, database) })
			assertBuiltAdoptionLeaseCatalog(t, ctx, database, artifacts)
			environment := builtAdoptionEnvironment(t, database, signer, artifacts)
			if test.tamper {
				tampered := tamperBuiltCutoverLease(t, directory, artifacts)
				environment = overrideBuiltCutoverEnvironment(
					environment, "HARPY_CUTOVER_LEASE_MANIFEST_PATH", tampered.path,
				)
			}
			if test.override != "" {
				dsn := builtCutoverRoleDSN(t, wrongDatabaseDSN, test.role, builtCutoverPassword)
				environment = overrideBuiltCutoverEnvironment(environment, test.override, dsn)
			}
			process := startBuiltCutoverProcess(
				t, ctx, binary, "cutover_adoption", environment,
			)
			if !test.tamper && test.override == "" {
				var mutate func(*cutoverv1.SessionSealReceipt)
				if test.wrongSeal {
					mutate = func(receipt *cutoverv1.SessionSealReceipt) {
						receipt.Payload.SourceSessions[0].PID += 1_000_000
					}
				}
				publishBuiltSessionSealWithMutation(
					t, ctx, database, signer, artifacts, process, mutate,
				)
			}
			outcome := waitBuiltCutoverProcess(t, process, 20*time.Second)
			assertBuiltCutoverPreflightFailure(t, outcome)
		})
	}
}

func prepareWrongBuiltCutoverDatabase(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
) string {
	t.Helper()
	const name = "harpy_cutover_wrong"
	if _, err := database.admin.Exec(ctx, "CREATE DATABASE "+name+" OWNER harpy_owner"); err != nil {
		t.Fatalf("create wrong cutover database: %v", err)
	}
	endpoint, err := url.Parse(database.dsn)
	if err != nil {
		t.Fatal(err)
	}
	endpoint.Path = "/" + name
	return endpoint.String()
}

func overrideBuiltCutoverEnvironment(values []string, key, value string) []string {
	result := append([]string(nil), values...)
	prefix := key + "="
	for index := range result {
		if strings.HasPrefix(result[index], prefix) {
			result[index] = prefix + value
			return result
		}
	}
	return append(result, prefix+value)
}

func tamperBuiltCutoverLease(
	t *testing.T, directory string, artifacts builtCutoverLeaseArtifacts,
) builtCutoverArtifact {
	t.Helper()
	lease := artifacts.lease
	lease.PlanSHA256 = digestMarker('d')
	return writeBuiltCutoverArtifact(t, directory, "tampered-lease.json", lease)
}

func (artifacts builtCutoverLeaseArtifacts) leaseExpiresAt(t *testing.T) time.Time {
	t.Helper()
	value, err := time.Parse(time.RFC3339Nano, artifacts.lease.ExpiresAt)
	if err != nil {
		t.Fatal(err)
	}
	return value
}

func assertBuiltAdoptionLeaseCatalog(
	t *testing.T, ctx context.Context, database cutoverIntegrationDatabase,
	artifacts builtCutoverLeaseArtifacts,
) {
	t.Helper()
	expiresAt := artifacts.leaseExpiresAt(t)
	assertBuiltLeaseCatalog(t, ctx, database.admin, "harpy_cutover_decoder", expiresAt, true, 2)
	assertBuiltLeaseCatalog(t, ctx, database.admin, "harpy_adopter", expiresAt, false, 1)
	assertBuiltBoundedRoleSessions(
		t, ctx, database, "harpy_cutover_decoder", expiresAt, true, 2,
	)
	assertBuiltBoundedRoleSessions(
		t, ctx, database, "harpy_adopter", expiresAt, false, 1,
	)
}
