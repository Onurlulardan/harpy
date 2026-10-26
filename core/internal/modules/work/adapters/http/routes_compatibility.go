package workhttp

import nethttp "net/http"

func compatibilityRoutes() []routeSpec {
	routes := []routeSpec{
		read(nethttp.MethodGet, "/api/workflow-service/boards/available-users", "listLegacyAvailableUsers", "work.membership.change", "available-users", true),
		read(nethttp.MethodGet, "/api/workflow-service/board-tasks/my-tasks", "listLegacyMyTasks", "work.task.list", "inbox", true),
		read(nethttp.MethodGet, "/api/workflow-service/board-tasks/calendar", "listLegacyCalendarTasks", "work.task.list", "inbox", true),
		read(nethttp.MethodGet, "/api/workflow-service/boards", "listLegacyWorkBoards", "work.board.list", "boards", true),
		write(nethttp.MethodPost, "/api/workflow-service/boards", "createLegacyWorkBoard", "work.board.create", "boards", true),
		read(nethttp.MethodGet, "/api/workflow-service/boards/{board_id}", "getLegacyWorkBoard", "work.board.get", "boards", false),
		write(nethttp.MethodPut, "/api/workflow-service/boards/{board_id}", "updateLegacyWorkBoard", "work.board.update", "boards", false),
		write(nethttp.MethodDelete, "/api/workflow-service/boards/{board_id}", "archiveLegacyWorkBoard", "work.board.archive", "boards", false),
		read(nethttp.MethodGet, "/api/workflow-service/boards/{board_id}/members", "listLegacyBoardMembers", "work.board.get", "memberships", true),
		write(nethttp.MethodPost, "/api/workflow-service/boards/{board_id}/members", "createLegacyBoardMember", "work.membership.change", "memberships", true),
		write(nethttp.MethodPut, "/api/workflow-service/boards/{board_id}/members/{member_id}", "updateLegacyBoardMember", "work.membership.change", "memberships", false),
		write(nethttp.MethodDelete, "/api/workflow-service/boards/{board_id}/members/{member_id}", "deleteLegacyBoardMember", "work.membership.change", "memberships", false),
		read(nethttp.MethodGet, "/api/workflow-service/boards/{board_id}/my-role", "getLegacyBoardRole", "work.board.get", "my-role", false),
		write(nethttp.MethodPost, "/api/workflow-service/boards/{board_id}/columns", "createLegacyBoardColumn", "work.column.create", "columns", true),
		write(nethttp.MethodPut, "/api/workflow-service/boards/{board_id}/columns/reorder", "reorderLegacyBoardColumns", "work.column.update", "columns", false),
		write(nethttp.MethodPut, "/api/workflow-service/boards/{board_id}/columns/{column_id}", "updateLegacyBoardColumn", "work.column.update", "columns", false),
		write(nethttp.MethodDelete, "/api/workflow-service/boards/{board_id}/columns/{column_id}", "deleteLegacyBoardColumn", "work.column.update", "columns", false),
		write(nethttp.MethodPost, "/api/workflow-service/boards/{board_id}/tasks", "createLegacyBoardTask", "work.task.create", "tasks", true),
		read(nethttp.MethodGet, "/api/workflow-service/board-tasks", "listLegacyBoardTasks", "work.task.list", "tasks", true),
		write(nethttp.MethodPost, "/api/workflow-service/board-tasks", "createLegacyBoardTask", "work.task.create", "tasks", true),
		read(nethttp.MethodGet, "/api/workflow-service/board-tasks/{task_id}", "getLegacyBoardTask", "work.task.get", "tasks", false),
		write(nethttp.MethodPut, "/api/workflow-service/board-tasks/{task_id}", "updateLegacyBoardTask", "work.task.update", "tasks", false),
		write(nethttp.MethodDelete, "/api/workflow-service/board-tasks/{task_id}", "archiveLegacyBoardTask", "work.task.archive", "tasks", false),
		write(nethttp.MethodPut, "/api/workflow-service/board-tasks/{task_id}/move", "moveLegacyBoardTask", "work.task.move", "tasks", false),
		read(nethttp.MethodGet, "/api/workflow-service/board-tasks/{task_id}/comments", "listLegacyTaskComments", "work.comment.list", "comments", true),
		write(nethttp.MethodPost, "/api/workflow-service/board-tasks/{task_id}/comments", "createLegacyTaskComment", "work.comment.create", "comments", true),
		write(nethttp.MethodDelete, "/api/workflow-service/board-tasks/{task_id}/comments/{comment_id}", "deleteLegacyTaskComment", "work.comment.redact", "comments", false),
		read(nethttp.MethodGet, "/api/workflow-service/calendar-events", "listLegacyCalendarEvents", "work.calendar.list", "calendar", true),
		write(nethttp.MethodPost, "/api/workflow-service/calendar-events", "createLegacyCalendarEvent", "work.calendar.change", "calendar", true),
		read(nethttp.MethodGet, "/api/workflow-service/calendar-events/{event_id}", "getLegacyCalendarEvent", "work.calendar.get", "calendar", false),
		write(nethttp.MethodPut, "/api/workflow-service/calendar-events/{event_id}", "updateLegacyCalendarEvent", "work.calendar.change", "calendar", false),
		write(nethttp.MethodDelete, "/api/workflow-service/calendar-events/{event_id}", "archiveLegacyCalendarEvent", "work.calendar.change", "calendar", false),
	}
	for index := range routes {
		routes[index].compatibility = true
	}
	return routes
}
