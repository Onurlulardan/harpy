package licenseauthority

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

type runtimeStoreFixture struct {
	grants []RuntimeGrant
	err    error
}

func (store runtimeStoreFixture) LoadLicenseRuntimeEntitlements(
	context.Context,
	uuid.UUID,
) ([]RuntimeGrant, error) {
	return append([]RuntimeGrant(nil), store.grants...), store.err
}

func TestReadCarriesDatabaseDecisionWithoutRecomputing(t *testing.T) {
	environmentID := uuid.MustParse("01900000-0000-7000-8000-000000000401")
	grants := runtimeGrantFixture(environmentID, RuntimeBuildOff)
	reader, err := NewRuntimeReader(runtimeStoreFixture{grants: grants})
	if err != nil {
		t.Fatal(err)
	}
	result, err := reader.ReadLicenseRuntimeEntitlements(
		t.Context(), environmentID.String(),
	)
	if err != nil || result.State != RuntimeBuildOff ||
		result.Projection.Boolean["workflow.enabled"] != true ||
		result.Projection.Boolean["api.builder.enabled"] != false ||
		result.Projection.Limits["iam.active_users"].Value != 0 ||
		result.Projection.Limits["iam.active_users"].Unlimited {
		t.Fatalf("runtime projection=%+v err=%v", result, err)
	}
}

func TestRuntimeReaderRejectsMalformedOrUnavailableAuthority(t *testing.T) {
	environmentID := uuid.MustParse("01900000-0000-7000-8000-000000000402")
	grants := runtimeGrantFixture(environmentID, RuntimeValid)
	grants[1].EntitlementKey = grants[0].EntitlementKey
	reader, _ := NewRuntimeReader(runtimeStoreFixture{grants: grants})
	if _, err := reader.ReadLicenseRuntimeEntitlements(
		t.Context(), environmentID.String(),
	); !errors.Is(err, ErrAuthorityState) {
		t.Fatalf("malformed runtime error=%v", err)
	}
	reader, _ = NewRuntimeReader(runtimeStoreFixture{err: errors.New("database unavailable")})
	if _, err := reader.ReadLicenseRuntimeEntitlements(
		t.Context(), environmentID.String(),
	); err == nil || errors.Is(err, ErrAuthorityState) {
		t.Fatalf("runtime dependency error=%v", err)
	}
}

func runtimeGrantFixture(
	environmentID uuid.UUID,
	state RuntimeState,
) []RuntimeGrant {
	licenseID := uuid.New()
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	booleans := []string{"ai.chat", "analytics.dashboard_builder.enabled", "api.access",
		"api.builder.enabled", "forms.builder.enabled", "workflow.builder.enabled", "workflow.enabled"}
	limits := []string{"analytics.active_dashboards", "api.active_endpoints",
		"documents.storage_bytes", "forms.active_forms", "forms.submissions",
		"iam.active_users", "workflow.active_definitions"}
	grants := make([]RuntimeGrant, 0, 14)
	for _, key := range booleans {
		value := key == "workflow.enabled" || key == "api.access"
		grants = append(grants, runtimeGrantHead(
			licenseID, environmentID, state, now, key, "boolean", &value, nil, false,
		))
	}
	for _, key := range limits {
		value := int64(0)
		grants = append(grants, runtimeGrantHead(
			licenseID, environmentID, state, now, key, "integer", nil, &value, false,
		))
	}
	return grants
}

func runtimeGrantHead(
	licenseID, environmentID uuid.UUID,
	state RuntimeState,
	now time.Time,
	key, valueType string,
	booleanValue *bool,
	integerValue *int64,
	unlimited bool,
) RuntimeGrant {
	return RuntimeGrant{
		LicenseID: licenseID, EnvironmentID: environmentID,
		SnapshotSequence: 9, SnapshotSHA256: strings.Repeat("a", 64),
		State: state, EntitlementKey: key, ValueType: valueType,
		BooleanValue: booleanValue, IntegerValue: integerValue,
		Unlimited: unlimited, Channel: "stable", DatabaseNow: now,
	}
}
