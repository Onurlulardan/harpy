package ports

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
)

type BoardStore interface {
	InsertBoard(context.Context, domain.Board) error
	LoadBoardForUpdate(context.Context, domain.ID) (domain.Board, error)
	PersistBoard(context.Context, domain.Board, int64) error
	InsertColumn(context.Context, domain.Column) error
	LoadColumnForUpdate(context.Context, domain.ID) (domain.Column, error)
	LoadColumnsForUpdate(context.Context, domain.ID) ([]domain.Column, error)
	PersistColumn(context.Context, domain.Column, int64) error
	DeleteColumn(context.Context, domain.ID, int64) error
	AssigneesHaveAccess(context.Context, domain.ID, []domain.Principal) (bool, error)
	LoadMembershipsForUpdate(context.Context, domain.ID) ([]domain.Membership, error)
	UpsertMembership(context.Context, domain.Membership) error
	DeactivateMembership(context.Context, domain.ID, int64) error
}
