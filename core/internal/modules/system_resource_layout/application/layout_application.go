package application

import layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"

type LayoutApplication struct {
	store Store
	ids   IDSource
	clock Clock
}

func NewLayoutApplication(
	store Store,
	ids IDSource,
	clock Clock,
) (layoutapi.LayoutAPI, error) {
	if store == nil || ids == nil || clock == nil {
		return nil, layoutapi.ErrInvalid
	}
	return &LayoutApplication{store: store, ids: ids, clock: clock}, nil
}
