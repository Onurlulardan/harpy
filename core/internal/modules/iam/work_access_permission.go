package iam

type workAccessTarget struct{ resource, action string }

var workAccessTargets = map[string]workAccessTarget{
	"work.board.list":        {"work.boards", "read"},
	"work.board.get":         {"work.boards", "read"},
	"work.board.create":      {"work.boards", "create"},
	"work.board.update":      {"work.boards", "update"},
	"work.board.archive":     {"work.boards", "archive"},
	"work.board.restore":     {"work.boards", "restore"},
	"work.column.create":     {"work.columns", "create"},
	"work.column.update":     {"work.columns", "update"},
	"work.membership.change": {"work.memberships", "manage"},
	"work.task.list":         {"work.tasks", "read"},
	"work.task.get":          {"work.tasks", "read"},
	"work.task.create":       {"work.tasks", "create"},
	"work.task.update":       {"work.tasks", "update"},
	"work.task.move":         {"work.tasks", "move"},
	"work.task.assign":       {"work.tasks", "assign"},
	"work.task.transition":   {"work.tasks", "transition"},
	"work.task.archive":      {"work.tasks", "archive"},
	"work.task.restore":      {"work.tasks", "restore"},
	"work.comment.list":      {"work.comments", "read"},
	"work.comment.create":    {"work.comments", "create"},
	"work.comment.redact":    {"work.comments", "redact"},
	"work.label.change":      {"work.labels", "manage"},
	"work.checklist.change":  {"work.checklists", "manage"},
	"work.calendar.list":     {"work.calendar", "read"},
	"work.calendar.get":      {"work.calendar", "read"},
	"work.calendar.change":   {"work.calendar", "manage"},
	"work.recurrence.change": {"work.recurrence", "manage"},
	"work.inbox.list":        {"work.inbox", "read"},
}

func resolveWorkAccess(authority, operation string) (workAccessTarget, bool) {
	target, found := workAccessTargets[authority]
	if !found || operation == "" || !workOperationAllowed(authority, operation) {
		return workAccessTarget{}, false
	}
	return target, true
}

func workOperationAllowed(authority, operation string) bool {
	allowed, found := workOperations[authority]
	if !found {
		return false
	}
	for _, candidate := range allowed {
		if candidate == operation {
			return true
		}
	}
	return false
}
