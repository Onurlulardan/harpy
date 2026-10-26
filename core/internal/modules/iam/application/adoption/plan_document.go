package adoption

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

func EncodePlan(value Plan) ([]byte, string, error) {
	if value.ContractVersion != PlanContractVersion || value.TenantID == "" ||
		value.SourceSHA256 == "" || value.MappingSHA256 == "" {
		return nil, "", fmt.Errorf("IAM V1 adoption plan is incomplete")
	}
	document, err := json.Marshal(value)
	if err != nil {
		return nil, "", fmt.Errorf("marshal IAM V1 adoption plan: %w", err)
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(
		document, maximumCanonicalAdoptionBytes,
	)
	if err != nil {
		return nil, "", fmt.Errorf("canonicalize IAM V1 adoption plan: %w", err)
	}
	digest := sha256.Sum256(canonical)
	return canonical, hex.EncodeToString(digest[:]), nil
}
