//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"
)

const (
	retentionReleaseDeadline = "2025-10-01 00:00:00+00"
	retentionReleaseDigest   = "5fd2c65bfa1954fad42e5b6c6e3d97a672570318b9b273cf3aaaca32dca8f954"
)

func assertRetentionReleaseCases(
	t *testing.T,
	ctx context.Context,
	connection *sql.Conn,
) {
	t.Helper()
	if released := callRetentionRelease(
		t, ctx, connection, retentionSuccessJob, asyncTenantA,
		"operations", "workflow", 0, retentionReleaseDeadline,
	); !released {
		t.Fatal("valid retention release was rejected")
	}
	if released := callRetentionRelease(
		t, ctx, connection, retentionSuccessJob, asyncTenantA,
		"operations", "workflow", 0, retentionReleaseDeadline,
	); released {
		t.Fatal("stale retention revision was accepted")
	}
	if released := callRetentionRelease(
		t, ctx, connection, retentionSuccessJob, asyncTenantA,
		"operations", "workflow", 1, retentionReleaseDeadline,
	); !released {
		t.Fatal("next monotonic retention revision was rejected")
	}
	assertRetentionReleaseRow(t, ctx, connection)

	if released := callRetentionRelease(
		t, ctx, connection, retentionTenantBJob, asyncTenantA,
		"operations", "workflow", 0, retentionReleaseDeadline,
	); released {
		t.Fatal("tenant A released tenant B job retention")
	}
	var tenantBRevision int64
	if err := connection.QueryRowContext(ctx, `
SELECT retention_release_revision FROM platform.jobs WHERE id = $1`,
		retentionTenantBJob).Scan(&tenantBRevision); err != nil || tenantBRevision != 0 {
		t.Fatalf("tenant B release revision = %d, error %v", tenantBRevision, err)
	}

	for name, jobID := range map[string]string{
		"retry child":         retentionRetryJob,
		"unfinished attempt":  retentionAttemptJob,
		"live dependency":     retentionDependencyJob,
		"live effect receipt": retentionEffectJob,
	} {
		t.Run(name, func(t *testing.T) {
			if released := callRetentionRelease(
				t, ctx, connection, jobID, asyncTenantA,
				"operations", "workflow", 0, retentionReleaseDeadline,
			); released {
				t.Fatal("retention release ignored a live reference")
			}
		})
	}

	if released := callRetentionRelease(
		t, ctx, connection, retentionRetryJob, asyncTenantA,
		"communications", "notifications", 0, retentionReleaseDeadline,
	); released {
		t.Fatal("different owner released job retention")
	}
}

func callRetentionRelease(
	t *testing.T,
	ctx context.Context,
	connection *sql.Conn,
	jobID string,
	scopeID string,
	groupID string,
	domain string,
	expectedRevision int64,
	retainUntil string,
) bool {
	t.Helper()
	var released bool
	err := connection.QueryRowContext(ctx, releaseJobRetentionSQL,
		jobID, scopeID, groupID, int64(1), domain, int64(1),
		expectedRevision, retentionReleaseDigest, retainUntil,
	).Scan(&released)
	if err != nil {
		t.Fatalf("call retention release: %v", err)
	}
	return released
}

func assertRetentionReleaseRow(t *testing.T, ctx context.Context, connection *sql.Conn) {
	t.Helper()
	var revision int64
	var digest string
	var releasedAt sql.NullTime
	err := connection.QueryRowContext(ctx, `
SELECT retention_release_revision, retention_release_sha256, retention_released_at
FROM platform.jobs WHERE id = $1`, retentionSuccessJob).
		Scan(&revision, &digest, &releasedAt)
	if err != nil || revision != 2 || digest != retentionReleaseDigest || !releasedAt.Valid {
		t.Fatalf("released row = revision %d digest %s time %v error %v",
			revision, digest, releasedAt.Valid, err)
	}
}
