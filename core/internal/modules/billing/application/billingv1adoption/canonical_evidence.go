package billingv1adoptionapp

import (
	"encoding/json"
	"fmt"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

const maximumPlanBytes = 64 * 1024 * 1024

func categoryEvidence(category string, rows any, count int) (CategoryEvidence, error) {
	document, err := json.Marshal(rows)
	if err != nil {
		return CategoryEvidence{}, fmt.Errorf("marshal Billing V1 %s evidence: %w", category, err)
	}
	digest, err := canonicaljsonv1.DigestHex(document, maximumPlanBytes)
	if err != nil {
		return CategoryEvidence{}, fmt.Errorf("digest Billing V1 %s evidence: %w", category, err)
	}
	return CategoryEvidence{Category: category, RowCount: count, RootSHA256: digest}, nil
}

func sourceRoot(evidence []CategoryEvidence) (string, error) {
	document, err := json.Marshal(evidence)
	if err != nil {
		return "", fmt.Errorf("marshal Billing V1 source evidence: %w", err)
	}
	digest, err := canonicaljsonv1.DigestHex(document, maximumPlanBytes)
	if err != nil {
		return "", fmt.Errorf("digest Billing V1 source evidence: %w", err)
	}
	return digest, nil
}

func recordDigest(value any) (string, error) {
	document, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("marshal Billing V1 canonical record: %w", err)
	}
	digest, err := canonicaljsonv1.DigestHex(document, maximumPlanBytes)
	if err != nil {
		return "", fmt.Errorf("digest Billing V1 canonical record: %w", err)
	}
	return digest, nil
}
