package query

import (
	"encoding/hex"
	"encoding/json"
	"math"

	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
	analyticsdependency "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/dependency"
)

type dependencyDocument struct {
	OwnerDomain          string `json:"owner_domain"`
	ResourceKind         string `json:"resource_kind"`
	ResourceID           string `json:"resource_id"`
	ResourceVersionID    string `json:"resource_version_id"`
	OwnerGeneration      uint64 `json:"owner_generation"`
	CatalogVersion       uint64 `json:"catalog_version"`
	QueryContractVersion uint64 `json:"query_contract_version"`
	PolicyVersion        uint64 `json:"policy_version"`
	DataRevision         string `json:"data_revision_capability"`
	DependencySHA256     string `json:"dependency_sha256"`
}

func encodeDependencies(values []queryapp.DependencySnapshot) ([]byte, error) {
	if len(values) == 0 || len(values) > 32 {
		return nil, queryapp.ErrInvalid
	}
	documents := make([]dependencyDocument, len(values))
	seen := make(map[queryapp.DependencyRequest]struct{}, len(values))
	for index, value := range values {
		if !validDependencySnapshot(value) {
			return nil, queryapp.ErrInvalid
		}
		if _, duplicate := seen[value.DependencyRequest]; duplicate {
			return nil, queryapp.ErrInvalid
		}
		seen[value.DependencyRequest] = struct{}{}
		documents[index] = dependencyDocument{
			OwnerDomain: value.OwnerDomain, ResourceKind: value.ResourceKind,
			ResourceID: value.ResourceID.String(), ResourceVersionID: value.VersionID.String(),
			OwnerGeneration: value.OwnerGeneration, CatalogVersion: value.CatalogVersion,
			QueryContractVersion: value.QueryContractVersion, PolicyVersion: value.PolicyVersion,
			DataRevision:     value.DataRevisionCapability,
			DependencySHA256: hex.EncodeToString(value.DependencySHA256[:]),
		}
	}
	encoded, err := json.Marshal(documents)
	if err != nil {
		return nil, queryapp.ErrUnavailable
	}
	return encoded, nil
}

func validDependencySnapshot(value queryapp.DependencySnapshot) bool {
	if !value.WriterV2 || value.Retired || !validUUIDv7(value.ResourceID) ||
		!validUUIDv7(value.VersionID) || value.OwnerGeneration == 0 ||
		value.OwnerGeneration > math.MaxInt64 || value.CatalogVersion == 0 ||
		value.CatalogVersion > math.MaxInt64 || value.QueryContractVersion == 0 ||
		value.QueryContractVersion > math.MaxInt64 || value.PolicyVersion == 0 ||
		value.PolicyVersion > math.MaxInt64 || value.DependencySHA256 == [32]byte{} {
		return false
	}
	_, err := analyticsdependency.NewEntry(value.OwnerDomain, value.ResourceKind,
		[16]byte(value.ResourceID), [16]byte(value.VersionID), value.CatalogVersion,
		value.QueryContractVersion, value.PolicyVersion,
		value.DataRevisionCapability, value.DependencySHA256)
	return err == nil
}
