package v1adoption

import (
	"encoding/json"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
)

type LegacyTask struct {
	ID, BoardID, ColumnID, CreatedBy       domain.ID
	Title, Description, TaskType, Priority string
	Status, FormSlug, AssigneeType         string
	ColumnOrder                            int64
	FormDataID, AssigneeID, WorkflowTaskID *domain.ID
	DueDate, StartDate, CompletedAt        *time.Time
	RecurrenceEnd, DeletedAt               *time.Time
	IsRecurring                            bool
	RecurrenceRule                         string
	FormFilter, Labels, Checklist          json.RawMessage
	CreatedAt, UpdatedAt                   time.Time
}

type LegacyChecklistItem struct {
	ID      string `json:"id"`
	Text    string `json:"text"`
	Checked bool   `json:"checked"`
}

type TargetTaskDocument struct {
	BoardID, ColumnID, CreatorID, Kind, Status           string
	Priority, Title, Description                         string
	Position                                             int64
	StartsAt, DueAt, CompletedAt, CancelledAt, DeletedAt *time.Time
	IsRecurring                                          bool
	RecurrenceRule                                       string
	RecurrenceEnd                                        *time.Time
	FormFilter                                           json.RawMessage
	CreatedAt, UpdatedAt                                 *time.Time
}

type TargetAssigneeDocument struct {
	TaskID, PrincipalKind, PrincipalID string
}

type TargetLabelDocument struct {
	BoardID, Name, Color string
}

type TargetTaskLabelDocument struct{ TaskID, LabelID string }

type TargetChecklistDocument struct {
	TaskID, Text string
	Completed    bool
	Position     int64
}

type TargetFormLinkDocument struct {
	ResourceKind, ResourceID, FormID, VersionID string
	ReferenceEpoch, PermissionRevision          int64
}
