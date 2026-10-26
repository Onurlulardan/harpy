package v1adoption

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	"github.com/google/uuid"
)

type Store interface {
	BeginAndLock(context.Context, Plan) (Run, error)
	SourceExists(context.Context) (bool, error)
	LoadSourcePage(context.Context, *uuid.UUID, int) (SourcePage, error)
	Adopt(context.Context, MappedRecord) ([32]byte, bool, error)
	LogIssue(context.Context, uuid.UUID, uuid.UUID, [32]byte, string) error
	SaveCheckpoint(context.Context, uuid.UUID, Checkpoint) error
	SealRun(context.Context, uuid.UUID, Checkpoint) error
}

type Binder interface {
	BindAdoption(context.Context, api.Scope) (Store, error)
}
