package workhttp

import nethttp "net/http"

func canonicalRoutes() []routeSpec {
	routes := []routeSpec{
		read(nethttp.MethodGet, "/api/work/boards", "listWorkBoards", "work.board.list", "boards", true),
		write(nethttp.MethodPost, "/api/work/boards", "createWorkBoard", "work.board.create", "boards", true),
		read(nethttp.MethodGet, "/api/work/boards/{board_id}", "getWorkBoard", "work.board.get", "boards", false),
		write(nethttp.MethodPatch, "/api/work/boards/{board_id}", "updateWorkBoard", "work.board.update", "boards", false),
		write(nethttp.MethodPost, "/api/work/boards/{board_id}/archive", "archiveWorkBoard", "work.board.archive", "boards", false),
		write(nethttp.MethodPost, "/api/work/boards/{board_id}/restore", "restoreWorkBoard", "work.board.restore", "boards", false),
		write(nethttp.MethodPost, "/api/work/boards/{board_id}/columns", "createWorkColumn", "work.column.create", "columns", true),
		write(nethttp.MethodPut, "/api/work/boards/{board_id}/memberships", "changeWorkMembership", "work.membership.change", "memberships", false),
		write(nethttp.MethodPatch, "/api/work/columns/{column_id}", "updateWorkColumn", "work.column.update", "columns", false),
		read(nethttp.MethodGet, "/api/work/tasks", "listWorkTasks", "work.task.list", "tasks", true),
		write(nethttp.MethodPost, "/api/work/tasks", "createWorkTask", "work.task.create", "tasks", true),
		read(nethttp.MethodGet, "/api/work/tasks/{task_id}", "getWorkTask", "work.task.get", "tasks", false),
		write(nethttp.MethodPatch, "/api/work/tasks/{task_id}", "updateWorkTask", "work.task.update", "tasks", false),
		write(nethttp.MethodPost, "/api/work/tasks/{task_id}/moves", "moveWorkTask", "work.task.move", "tasks", false),
		write(nethttp.MethodPut, "/api/work/tasks/{task_id}/assignments", "assignWorkTask", "work.task.assign", "tasks", false),
		write(nethttp.MethodPost, "/api/work/tasks/{task_id}/completions", "completeWorkTask", "work.task.transition", "tasks", false),
		write(nethttp.MethodPost, "/api/work/tasks/{task_id}/cancellations", "cancelWorkTask", "work.task.transition", "tasks", false),
		write(nethttp.MethodPost, "/api/work/tasks/{task_id}/reopens", "reopenWorkTask", "work.task.transition", "tasks", false),
		write(nethttp.MethodPost, "/api/work/tasks/{task_id}/archive", "archiveWorkTask", "work.task.archive", "tasks", false),
		write(nethttp.MethodPost, "/api/work/tasks/{task_id}/restore", "restoreWorkTask", "work.task.restore", "tasks", false),
	}
	return append(routes, collaborationRoutes()...)
}
