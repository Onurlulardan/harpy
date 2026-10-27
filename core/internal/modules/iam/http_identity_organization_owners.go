package iam

import (
	"encoding/json"

	"github.com/google/uuid"
)

func decodeIdentityOrganizationOwners(
	ownerID json.RawMessage,
	ownerIDs json.RawMessage,
	required bool,
) ([]uuid.UUID, bool, bool) {
	ownerFound, ownersFound := len(ownerID) != 0, len(ownerIDs) != 0
	if ownerFound && ownersFound {
		return nil, false, false
	}
	if ownerFound {
		var value string
		if json.Unmarshal(ownerID, &value) != nil {
			return nil, false, false
		}
		parsed, valid := decodeUUID(value)
		if !valid {
			return nil, false, false
		}
		return []uuid.UUID{parsed}, true, true
	}
	if ownersFound {
		var values []string
		if json.Unmarshal(ownerIDs, &values) != nil {
			return nil, false, false
		}
		parsed, valid := decodeUUIDList(values)
		if !valid || len(parsed) == 0 {
			return nil, false, false
		}
		return identitySortedIDs(parsed), true, true
	}
	return nil, false, !required
}
