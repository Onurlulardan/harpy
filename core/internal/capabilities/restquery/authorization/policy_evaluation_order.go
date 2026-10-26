package authorization

import (
	"crypto/sha256"
	"slices"

	"github.com/google/uuid"
)

func orderedPolicyRequests(
	registry PolicyRegistry,
	tenantID uuid.UUID,
	actorID uuid.UUID,
	requests []PolicyRequest,
) ([]PolicyRequest, error) {
	if registry.Validate() != nil || !validAuthorizationUUID(tenantID) ||
		!validAuthorizationUUID(actorID) || len(requests) < 1 ||
		len(requests) > maximumPolicyOwners {
		return nil, errPolicyEvaluationContract
	}
	owned := append([]PolicyRequest{}, requests...)
	slices.SortFunc(owned, func(left, right PolicyRequest) int {
		return compareOwner(left.owner, right.owner)
	})
	queryIdentity := [sha256.Size]byte{}
	for index, request := range owned {
		if request.Validate() != nil || request.tenantID != tenantID ||
			request.actorID != actorID {
			return nil, errPolicyEvaluationContract
		}
		if index == 0 {
			queryIdentity = request.queryIdentity
		} else if request.owner == owned[index-1].owner ||
			request.queryIdentity != queryIdentity {
			return nil, errPolicyEvaluationContract
		}
		if _, found := registry.lookup(request.owner); !found {
			return nil, errPolicyEvaluationContract
		}
	}
	return owned, nil
}

func compareOwner(left, right string) int {
	switch {
	case left < right:
		return -1
	case left > right:
		return 1
	default:
		return 0
	}
}
