package recoveryanchor

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/recoveryanchor"
	anchorports "github.com/Onurlulardan/harpy/core/internal/modules/iam/ports/recoveryanchor"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type Service struct {
	store     anchorports.Store[platformtx.TenantScope]
	authority anchorports.AuthorityPlanner
}

func NewService(
	store anchorports.Store[platformtx.TenantScope],
	authority anchorports.AuthorityPlanner,
) (*Service, error) {
	if store == nil || authority == nil {
		return nil, fmt.Errorf("recovery anchor service dependencies are incomplete")
	}
	return &Service{store: store, authority: authority}, nil
}

func scopeValid(scope platformtx.TenantScope, subject recoveryanchor.Subject) bool {
	return subject.Valid() && scope.ActorID() == subject.ActorID &&
		scope.TenantID() == subject.TenantID
}
