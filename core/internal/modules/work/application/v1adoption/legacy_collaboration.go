package v1adoption

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
)

type LegacyComment struct {
	ID, TaskID, UserID   domain.ID
	Content              string
	CreatedAt, UpdatedAt time.Time
	DeletedAt            *time.Time
}

type LegacyCalendar struct {
	ID, OrganizationID, CreatedBy domain.ID
	Title, Description, Color     string
	StartDate, EndDate            time.Time
	AllDay                        bool
	FormSlug, FormDataID          string
	CreatedAt, UpdatedAt          time.Time
	DeletedAt                     *time.Time
}

type TargetCommentDocument struct {
	BoardID, TaskID, AuthorID, Body                   string
	State, RedactionKind, RedactedBy, RedactionReason string
	CreatedAt, RedactedAt                             *time.Time
}

type TargetCalendarDocument struct {
	OwnerID, Audience, BoardID, OrganizationID        string
	Title, Description, Color, Timezone               string
	StartsAt, EndsAt, DeletedAt, CreatedAt, UpdatedAt *time.Time
	StartDate, EndDate                                string
}

type TargetRecurrenceDocument struct {
	EventID, DTStart, Timezone, TZDBVersion, RRULE string
	HorizonEnd                                     time.Time
}

type TargetProjectionDocument struct {
	SourceID, VersionID, Summary string
	Revision, PermissionRevision int64
	DueAt                        *time.Time
	VisibleToPrincipals          []string
}
