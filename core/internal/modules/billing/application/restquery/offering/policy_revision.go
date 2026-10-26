package offering

import (
	"crypto/sha256"
	"encoding/json"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type policyRevisionDocument struct {
	Version           string `json:"version"`
	EnvironmentID     string `json:"environment_id"`
	GroupGeneration   int64  `json:"group_generation"`
	BillingGeneration int64  `json:"billing_generation"`
	CatalogSHA256     string `json:"catalog_sha256"`
	Audience          string `json:"audience"`
	Operation         string `json:"operation"`
}

func policyRevision(
	catalogs *Catalog,
	authority transaction.GlobalSnapshotAuthority,
	public bool,
	operation string,
) ([sha256.Size]byte, error) {
	reference, err := catalogs.Reference()
	if err != nil {
		return [sha256.Size]byte{}, ErrContract
	}
	hash, err := reference.CatalogHash()
	hashHex, hexErr := hash.Hex()
	audience := "authenticated"
	if public {
		audience = "public"
	}
	document := policyRevisionDocument{
		Version:           "billing-offering-rest-query-policy.v1",
		EnvironmentID:     authority.EnvironmentID().String(),
		GroupGeneration:   authority.GroupGeneration(),
		BillingGeneration: authority.BillingGeneration(),
		CatalogSHA256:     hashHex, Audience: audience, Operation: operation,
	}
	canonical, marshalErr := json.Marshal(document)
	if err != nil || hexErr != nil || marshalErr != nil ||
		document.EnvironmentID == "00000000-0000-0000-0000-000000000000" ||
		document.GroupGeneration < 1 || document.BillingGeneration < 1 {
		return [sha256.Size]byte{}, ErrContract
	}
	digest := sha256.Sum256(canonical)
	if digest == [sha256.Size]byte{} {
		return [sha256.Size]byte{}, ErrContract
	}
	return digest, nil
}
