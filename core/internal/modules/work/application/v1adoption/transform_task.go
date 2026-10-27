package v1adoption

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (service *Service) transformTasks(ctx context.Context, scope ports.Scope,
	plan *Plan, index legacyIndex,
) error {
	positions := taskPositions(index)
	for id, task := range index.tasks {
		source := index.source(KindTask, id)
		board, boardExists := index.boards[task.BoardID]
		column, columnExists := index.columns[task.ColumnID]
		if !boardExists || !columnExists {
			appendIssue(plan, source, IssueOrphan)
			continue
		}
		if column.BoardID != task.BoardID {
			appendIssue(plan, source, IssueCrossBoard)
			continue
		}
		if task.WorkflowTaskID != nil || task.TaskType == "APPROVAL" {
			if err := service.transformWorkflowTask(ctx, scope, plan, source, task); err != nil {
				return err
			}
			continue
		}
		kind, kindValid := mapTaskKind(task.TaskType)
		status, statusValid := mapStatus(task.Status)
		priority := domain.TaskPriority(task.Priority)
		if !kindValid || !statusValid || !priority.Valid() || !task.CreatedBy.Valid() ||
			strings.TrimSpace(task.Title) == "" ||
			len(task.Title) > 300 || len(task.Description) > 16_000 ||
			positions[id] < 1 || task.CreatedAt.IsZero() || task.UpdatedAt.IsZero() {
			appendIssue(plan, source, IssueInvalidEnum)
			continue
		}
		completed, cancelled := terminalTimes(task)
		if status == "DONE" && completed == nil {
			appendIssue(plan, source, IssueInvalidEnum)
			continue
		}
		legacy, legacyErr := domain.NormalizeLegacyTaskMetadata(domain.LegacyTaskMetadata{
			IsRecurring: task.IsRecurring, RecurrenceRule: task.RecurrenceRule,
			RecurrenceEnd: task.RecurrenceEnd, FormFilter: task.FormFilter})
		if legacyErr != nil {
			appendIssue(plan, source, IssueFutureWork)
			legacy, _ = domain.NormalizeLegacyTaskMetadata(domain.LegacyTaskMetadata{})
		}
		document := TargetTaskDocument{BoardID: task.BoardID.String(),
			ColumnID: task.ColumnID.String(), CreatorID: task.CreatedBy.String(),
			Kind: kind, Status: status, Priority: string(priority),
			Title: task.Title, Description: task.Description, Position: positions[id],
			StartsAt: utcPointer(task.StartDate), DueAt: utcPointer(task.DueDate),
			CompletedAt: completed, CancelledAt: cancelled,
			IsRecurring: legacy.IsRecurring, RecurrenceRule: legacy.RecurrenceRule,
			RecurrenceEnd: utcPointer(legacy.RecurrenceEnd), FormFilter: legacy.FormFilter,
			DeletedAt: utcPointer(task.DeletedAt), CreatedAt: utcPointer(&task.CreatedAt),
			UpdatedAt: utcPointer(&task.UpdatedAt)}
		if err := appendTarget(plan, source, TargetTask, id, document); err != nil {
			return err
		}
		if err := service.transformTaskRelations(ctx, scope, plan, source, board, task, index); err != nil {
			return err
		}
	}
	return nil
}

func taskPositions(index legacyIndex) map[domain.ID]int64 {
	grouped := map[domain.ID][]LegacyTask{}
	for _, task := range index.tasks {
		grouped[task.ColumnID] = append(grouped[task.ColumnID], task)
	}
	positions := map[domain.ID]int64{}
	for _, tasks := range grouped {
		sort.Slice(tasks, func(left, right int) bool {
			if tasks[left].ColumnOrder != tasks[right].ColumnOrder {
				return tasks[left].ColumnOrder < tasks[right].ColumnOrder
			}
			return tasks[left].ID.String() < tasks[right].ID.String()
		})
		if len(tasks) > domain.MaxColumnTasks {
			continue
		}
		for index, task := range tasks {
			positions[task.ID] = int64(index+1) * domain.SparseGap
		}
	}
	return positions
}

func terminalTimes(task LegacyTask) (*time.Time, *time.Time) {
	if task.Status == "DONE" {
		return utcPointer(task.CompletedAt), nil
	}
	if task.Status == "CANCELLED" {
		return nil, utcPointer(&task.UpdatedAt)
	}
	return nil, nil
}

func (service *Service) transformWorkflowTask(ctx context.Context, scope ports.Scope,
	plan *Plan, source SourceRecord, task LegacyTask,
) error {
	if task.WorkflowTaskID == nil || !task.WorkflowTaskID.Valid() {
		appendIssue(plan, source, IssueFutureWork)
		return nil
	}
	projection, err := service.dependencies.Workflow.ResolveTask(ctx, scope, *task.WorkflowTaskID)
	if err != nil || !projection.TaskID.Valid() || !projection.VersionID.Valid() ||
		projection.Revision < 1 || projection.PermissionRevision < 1 ||
		strings.TrimSpace(projection.Summary) == "" ||
		len(projection.VisibleTo) == 0 || len(projection.VisibleTo) > 100 {
		appendIssue(plan, source, IssueFutureWork)
		return nil
	}
	visibleTo := make([]string, len(projection.VisibleTo))
	seen := make(map[domain.Principal]struct{}, len(projection.VisibleTo))
	for index, principal := range projection.VisibleTo {
		if !principal.Valid() {
			appendIssue(plan, source, IssueFutureWork)
			return nil
		}
		if _, duplicate := seen[principal]; duplicate {
			appendIssue(plan, source, IssueFutureWork)
			return nil
		}
		seen[principal] = struct{}{}
		visibleTo[index] = string(principal.Kind) + ":" + principal.ID.String()
	}
	sort.Strings(visibleTo)
	document := TargetProjectionDocument{SourceID: projection.TaskID.String(),
		VersionID: projection.VersionID.String(), Revision: projection.Revision,
		PermissionRevision: projection.PermissionRevision, Summary: projection.Summary,
		DueAt: utcPointer(projection.DueAt), VisibleToPrincipals: visibleTo}
	id := derivedID("workflow-projection", projection.TaskID.String())
	return appendTarget(plan, source, TargetProjection, id, document)
}
