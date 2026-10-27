package billingv1

import (
	"encoding/json"
	"math"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func decodePlanDocuments(
	limitsDocument, featuresDocument *json.RawMessage,
) (planLimits, planFeatures, error) {
	var limits planLimits
	var features planFeatures
	if limitsDocument != nil &&
		json.Unmarshal(*limitsDocument, &limits) != nil {
		return planLimits{}, planFeatures{}, ErrRequestInvalid
	}
	if featuresDocument != nil &&
		json.Unmarshal(*featuresDocument, &features) != nil {
		return planLimits{}, planFeatures{}, ErrRequestInvalid
	}
	if !validLegacyLimit(limits.SubmissionsPerMonth) ||
		!validLegacyLimit(limits.MaxUsers) || !validLegacyLimit(limits.MaxForms) ||
		!validLegacyLimit(limits.MaxStorageMB) ||
		limits.MaxStorageMB > math.MaxInt64/(1024*1024) {
		return planLimits{}, planFeatures{}, ErrRequestInvalid
	}
	return limits, features, nil
}

func validLegacyLimit(value int64) bool { return value == -1 || value >= 0 }

func legacyEntitlements(
	limits planLimits,
	features planFeatures,
) []billingapi.EntitlementInput {
	return []billingapi.EntitlementInput{
		integerEntitlement("forms.submissions", limits.SubmissionsPerMonth),
		integerEntitlement("iam.active_users", limits.MaxUsers),
		integerEntitlement("forms.active_forms", limits.MaxForms),
		integerEntitlement("documents.storage_bytes", storageBytes(limits.MaxStorageMB)),
		booleanEntitlement("ai.chat", features.AIChat),
		booleanEntitlement("workflow.enabled", features.Workflows),
		booleanEntitlement("api.access", features.APIAccess),
	}
}

func integerEntitlement(key string, value int64) billingapi.EntitlementInput {
	if value == -1 {
		return billingapi.EntitlementInput{Key: key, Kind: "integer", Unlimited: true}
	}
	copy := value
	return billingapi.EntitlementInput{Key: key, Kind: "integer", Integer: &copy}
}

func booleanEntitlement(key string, value bool) billingapi.EntitlementInput {
	copy := value
	return billingapi.EntitlementInput{Key: key, Kind: "boolean", Boolean: &copy}
}

func storageBytes(megabytes int64) int64 {
	if megabytes == -1 {
		return -1
	}
	return megabytes * 1024 * 1024
}
