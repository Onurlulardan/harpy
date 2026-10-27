package reports

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
	Resource          string `json:"resource"`
	Operation         string `json:"operation"`
}

func policyRevision(
	catalogs *Catalog,
	spec resourceSpec,
	authority transaction.GlobalSnapshotAuthority,
	operation string,
) ([sha256.Size]byte, error) {
	entry, found := catalogs.entries[spec.Name]
	if !found {
		return [sha256.Size]byte{}, ErrContract
	}
	hash, err := entry.reference.CatalogHash()
	hashHex, hexErr := hash.Hex()
	document := policyRevisionDocument{
		Version:           "billing-platform-report-rest-query-policy.v1",
		EnvironmentID:     authority.EnvironmentID().String(),
		GroupGeneration:   authority.GroupGeneration(),
		BillingGeneration: authority.BillingGeneration(),
		CatalogSHA256:     hashHex, Resource: spec.Name, Operation: operation,
	}
	canonical, marshalErr := json.Marshal(document)
	if err != nil || hexErr != nil || marshalErr != nil ||
		document.EnvironmentID == "00000000-0000-0000-0000-000000000000" ||
		document.GroupGeneration < 1 || document.BillingGeneration < 1 ||
		!allowedReportOperation(spec, operation) {
		return [sha256.Size]byte{}, ErrContract
	}
	digest := sha256.Sum256(canonical)
	if digest == [sha256.Size]byte{} {
		return [sha256.Size]byte{}, ErrContract
	}
	return digest, nil
}
