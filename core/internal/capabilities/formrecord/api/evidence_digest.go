package recordkernel

import (
	"crypto/sha256"
	"encoding/json"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

func evidenceDigest(value any) ([32]byte, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return [32]byte{}, ErrCatalogInvalid
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(encoded, MaximumCatalogBytes)
	if err != nil {
		return [32]byte{}, ErrCatalogInvalid
	}
	return sha256.Sum256(canonical), nil
}

type BudgetProfile struct {
	ID     string
	Limits BudgetLimits
}

type EvidenceField struct {
	FieldID    string `json:"field_id"`
	PublicName string `json:"public_name"`
	Type       string `json:"type"`
	Role       string `json:"role"`
	Generation uint32 `json:"generation"`
}
