package iam

import (
	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	"github.com/google/uuid"
)

func identityOrganizationOwnerCommand(
	operationID uuid.UUID,
	ownerIDs []uuid.UUID,
) ([]identityfacade.IdentityID, []identityfacade.IdentityID) {
	owners := identitySortedIDs(ownerIDs)
	commandOwners := make([]identityfacade.IdentityID, len(owners))
	memberships := make([]identityfacade.IdentityID, len(owners))
	for index, ownerID := range owners {
		commandOwners[index] = identityfacade.IdentityID(ownerID)
		memberships[index] = identityfacade.IdentityID(identityChildID(
			operationID, "organization-owner-membership:"+ownerID.String(),
		))
	}
	return commandOwners, memberships
}
