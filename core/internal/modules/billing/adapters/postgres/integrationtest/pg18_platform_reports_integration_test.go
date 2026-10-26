//go:build integration

package integrationtest

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
)

type platformReportFixture struct {
	globalAuthorityID uuid.UUID
	actorID           uuid.UUID
	a                 billingTenantQueryFixture
	b                 billingTenantQueryFixture
}

func TestBillingPlatformReportsAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := billingTestContext(t)
			defer cancel()
			database := startBillingPG18(t, ctx, "platform_reports_"+origin)
			prepareBillingPrerequisites(t, ctx, database, origin)
			applyBillingDocuments(
				t, ctx, database, platformReportDocuments(t, origin),
			)
			if _, err := database.ExecContext(ctx, `
GRANT harpy_runtime TO harpy_v2_runtime
WITH INHERIT FALSE, SET TRUE`); err != nil {
				t.Fatal(err)
			}
			connection := billingRuntimeConnection(t, ctx, database)
			globalAuthorityID := billingGlobalAuthorityID(t, ctx, database)
			a, b := preparePlatformReportTenants(
				t, ctx, connection, globalAuthorityID,
			)
			fixture := platformReportFixture{
				globalAuthorityID: globalAuthorityID,
				actorID:           billingV7(t), a: a, b: b,
			}
			seedPlatformReportAudit(t, ctx, connection, fixture, a)
			seedPlatformReportAudit(t, ctx, connection, fixture, b)
			assertPlatformRevenue(t, ctx, connection, fixture)
			assertPlatformReportViews(t, ctx, connection, fixture)
			assertPlatformReportAuthorityCuts(t, ctx, connection, fixture)
		})
	}
}

func platformReportDocuments(t *testing.T, origin string) []billingDocument {
	t.Helper()
	all := billingRESTQueryDocuments(t, origin)
	selected := make([]billingDocument, 0, len(all))
	for _, document := range all {
		if strings.Compare(filepath.Base(document.path), "20260717010229") < 0 {
			selected = append(selected, document)
		}
	}
	return selected
}
