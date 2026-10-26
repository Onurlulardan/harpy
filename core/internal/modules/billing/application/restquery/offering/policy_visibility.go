package offering

import (
	"crypto/sha256"

	restauthorization "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/authorization"
)

func policyVisibilities(
	resourceID [sha256.Size]byte,
	public bool,
	allowed bool,
	request restauthorization.GlobalPolicyRequest,
) ([]restauthorization.Visibility, error) {
	if !allowed {
		return []restauthorization.Visibility{}, nil
	}
	if !public {
		visibility, err := restauthorization.NewAllVisibility(resourceID)
		if err != nil {
			return nil, ErrContract
		}
		return []restauthorization.Visibility{visibility}, nil
	}
	snapshotAt, err := request.SnapshotAt()
	if err != nil {
		return nil, ErrContract
	}
	filter, err := publicFilter(snapshotAt)
	if err != nil {
		return nil, ErrContract
	}
	visibility, err := restauthorization.NewPredicateVisibility(resourceID, filter)
	if err != nil {
		return nil, ErrContract
	}
	return []restauthorization.Visibility{visibility}, nil
}
