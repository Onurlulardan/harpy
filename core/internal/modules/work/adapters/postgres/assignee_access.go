package postgres

import (
	"context"
	_ "embed"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/google/uuid"
)

//go:embed sql/tenant/assignees_have_access.sql
var assigneesHaveAccessSQL string

func (store *Store) AssigneesHaveAccess(ctx context.Context, boardID domain.ID,
	principals []domain.Principal,
) (bool, error) {
	if !boardID.Valid() || len(principals) > 100 {
		return false, application.ErrInvalid
	}
	if err := store.check(ctx); err != nil {
		return false, err
	}
	kinds := make([]string, 0, len(principals))
	ids := make([]uuid.UUID, 0, len(principals))
	for _, principal := range principals {
		if !principal.Valid() {
			return false, application.ErrInvalid
		}
		kinds = append(kinds, string(principal.Kind))
		ids = append(ids, uuid.UUID(principal.ID))
	}
	var allowed bool
	err := store.database.QueryRow(ctx, assigneesHaveAccessSQL,
		boardID, kinds, ids).Scan(&allowed)
	return allowed, err
}
