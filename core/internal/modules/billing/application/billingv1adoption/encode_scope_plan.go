package billingv1adoptionapp

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

func EncodeScopePlan(plan ScopePlan) ([]byte, string, error) {
	if plan.ContractVersion != ScopePlanContractVersion ||
		(plan.ScopeKind != "global" && plan.ScopeKind != "tenant") ||
		plan.ScopeID == "" || plan.SourceTimezone != "UTC" ||
		!canonicalInstant(plan.SnapshotAt) || !validDigest(plan.SourceSHA256) ||
		plan.Eligible != (len(plan.Issues) == 0) {
		return nil, "", fmt.Errorf("Billing V1 adoption plan is incomplete")
	}
	document, err := json.Marshal(plan)
	if err != nil {
		return nil, "", fmt.Errorf("marshal Billing V1 adoption plan: %w", err)
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, maximumPlanBytes)
	if err != nil {
		return nil, "", fmt.Errorf("canonicalize Billing V1 adoption plan: %w", err)
	}
	digest := sha256.Sum256(canonical)
	return canonical, hex.EncodeToString(digest[:]), nil
}
