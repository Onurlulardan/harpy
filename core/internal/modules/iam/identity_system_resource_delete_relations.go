package iam

import (
	"fmt"
	"math"

	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	"github.com/google/uuid"
)

const identityUserDeleteRelationBatchSize = 100

func terminalizeIdentityUserDeleteRelations(
	command identitySystemResourceCommand,
	deletion identitySystemResourceDelete,
	operationID uuid.UUID,
	requestSHA [32]byte,
	relations identityUserDeleteCommands,
) (int64, error) {
	revision, err := revokeIdentityUserDeleteAssignments(
		command, deletion, operationID, requestSHA, relations,
	)
	if err != nil {
		return 0, err
	}
	return tombstoneIdentityUserDeleteMemberships(
		command, deletion, operationID, requestSHA, revision, relations,
	)
}

func revokeIdentityUserDeleteAssignments(
	command identitySystemResourceCommand,
	deletion identitySystemResourceDelete,
	operationID uuid.UUID,
	requestSHA [32]byte,
	relations identityUserDeleteCommands,
) (int64, error) {
	revision := deletion.expectedRevision
	var after identityfacade.IdentityID
	for {
		rows, err := relations.ListAssignments(
			command.ctx, command.scope, identityfacade.ListAssignmentsQuery{
				AfterID: after, Limit: identityUserDeleteRelationBatchSize,
				Evidence: identityQueryEvidence(command.evidence, command.permissionID),
			},
		)
		if err != nil {
			return 0, fmt.Errorf("list user assignments: %w", err)
		}
		next, done, err := identityUserDeleteAssignmentCursor(after, rows)
		if err != nil {
			return 0, fmt.Errorf("advance assignment cursor: %w", err)
		}
		for _, assignment := range rows {
			if uuid.UUID(assignment.UserID) != deletion.id ||
				assignment.State == identityfacade.AssignmentRevoked {
				continue
			}
			result, revokeErr := relations.RevokeAssignment(
				command.ctx, command.scope, identityfacade.RevokeAssignmentCommand{
					ID: assignment.ID, ExpectedRevision: assignment.Revision,
					Evidence: identityUserDeleteChildEvidence(
						command, operationID, requestSHA, "assignment-revoke", assignment.ID,
					),
				},
			)
			if revokeErr != nil {
				return 0, fmt.Errorf("revoke user assignment: %w", revokeErr)
			}
			revision, err = advanceIdentityUserDeleteRevision(revision, result.Replayed)
			if err != nil {
				return 0, err
			}
		}
		if done {
			return revision, nil
		}
		after = next
	}
}

func tombstoneIdentityUserDeleteMemberships(
	command identitySystemResourceCommand,
	deletion identitySystemResourceDelete,
	operationID uuid.UUID,
	requestSHA [32]byte,
	revision int64,
	relations identityUserDeleteCommands,
) (int64, error) {
	var after identityfacade.IdentityID
	for {
		rows, err := relations.ListMemberships(
			command.ctx, command.scope, identityfacade.ListMembershipsQuery{
				AfterID: after, Limit: identityUserDeleteRelationBatchSize,
				Evidence: identityQueryEvidence(command.evidence, command.permissionID),
			},
		)
		if err != nil {
			return 0, fmt.Errorf("list user memberships: %w", err)
		}
		next, done, err := identityUserDeleteMembershipCursor(after, rows)
		if err != nil {
			return 0, fmt.Errorf("advance membership cursor: %w", err)
		}
		for _, membership := range rows {
			if uuid.UUID(membership.UserID) != deletion.id ||
				membership.State == identityfacade.MembershipTombstoned {
				continue
			}
			result, deleteErr := relations.DeleteMembership(
				command.ctx, command.scope, identityfacade.DeleteMembershipCommand{
					ID: membership.ID, ExpectedRevision: membership.Revision,
					Evidence: identityUserDeleteChildEvidence(
						command, operationID, requestSHA, "membership-delete", membership.ID,
					),
				},
			)
			if deleteErr != nil {
				return 0, fmt.Errorf("delete user membership: %w", deleteErr)
			}
			revision, err = advanceIdentityUserDeleteRevision(revision, result.Replayed)
			if err != nil {
				return 0, err
			}
		}
		if done {
			return revision, nil
		}
		after = next
	}
}

func identityUserDeleteChildEvidence(
	command identitySystemResourceCommand,
	operationID uuid.UUID,
	requestSHA [32]byte,
	purpose string,
	relationID identityfacade.IdentityID,
) identityfacade.MutationEvidence {
	childID := identityChildID(
		operationID, purpose+":"+uuid.UUID(relationID).String(),
	)
	return identityMutationEvidence(
		command.evidence, command.permissionID, childID, requestSHA,
	)
}

func advanceIdentityUserDeleteRevision(revision int64, replayed bool) (int64, error) {
	if replayed {
		return revision, nil
	}
	if revision == math.MaxInt64 {
		return 0, ErrIdentityHTTPUnavailable
	}
	return revision + 1, nil
}
