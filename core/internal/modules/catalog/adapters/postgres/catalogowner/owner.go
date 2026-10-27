package catalogowner

import (
	catalogcomponent "github.com/Onurlulardan/harpy/core/internal/capabilities/catalogcomponent/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/demomaterialization"
)

type Owner struct {
	*catalogcomponent.PersistentOwner
	*demomaterialization.Application
}

func New(
	ids demomaterialization.IDSource,
	clock demomaterialization.Clock,
) (*Owner, error) {
	return newOwner(ids, clock)
}

func newOwner(
	ids demomaterialization.IDSource,
	clock demomaterialization.Clock,
) (*Owner, error) {
	factory, err := catalogcomponent.NewStoreFactory(owner, catalogcomponent.Queries{
		Read: readSQL, Apply: applySQL, Blockers: blockersSQL, Allows: allowsSQL,
	})
	if err != nil {
		return nil, err
	}
	persistent, err := catalogcomponent.NewPersistentOwner(
		owner, &binder{factory: factory},
	)
	if err != nil {
		return nil, err
	}
	provision, err := demomaterialization.New(&provisionBinder{}, ids, clock)
	if err != nil {
		return nil, err
	}
	return &Owner{PersistentOwner: persistent, Application: provision}, nil
}
