package work

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
	"github.com/Onurlulardan/harpy/core/internal/platform/randomid"
)

type runtimeClock struct{}

func (runtimeClock) Now() time.Time { return time.Now().UTC() }

type runtimeIDs struct{}

func (runtimeIDs) New(at time.Time) (ports.ID, error) {
	id, err := randomid.NewUUIDv7(at)
	return ports.ID(id), err
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
