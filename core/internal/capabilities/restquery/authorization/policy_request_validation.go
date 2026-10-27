package authorization

import (
	"bytes"

	"github.com/google/uuid"
)

func validAuthorizationUUID(value uuid.UUID) bool {
	return value != uuid.Nil && value.Variant() == uuid.RFC4122
}

func validPolicyRequirements(owner string, requirements []Requirement) bool {
	resourceRead := false
	for index, requirement := range requirements {
		if requirement.Validate() != nil {
			return false
		}
		catalogOwner, err := requirement.catalog.Owner()
		if err != nil || catalogOwner != owner {
			return false
		}
		if index > 0 && bytes.Compare(
			requirements[index-1].id[:], requirement.id[:],
		) >= 0 {
			return false
		}
		resourceRead = resourceRead || requirement.kind == RequirementResourceRead
	}
	return resourceRead
}
