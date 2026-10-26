package actorbridgeissue

import (
	"errors"

	actorbridgeissueapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/actorbridgeissue"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

var errUnavailable = errors.New("IAM actor bridge issue authority unavailable")

type boundStore struct {
	database platformpostgres.BoundExecutor
}

type store struct {
	binder platformpostgres.OwnerBinder[*boundStore]
	valid  bool
}

func NewStore() (actorbridgeissueapp.Store, error) {
	binder, err := platformpostgres.NewOwnerBinder(
		"iam", func(database platformpostgres.BoundExecutor) *boundStore {
			return &boundStore{database: database}
		},
	)
	if err != nil {
		return nil, errUnavailable
	}
	return &store{binder: binder, valid: true}, nil
}

func (store *store) bind(
	scope actorbridgeissueapp.TenantScope,
) (*boundStore, error) {
	if store == nil || !store.valid {
		return nil, errUnavailable
	}
	bound, err := platformtx.Bind(scope, store.binder)
	if err != nil || bound == nil {
		return nil, errUnavailable
	}
	return bound, nil
}

var _ actorbridgeissueapp.Store = (*store)(nil)
