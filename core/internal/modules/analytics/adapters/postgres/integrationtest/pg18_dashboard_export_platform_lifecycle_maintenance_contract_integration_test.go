//go:build integration

package integrationtest

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	"github.com/google/uuid"
)

const (
	analyticsDeleteEffectNamespace = "bfce9d42-fd8b-453a-a54e-d6f35a886868"
	analyticsExpireEffectNamespace = "1bce05e3-fc4d-40a9-82e9-f125b12fa10a"
)

func analyticsDashboardExportMaintenanceRequest(t *testing.T, exportID, jobID,
	attemptID, kind string, requestedAt time.Time,
) (effectsv1.Request, string) {
	t.Helper()
	operation, namespace := analyticsDashboardExportMaintenanceOperation(t, kind)
	attempt := uuid.MustParse(attemptID)
	effectID := uuid.NewSHA1(namespace, attempt[:])
	payloadDigest := sha256.Sum256([]byte(exportID))
	tenantID := analyticsTenantA
	request, err := effectsv1.NewRequest(effectID.String(), kind, 1, &tenantID,
		effectsv1.Authority{ScopeKind: effectsv1.ScopeTenant,
			ScopeID: analyticsTenantA, GroupID: "analytics", GroupGeneration: 1,
			Domain: "analytics", CutoverGeneration: 1, WriterOrigin: effectsv1.WriterV2},
		"analytics.dashboard-export", operation, kind+":"+exportID, jobID, jobID,
		exportID, hex.EncodeToString(payloadDigest[:]),
		requestedAt.UTC().Format(time.RFC3339Nano))
	if err != nil {
		t.Fatal(err)
	}
	document, err := request.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	digest, err := canonicaljsonv1.DigestHex(document, effectsv1.MaximumDocumentBytes)
	if err != nil {
		t.Fatal(err)
	}
	return request, digest
}

func analyticsDashboardExportMaintenanceOperation(t *testing.T,
	kind string,
) (string, uuid.UUID) {
	t.Helper()
	switch kind {
	case "analytics.dashboard.export.delete":
		return "delete", uuid.MustParse(analyticsDeleteEffectNamespace)
	case "analytics.dashboard.export.expire":
		return "expire", uuid.MustParse(analyticsExpireEffectNamespace)
	default:
		t.Fatalf("unexpected Analytics maintenance kind %q", kind)
		return "", uuid.Nil
	}
}

func analyticsDashboardExportMaintenanceEffectID(t *testing.T, kind,
	attemptID string,
) string {
	t.Helper()
	_, namespace := analyticsDashboardExportMaintenanceOperation(t, kind)
	attempt := uuid.MustParse(attemptID)
	return uuid.NewSHA1(namespace, attempt[:]).String()
}

func assertAnalyticsDashboardExportMaintenancePlannerSource(t *testing.T) {
	t.Helper()
	source := loadAnalyticsDashboardExportProductionSource(t,
		"modules/analytics/dashboard_export_job_planner.go")
	for _, expected := range []string{analyticsDeleteEffectNamespace,
		analyticsExpireEffectNamespace, `dashboardExportEffectProvider = "analytics.dashboard-export"`,
		"uuid.NewSHA1(planner.namespace, attemptID[:])"} {
		if !strings.Contains(source, expected) {
			t.Fatalf("Analytics maintenance planner source lost %q", expected)
		}
	}
}
