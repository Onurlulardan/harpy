package queryhttp

import (
	httptransport "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/httptransport"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
	"github.com/google/uuid"
)

func dependencies(values []dependencyDocument) ([]queryapp.DependencyRequest, bool) {
	if len(values) < 1 || len(values) > 32 {
		return nil, false
	}
	result := make([]queryapp.DependencyRequest, len(values))
	seen := make(map[[32]byte]struct{}, len(values))
	for index, value := range values {
		resourceID, resourceOK := httptransport.WireV7(value.ResourceID)
		versionID, versionOK := httptransport.WireV7(value.VersionID)
		if !resourceOK || !versionOK || value.OwnerDomain == "" || value.ResourceKind == "" {
			return nil, false
		}
		key := dependencyKey(resourceID, versionID)
		if _, exists := seen[key]; exists {
			return nil, false
		}
		seen[key] = struct{}{}
		result[index] = queryapp.DependencyRequest{OwnerDomain: value.OwnerDomain,
			ResourceKind: value.ResourceKind, ResourceID: resourceID, VersionID: versionID}
	}
	return result, true
}

func dependencyKey(resourceID, versionID uuid.UUID) [32]byte {
	result := [32]byte{}
	copy(result[:16], resourceID[:])
	copy(result[16:], versionID[:])
	return result
}
