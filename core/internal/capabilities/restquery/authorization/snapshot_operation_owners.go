package authorization

import "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"

const maximumSnapshotOperationOwners = 4

func exactSnapshotGenerations(
	spec resolved.QuerySpec,
	verifiedGenerations map[string]int64,
) (map[string]int64, error) {
	owners, err := snapshotOwners(spec)
	if err != nil || len(verifiedGenerations) != len(owners) {
		return nil, errSnapshotOperationAdmission
	}
	owned := make(map[string]int64, len(owners))
	for owner := range owners {
		generation, found := verifiedGenerations[owner]
		if !found || generation <= 0 {
			return nil, errSnapshotOperationAdmission
		}
		owned[owner] = generation
	}
	return owned, nil
}

func snapshotOwners(spec resolved.QuerySpec) (map[string]struct{}, error) {
	requirements, err := extractRequirements(spec)
	if err != nil || !requirements.validate() {
		return nil, errSnapshotOperationAdmission
	}
	owners := make(map[string]struct{}, len(requirements.resources))
	for _, resource := range requirements.resources {
		reference, referenceErr := resource.catalog.Reference()
		if referenceErr != nil {
			return nil, errSnapshotOperationAdmission
		}
		owner, ownerErr := reference.Owner()
		if ownerErr != nil {
			return nil, errSnapshotOperationAdmission
		}
		owners[owner] = struct{}{}
	}
	if len(owners) < 1 || len(owners) > maximumSnapshotOperationOwners {
		return nil, errSnapshotOperationAdmission
	}
	return owners, nil
}
