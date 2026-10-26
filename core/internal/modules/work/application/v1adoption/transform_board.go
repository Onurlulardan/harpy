package v1adoption

import (
	"sort"
	"strings"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
)

func transformBoards(plan *Plan, index legacyIndex) error {
	memberKeys := map[string]bool{}
	for _, membership := range index.memberships {
		memberKeys[membership.BoardID.String()+"/"+membership.UserID.String()] = true
	}
	for id, board := range index.boards {
		source := index.source(KindBoard, id)
		if !board.ID.Valid() || !board.CreatedBy.Valid() || !board.OrganizationID.Valid() ||
			strings.TrimSpace(board.Name) == "" || len(board.Name) > 200 ||
			len(board.Description) > 16_000 || board.CreatedAt.IsZero() || board.UpdatedAt.IsZero() {
			appendIssue(plan, source, IssueFutureWork)
			continue
		}
		lifecycle := "ACTIVE"
		if board.DeletedAt != nil {
			lifecycle = "ARCHIVED"
		}
		document := TargetBoardDocument{CreatorID: board.CreatedBy.String(), Name: board.Name,
			Description: board.Description, Lifecycle: lifecycle,
			CreatedAt: board.CreatedAt.UTC(), UpdatedAt: board.UpdatedAt.UTC(),
			ArchivedAt: utcPointer(board.DeletedAt)}
		if err := appendTarget(plan, source, TargetBoard, id, document); err != nil {
			return err
		}
		if !memberKeys[id.String()+"/"+board.CreatedBy.String()] {
			membershipID := derivedID("board-creator", id.String(), board.CreatedBy.String())
			membership := TargetMembershipDocument{BoardID: id.String(), PrincipalKind: "USER",
				PrincipalID: board.CreatedBy.String(), Role: "ADMIN",
				CreatedAt: board.CreatedAt.UTC(), UpdatedAt: board.UpdatedAt.UTC()}
			if err := appendTarget(plan, source, TargetMembership, membershipID, membership); err != nil {
				return err
			}
		}
		organizationID := derivedID("board-organization", id.String(), board.OrganizationID.String())
		organization := TargetMembershipDocument{BoardID: id.String(), PrincipalKind: "ORGANIZATION",
			PrincipalID: board.OrganizationID.String(), Role: "VIEWER",
			CreatedAt: board.CreatedAt.UTC(), UpdatedAt: board.UpdatedAt.UTC()}
		if err := appendTarget(plan, source, TargetMembership, organizationID, organization); err != nil {
			return err
		}
	}
	return transformColumns(plan, index)
}

func transformColumns(plan *Plan, index legacyIndex) error {
	grouped := map[domain.ID][]LegacyColumn{}
	for _, column := range index.columns {
		grouped[column.BoardID] = append(grouped[column.BoardID], column)
	}
	for boardID, columns := range grouped {
		sort.Slice(columns, func(left, right int) bool {
			if columns[left].Order != columns[right].Order {
				return columns[left].Order < columns[right].Order
			}
			return columns[left].ID.String() < columns[right].ID.String()
		})
		for position, column := range columns {
			source := index.source(KindColumn, column.ID)
			status, valid := mapColumnStatus(column.LinkedStatus)
			color := column.Color
			if color == "" {
				color = domain.DefaultColor
			}
			if _, exists := index.boards[boardID]; !exists {
				appendIssue(plan, source, IssueOrphan)
				continue
			}
			if !valid || !domain.ValidColor(color) || strings.TrimSpace(column.Name) == "" ||
				len(columns) > domain.MaxBoardColumns {
				appendIssue(plan, source, IssueInvalidEnum)
				continue
			}
			document := TargetColumnDocument{BoardID: boardID.String(), Name: column.Name,
				Color:  color,
				Status: status, Position: int64(position+1) * domain.SparseGap,
				CreatedAt: column.CreatedAt.UTC(), UpdatedAt: column.UpdatedAt.UTC()}
			if err := appendTarget(plan, source, TargetColumn, column.ID, document); err != nil {
				return err
			}
		}
	}
	return nil
}

func utcPointer(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	result := value.UTC()
	return &result
}
