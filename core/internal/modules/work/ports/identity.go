package ports

import "github.com/Onurlulardan/harpy/core/internal/modules/work/domain"

type ID = domain.ID
type TaskStatus = domain.TaskStatus
type TaskKind = domain.TaskKind
type TaskPriority = domain.TaskPriority
type Principal = domain.Principal
type PrincipalKind = domain.PrincipalKind
type ChecklistItem = domain.ChecklistItem
type CalendarWindow = domain.CalendarWindow
type Audience = domain.Audience
type BoardRole = domain.BoardRole
type Label = domain.Label
type SourceOwner = domain.SourceOwner

const (
	StatusOpen           = domain.StatusOpen
	StatusInProgress     = domain.StatusInProgress
	StatusInReview       = domain.StatusInReview
	StatusDone           = domain.StatusDone
	StatusCancelled      = domain.StatusCancelled
	TaskCustom           = domain.TaskCustom
	TaskFormFill         = domain.TaskFormFill
	TaskFormReview       = domain.TaskFormReview
	PriorityLow          = domain.PriorityLow
	PriorityMedium       = domain.PriorityMedium
	PriorityHigh         = domain.PriorityHigh
	PriorityUrgent       = domain.PriorityUrgent
	PrincipalUser        = domain.PrincipalUser
	AudiencePrivate      = domain.AudiencePrivate
	AudienceBoard        = domain.AudienceBoard
	AudienceOrganization = domain.AudienceOrganization
	SourceWork           = domain.SourceWork
	SourceWorkflow       = domain.SourceWorkflow
	SourceForms          = domain.SourceForms
)

func ParseID(value string) (ID, error) { return domain.ParseID(value) }
