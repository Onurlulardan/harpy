package workhttp

import nethttp "net/http"

func collaborationRoutes() []routeSpec {
	return []routeSpec{
		read(nethttp.MethodGet, "/api/work/tasks/{task_id}/comments", "listWorkComments", "work.comment.list", "comments", true),
		write(nethttp.MethodPost, "/api/work/tasks/{task_id}/comments", "createWorkComment", "work.comment.create", "comments", true),
		write(nethttp.MethodPost, "/api/work/comments/{comment_id}/redactions", "redactWorkComment", "work.comment.redact", "comments", false),
		write(nethttp.MethodPost, "/api/work/labels", "createWorkLabel", "work.label.change", "labels", true),
		write(nethttp.MethodPatch, "/api/work/labels/{label_id}", "updateWorkLabel", "work.label.change", "labels", false),
		write(nethttp.MethodPut, "/api/work/tasks/{task_id}/labels/{label_id}", "setWorkTaskLabel", "work.label.change", "labels", false),
		write(nethttp.MethodPost, "/api/work/tasks/{task_id}/checklist-items", "createWorkChecklistItem", "work.checklist.change", "checklists", true),
		write(nethttp.MethodPatch, "/api/work/checklist-items/{checklist_id}", "updateWorkChecklistItem", "work.checklist.change", "checklists", false),
		read(nethttp.MethodGet, "/api/work/calendar-events", "listWorkCalendarEvents", "work.calendar.list", "calendar", true),
		write(nethttp.MethodPost, "/api/work/calendar-events", "createWorkCalendarEvent", "work.calendar.change", "calendar", true),
		read(nethttp.MethodGet, "/api/work/calendar-events/{event_id}", "getWorkCalendarEvent", "work.calendar.get", "calendar", false),
		write(nethttp.MethodPatch, "/api/work/calendar-events/{event_id}", "updateWorkCalendarEvent", "work.calendar.change", "calendar", false),
		write(nethttp.MethodPost, "/api/work/recurrence-rules", "createWorkRecurrenceRule", "work.recurrence.change", "recurrence", true),
		write(nethttp.MethodPatch, "/api/work/recurrence-rules/{rule_id}", "updateWorkRecurrenceRule", "work.recurrence.change", "recurrence", false),
		read(nethttp.MethodGet, "/api/work/inbox", "listWorkInbox", "work.inbox.list", "inbox", true),
		read(nethttp.MethodGet, "/api/work/calendar", "listWorkCalendar", "work.calendar.list", "calendar-unified", true),
	}
}
