package v1adoption

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (service *Service) transformAssignee(ctx context.Context, scope ports.Scope,
	plan *Plan, source SourceRecord, board LegacyBoard, task LegacyTask, index legacyIndex,
) error {
	if task.AssigneeID == nil && task.AssigneeType == "" {
		return nil
	}
	if task.AssigneeID == nil || !task.AssigneeID.Valid() ||
		(task.AssigneeType != "USER" && task.AssigneeType != "ROLE") {
		appendIssue(plan, source, IssueInvalidEnum)
		return nil
	}
	principal := domain.Principal{Kind: domain.PrincipalKind(task.AssigneeType), ID: *task.AssigneeID}
	exists, revision, err := service.dependencies.Principals.Exists(ctx, scope, principal)
	if err != nil || !exists || revision < 1 {
		appendIssue(plan, source, IssueOrphan)
		return nil
	}
	if !hasBoardAccess(index, task.BoardID, principal) {
		membershipID := derivedID("assignee-membership", task.BoardID.String(),
			string(principal.Kind), principal.ID.String())
		document := TargetMembershipDocument{BoardID: task.BoardID.String(),
			PrincipalKind: string(principal.Kind), PrincipalID: principal.ID.String(), Role: "VIEWER",
			CreatedAt: board.CreatedAt.UTC(), UpdatedAt: board.UpdatedAt.UTC()}
		if err = appendTarget(plan, source, TargetMembership, membershipID, document); err != nil {
			return err
		}
	}
	assigneeID := derivedID("task-assignee", task.ID.String(), string(principal.Kind), principal.ID.String())
	document := TargetAssigneeDocument{TaskID: task.ID.String(),
		PrincipalKind: string(principal.Kind), PrincipalID: principal.ID.String()}
	return appendTarget(plan, source, TargetAssignee, assigneeID, document)
}

func hasBoardAccess(index legacyIndex, boardID domain.ID, principal domain.Principal) bool {
	if principal.Kind != domain.PrincipalUser {
		return false
	}
	if board, exists := index.boards[boardID]; exists && board.CreatedBy == principal.ID {
		return true
	}
	for _, membership := range index.memberships {
		if membership.BoardID == boardID && membership.UserID == principal.ID {
			return true
		}
	}
	return false
}
