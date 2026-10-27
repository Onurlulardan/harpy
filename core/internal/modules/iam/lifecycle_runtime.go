package iam

import (
	lifecyclepostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
)

type LifecycleRuntime struct{ service *lifecycle.Service }

func NewLifecycleRuntime(catalogs platformasync.WriteCatalogs,
	now LifecycleClock) (*LifecycleRuntime, error) {
	store, err := lifecyclepostgres.NewStore()
	if err != nil {
		return nil, err
	}
	sink, err := NewLifecycleChangeSink(catalogs, now)
	if err != nil {
		return nil, err
	}
	service, err := lifecycle.NewService(store, sink)
	if err != nil {
		return nil, err
	}
	return &LifecycleRuntime{service: service}, nil
}

func (runtime *LifecycleRuntime) Valid() bool {
	return runtime != nil && runtime.service != nil && runtime.service.Valid()
}
