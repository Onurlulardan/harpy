package postgres

import (
	"strings"
	"testing"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
)

func TestRetentionReleaseAuthorityIsTenantAndOwnerExact(t *testing.T) {
	t.Parallel()
	stamp := authorityFixture()
	tenant := authorityTenant
	value := jobsv1.RetentionRelease{
		TenantID: &tenant,
		Authority: jobsv1.Authority{
			ScopeKind: jobsv1.ScopeTenant, ScopeID: tenant,
			GroupID: "operations", GroupGeneration: 7,
			Domain: "workflow", CutoverGeneration: 5, WriterOrigin: jobsv1.WriterV2,
		},
	}
	if !stamp.matchesRetentionRelease(value) {
		t.Fatal("exact retention authority did not match")
	}
	otherTenant := "22222222-2222-4222-8222-222222222222"
	value.TenantID = &otherTenant
	if stamp.matchesRetentionRelease(value) {
		t.Fatal("cross-tenant retention release matched bound authority")
	}
	value.TenantID = &tenant
	value.Authority.Domain = "forms"
	if stamp.matchesRetentionRelease(value) {
		t.Fatal("cross-owner retention release matched bound authority")
	}
}

func TestRetentionReleaseSQLCallsOnlyTypedRoutine(t *testing.T) {
	t.Parallel()
	markers := []string{
		"SELECT platform.platform_async_release_job_retention_v1(",
		"$1::uuid", "$2::uuid", "$4::bigint", "$9::timestamptz",
	}
	for _, marker := range markers {
		if !strings.Contains(releaseJobRetentionSQL, marker) {
			t.Fatalf("retention release SQL lacks %q", marker)
		}
	}
	if strings.Contains(releaseJobRetentionSQL, "UPDATE platform.jobs") {
		t.Fatal("adapter bypasses the typed retention routine")
	}
}
