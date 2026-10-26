package dashboardexport

import (
	"time"

	"github.com/google/uuid"
)

const (
	OperationCreate             = "analytics.dashboard.export.create"
	OperationGet                = "analytics.dashboard.export.get"
	OperationDelete             = "analytics.dashboard.export.delete"
	OperationContent            = "analytics.dashboard.export.content"
	OperationWorker             = "analytics.dashboard.export.worker"
	PermissionExport Permission = "analytics.dashboard.export"
)

type Permission string

type JobAction string

const (
	JobRender JobAction = "render"
	JobDelete JobAction = "delete"
	JobExpire JobAction = "expire"
)

type Job struct {
	JobID, ExportID uuid.UUID
	Action          JobAction
	RunAt           time.Time
}
