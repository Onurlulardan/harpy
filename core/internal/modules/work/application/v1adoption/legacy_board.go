package v1adoption

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
)

type LegacyBoard struct {
	ID, OrganizationID, CreatedBy domain.ID
	Name, Description             string
	CreatedAt, UpdatedAt          time.Time
	DeletedAt                     *time.Time
}

type LegacyColumn struct {
	ID, BoardID               domain.ID
	Name, Color, LinkedStatus string
	Order                     int64
	CreatedAt, UpdatedAt      time.Time
}

type LegacyMembership struct {
	ID, BoardID, UserID  domain.ID
	Role                 string
	CreatedAt, UpdatedAt time.Time
}

type TargetBoardDocument struct {
	CreatorID, Name, Description string
	Lifecycle                    string
	CreatedAt, UpdatedAt         time.Time
	ArchivedAt                   *time.Time
}

type TargetColumnDocument struct {
	BoardID, Name, Color, Status string
	Position                     int64
	CreatedAt, UpdatedAt         time.Time
}

type TargetMembershipDocument struct {
	BoardID, PrincipalKind, PrincipalID, Role string
	CreatedAt, UpdatedAt                      time.Time
}
