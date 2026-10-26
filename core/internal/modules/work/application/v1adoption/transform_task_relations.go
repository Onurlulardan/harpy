package v1adoption

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (service *Service) transformTaskRelations(ctx context.Context, scope ports.Scope,
	plan *Plan, source SourceRecord, board LegacyBoard, task LegacyTask, index legacyIndex,
) error {
	if err := service.transformFormLink(ctx, scope, plan, source, task); err != nil {
		return err
	}
	if err := service.transformAssignee(ctx, scope, plan, source, board, task, index); err != nil {
		return err
	}
	if err := transformLabels(plan, source, task); err != nil {
		return err
	}
	if err := transformChecklist(plan, source, task); err != nil {
		return err
	}
	return service.transformTaskRecurrence(ctx, scope, plan, source, board, task)
}

func (service *Service) transformFormLink(ctx context.Context, scope ports.Scope,
	plan *Plan, source SourceRecord, task LegacyTask,
) error {
	formKind := task.TaskType == "FORM_FILL" || task.TaskType == "FORM_REVIEW"
	if !formKind && task.FormSlug == "" && task.FormDataID == nil {
		return nil
	}
	if !formKind || strings.TrimSpace(task.FormSlug) == "" {
		appendIssue(plan, source, IssueUnresolvedForm)
		return nil
	}
	reference, err := service.dependencies.Forms.ResolveForm(
		ctx, scope, task.FormSlug, task.FormDataID, task.FormFilter)
	if err != nil || !reference.ID.Valid() || !reference.VersionID.Valid() ||
		reference.PermissionRevision < 1 {
		appendIssue(plan, source, IssueUnresolvedForm)
		return nil
	}
	document := TargetFormLinkDocument{ResourceKind: "TASK", ResourceID: task.ID.String(),
		FormID: reference.ID.String(), VersionID: reference.VersionID.String(),
		PermissionRevision: reference.PermissionRevision}
	id := derivedID("form-link", task.ID.String(), reference.ID.String(), reference.VersionID.String())
	return appendTarget(plan, source, TargetFormLink, id, document)
}

func transformLabels(plan *Plan, source SourceRecord, task LegacyTask) error {
	labels := []string{}
	if len(task.Labels) != 0 && json.Unmarshal(task.Labels, &labels) != nil {
		appendIssue(plan, source, IssueFutureWork)
		return nil
	}
	seen := map[string]bool{}
	for _, label := range labels {
		if label == "" || strings.TrimSpace(label) != label || len(label) > 100 || seen[label] {
			appendIssue(plan, source, IssueFutureWork)
			continue
		}
		seen[label] = true
		labelID := derivedID("label", task.BoardID.String(), label)
		if err := appendTarget(plan, source, TargetLabel, labelID,
			TargetLabelDocument{BoardID: task.BoardID.String(), Name: label, Color: "#64748b"}); err != nil {
			return err
		}
		linkID := derivedID("task-label", task.ID.String(), labelID.String())
		if err := appendTarget(plan, source, TargetTaskLabel, linkID,
			TargetTaskLabelDocument{TaskID: task.ID.String(), LabelID: labelID.String()}); err != nil {
			return err
		}
	}
	return nil
}

func transformChecklist(plan *Plan, source SourceRecord, task LegacyTask) error {
	items := []LegacyChecklistItem{}
	if len(task.Checklist) != 0 && json.Unmarshal(task.Checklist, &items) != nil {
		appendIssue(plan, source, IssueFutureWork)
		return nil
	}
	seen := map[domain.ID]bool{}
	for position, item := range items {
		id, err := domain.ParseID(item.ID)
		if err != nil || seen[id] || strings.TrimSpace(item.Text) == "" || len(item.Text) > 1000 {
			appendIssue(plan, source, IssueFutureWork)
			continue
		}
		seen[id] = true
		document := TargetChecklistDocument{TaskID: task.ID.String(), Text: item.Text,
			Completed: item.Checked, Position: int64(position+1) * domain.SparseGap}
		if err = appendTarget(plan, source, TargetChecklist, id, document); err != nil {
			return err
		}
	}
	return nil
}
