package v1cutover

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
)

type Store interface {
	CreateAuthority(context.Context, Origin) (Authority, error)
	PauseV1(context.Context, Authority) (Authority, error)
	RecordEvidence(context.Context, Inventory) (Evidence, error)
	DisableV1(context.Context, Authority, Evidence) (Authority, error)
	ActivateV2(context.Context, Authority) (Authority, error)
}

type Binder interface {
	BindCutover(context.Context, api.Scope) (Store, error)
}
