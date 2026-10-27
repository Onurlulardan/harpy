package jobs

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
	"github.com/google/uuid"
)

type runtimeIDs struct{}

func (runtimeIDs) New(time.Time) (domain.ID, error) {
	id, err := uuid.NewV7()
	return domain.ID(id), err
}

type runtimeTZDB struct{}

func (runtimeTZDB) Load(ctx context.Context, name, version string) (*time.Location, error) {
	if ctx == nil || ctx.Err() != nil || name == "" || version == "" {
		return nil, ports.ErrInvalid
	}
	location, err := time.LoadLocation(name)
	if err != nil {
		return nil, ports.ErrUnavailable
	}
	return location, nil
}

func NewRuntimeRunner(store Store) (*Runner, error) {
	return NewRunner(store, runtimeTZDB{}, runtimeIDs{})
}
