package v1adoption

import "strings"

func transformComments(plan *Plan, index legacyIndex) error {
	for id, comment := range index.comments {
		source := index.source(KindComment, id)
		task, taskExists := index.tasks[comment.TaskID]
		if !taskExists {
			appendIssue(plan, source, IssueOrphan)
			continue
		}
		if task.WorkflowTaskID != nil || task.TaskType == "APPROVAL" {
			appendIssue(plan, source, IssueFutureWork)
			continue
		}
		if !comment.UserID.Valid() || strings.TrimSpace(comment.Content) == "" ||
			len(comment.Content) > 16_000 || comment.CreatedAt.IsZero() {
			appendIssue(plan, source, IssueFutureWork)
			continue
		}
		document := TargetCommentDocument{BoardID: task.BoardID.String(),
			TaskID: task.ID.String(), AuthorID: comment.UserID.String(),
			Body: comment.Content, State: "VISIBLE", CreatedAt: utcPointer(&comment.CreatedAt)}
		if comment.DeletedAt != nil {
			document.Body, document.State, document.RedactionKind = "", "REDACTED", "AUTHOR"
			document.RedactedBy = comment.UserID.String()
			document.RedactedAt = utcPointer(comment.DeletedAt)
		}
		if err := appendTarget(plan, source, TargetComment, id, document); err != nil {
			return err
		}
	}
	return nil
}
