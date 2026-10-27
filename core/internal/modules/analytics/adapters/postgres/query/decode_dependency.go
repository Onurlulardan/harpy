package query

import (
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
	analyticsdependency "github.com/Onurlulardan/harpy/core/internal/modules/analytics/domain/dependency"
	"github.com/google/uuid"
)

func scanDependency(scanner interface{ Scan(...any) error }) (
	queryapp.DependencySnapshot, error,
) {
	var owner, kind, revision string
	var resourceID, versionID uuid.UUID
	var ownerGeneration, catalogVersion, queryVersion, policyVersion int64
	var digest []byte
	if err := scanner.Scan(&owner, &kind, &resourceID, &versionID, &ownerGeneration,
		&catalogVersion, &queryVersion, &policyVersion, &revision, &digest); err != nil {
		return queryapp.DependencySnapshot{}, queryapp.ErrUnavailable
	}
	if ownerGeneration < 1 || catalogVersion < 1 || queryVersion < 1 ||
		policyVersion < 1 || len(digest) != 32 {
		return queryapp.DependencySnapshot{}, queryapp.ErrUnavailable
	}
	var dependencySHA [32]byte
	copy(dependencySHA[:], digest)
	if _, err := analyticsdependency.NewEntry(owner, kind, [16]byte(resourceID),
		[16]byte(versionID), uint64(catalogVersion), uint64(queryVersion),
		uint64(policyVersion), revision, dependencySHA); err != nil {
		return queryapp.DependencySnapshot{}, queryapp.ErrUnavailable
	}
	return queryapp.DependencySnapshot{DependencyRequest: queryapp.DependencyRequest{
		OwnerDomain: owner, ResourceKind: kind, ResourceID: resourceID,
		VersionID: versionID}, CatalogVersion: uint64(catalogVersion),
		QueryContractVersion: uint64(queryVersion), PolicyVersion: uint64(policyVersion),
		OwnerGeneration: uint64(ownerGeneration), WriterV2: true,
		DataRevisionCapability: revision, DependencySHA256: dependencySHA}, nil
}
