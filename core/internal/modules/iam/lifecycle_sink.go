package iam

import (
	"time"

	lifecycleevent "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/lifecycleevent"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
)

const lifecyclePayloadSchemaSHA256 = "f5100cde78ae1d5057c538d22ee199765816f6648650a19f7f11526a46ff8230"

type LifecycleClock func() time.Time

type LifecycleChangeSink struct {
	sequences lifecycleevent.Sequences
	catalogs  platformasync.WriteCatalogs
	now       LifecycleClock
}

func NewLifecycleChangeSink(catalogs platformasync.WriteCatalogs,
	now LifecycleClock) (*LifecycleChangeSink, error) {
	if now == nil {
		return nil, lifecycle.ErrEffects
	}
	if _, _, err := catalogs.Snapshot(); err != nil {
		return nil, lifecycle.ErrEffects
	}
	sequences, err := lifecycleevent.NewStore()
	if err != nil {
		return nil, err
	}
	return &LifecycleChangeSink{sequences: sequences, catalogs: catalogs, now: now}, nil
}

var _ lifecycle.Effects = (*LifecycleChangeSink)(nil)
