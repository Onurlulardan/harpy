package permissioncache

import "context"

type Store interface {
	Purge(context.Context, Command) (Result, error)
}

type Application struct{ store Store }

func New(store Store) (*Application, error) {
	if store == nil {
		return nil, ErrUnavailable
	}
	return &Application{store: store}, nil
}
