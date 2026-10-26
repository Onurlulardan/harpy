package serviceprincipal

import (
	"strings"
	"testing"
)

func TestPublicationAuthoritySQLLocksAndResolvesExactIAMOwnership(t *testing.T) {
	for name, document := range map[string]string{
		"principal": publicationAuthorityPrincipalSQL,
		"grant":     publicationAuthorityGrantSQL,
	} {
		if !strings.Contains(document, "pg_catalog.pg_advisory_xact_lock") ||
			!strings.Contains(document, "harpy/iam/workflow-publication-") ||
			!strings.Contains(document, "authority_lock AS MATERIALIZED") ||
			!strings.Contains(document, "FROM authority_lock") {
			t.Fatalf("%s authority lock is not transaction-scoped", name)
		}
	}
	for _, marker := range []string{
		"permission.match_kind = 'exact'",
		"permission.is_protected",
		"resource.resource_key = 'iam.capability_grant'",
		"resource.owner_domain = 'iam'",
		"resource.resource_kind = 'capability_grant'",
		"action.action_key = 'manage'",
	} {
		if !strings.Contains(publicationAuthorityPermissionSQL, marker) {
			t.Fatalf("authority permission query misses %q", marker)
		}
	}
	for _, marker := range []string{
		"iam_capability_grants", "iam_capability_grant_revisions",
		"head.current_revision",
	} {
		if !strings.Contains(publicationAuthorityGrantSQL, marker) {
			t.Fatalf("authority grant query misses %q", marker)
		}
	}
}
