package authorization

import (
	"crypto/sha256"
	"slices"

	"github.com/google/uuid"
)

func groupPolicyRequests(
	tenantID uuid.UUID,
	actorID uuid.UUID,
	queryIdentity [sha256.Size]byte,
	requirements []Requirement,
) ([]PolicyRequest, error) {
	if !validAuthorizationUUID(tenantID) || !validAuthorizationUUID(actorID) ||
		queryIdentity == [sha256.Size]byte{} || len(requirements) < 1 ||
		len(requirements) > maximumPolicyRequirements {
		return nil, errPolicyEvaluationContract
	}
	grouped := make(map[string][]Requirement)
	for _, requirement := range requirements {
		if requirement.Validate() != nil {
			return nil, errPolicyEvaluationContract
		}
		owner, err := requirement.catalog.Owner()
		if err != nil {
			return nil, errPolicyEvaluationContract
		}
		grouped[owner] = append(grouped[owner], requirement)
	}
	if len(grouped) < 1 || len(grouped) > maximumPolicyOwners {
		return nil, errPolicyEvaluationContract
	}
	requests := make([]PolicyRequest, 0, len(grouped))
	for owner, owned := range grouped {
		request, err := newPolicyRequest(
			tenantID, actorID, queryIdentity, owner, owned,
		)
		if err != nil {
			return nil, errPolicyEvaluationContract
		}
		requests = append(requests, request)
	}
	slices.SortFunc(requests, func(left, right PolicyRequest) int {
		return compareOwner(left.owner, right.owner)
	})
	return requests, nil
}
