package licenseauthority

import (
	"context"
	"encoding/hex"
	"fmt"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
	"github.com/google/uuid"
)

type RuntimeReader struct{ store RuntimeStore }

var _ billingapi.LicenseRuntimeReader = (*RuntimeReader)(nil)

func NewRuntimeReader(store RuntimeStore) (*RuntimeReader, error) {
	if store == nil {
		return nil, fmt.Errorf("billing license runtime store is required")
	}
	return &RuntimeReader{store: store}, nil
}

func (reader *RuntimeReader) ReadLicenseRuntimeEntitlements(
	ctx context.Context,
	environment string,
) (billingapi.LicenseRuntimeEntitlements, error) {
	environmentID, err := uuid.Parse(environment)
	if err != nil || environmentID.Version() != 7 || ctx == nil || ctx.Err() != nil {
		return billingapi.LicenseRuntimeEntitlements{}, ErrAuthorityState
	}
	grants, err := reader.store.LoadLicenseRuntimeEntitlements(ctx, environmentID)
	if err != nil {
		return billingapi.LicenseRuntimeEntitlements{}, fmt.Errorf(
			"load billing license runtime entitlements: %w", err,
		)
	}
	return assembleRuntimeEntitlements(environmentID, grants)
}

func assembleRuntimeEntitlements(
	environmentID uuid.UUID,
	grants []RuntimeGrant,
) (billingapi.LicenseRuntimeEntitlements, error) {
	if len(grants) != 14 || !validRuntimeGrantHead(environmentID, grants[0]) {
		return billingapi.LicenseRuntimeEntitlements{}, ErrAuthorityState
	}
	first := grants[0]
	result := billingapi.LicenseRuntimeEntitlements{
		LicenseID: first.LicenseID.String(), EnvironmentID: environmentID.String(),
		SnapshotSequence: first.SnapshotSequence, SnapshotSHA256: first.SnapshotSHA256,
		State: first.State, DatabaseNow: first.DatabaseNow.UTC(),
		Projection: billingapi.LicenseEntitlementProjection{
			Boolean: map[string]bool{}, Limits: map[string]billingapi.LicenseLimit{},
			Channel: first.Channel,
		},
	}
	for _, grant := range grants {
		if !sameRuntimeGrantHead(first, grant) || !appendRuntimeGrant(&result, grant) {
			return billingapi.LicenseRuntimeEntitlements{}, ErrAuthorityState
		}
	}
	if len(result.Projection.Boolean) != 7 || len(result.Projection.Limits) != 7 {
		return billingapi.LicenseRuntimeEntitlements{}, ErrAuthorityState
	}
	return result, nil
}

func validRuntimeGrantHead(environmentID uuid.UUID, grant RuntimeGrant) bool {
	digest, err := hex.DecodeString(grant.SnapshotSHA256)
	return grant.LicenseID != uuid.Nil && grant.EnvironmentID == environmentID &&
		grant.SnapshotSequence > 0 && err == nil && len(digest) == 32 &&
		(grant.State == RuntimeValid || grant.State == RuntimeGrace ||
			grant.State == RuntimeBuildOff) && grant.Channel != "" && !grant.DatabaseNow.IsZero()
}

func sameRuntimeGrantHead(expected, grant RuntimeGrant) bool {
	return grant.LicenseID == expected.LicenseID &&
		grant.EnvironmentID == expected.EnvironmentID &&
		grant.SnapshotSequence == expected.SnapshotSequence &&
		grant.SnapshotSHA256 == expected.SnapshotSHA256 &&
		grant.State == expected.State && grant.Channel == expected.Channel &&
		grant.DatabaseNow.Equal(expected.DatabaseNow)
}

func appendRuntimeGrant(
	result *billingapi.LicenseRuntimeEntitlements,
	grant RuntimeGrant,
) bool {
	if grant.ValueType == "boolean" && runtimeBooleanKey(grant.EntitlementKey) &&
		grant.BooleanValue != nil && grant.IntegerValue == nil && !grant.Unlimited {
		if _, exists := result.Projection.Boolean[grant.EntitlementKey]; exists {
			return false
		}
		result.Projection.Boolean[grant.EntitlementKey] = *grant.BooleanValue
		return true
	}
	if grant.ValueType != "integer" || !runtimeLimitKey(grant.EntitlementKey) ||
		grant.BooleanValue != nil || (grant.Unlimited == (grant.IntegerValue != nil)) ||
		(grant.IntegerValue != nil && *grant.IntegerValue < 0) {
		return false
	}
	if _, exists := result.Projection.Limits[grant.EntitlementKey]; exists {
		return false
	}
	limit := billingapi.LicenseLimit{Unlimited: grant.Unlimited}
	if grant.IntegerValue != nil {
		limit.Value = *grant.IntegerValue
	}
	result.Projection.Limits[grant.EntitlementKey] = limit
	return true
}

func runtimeBooleanKey(key string) bool {
	switch key {
	case "ai.chat", "analytics.dashboard_builder.enabled", "api.access",
		"api.builder.enabled", "forms.builder.enabled",
		"workflow.builder.enabled", "workflow.enabled":
		return true
	default:
		return false
	}
}

func runtimeLimitKey(key string) bool {
	switch key {
	case "analytics.active_dashboards", "api.active_endpoints",
		"documents.storage_bytes", "forms.active_forms", "forms.submissions",
		"iam.active_users", "workflow.active_definitions":
		return true
	default:
		return false
	}
}
