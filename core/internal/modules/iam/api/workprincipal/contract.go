package workprincipal

import (
	"context"
	"encoding/json"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type Principal struct {
	Kind     string
	ID       [16]byte
	Revision int64
}

type Evidence struct {
	Principals            []Principal
	AuthorizationRevision int64
	TargetExists          bool
	TargetRevision        int64
}

type Directory interface {
	Evidence(context.Context, platformtx.TenantScope, string, [16]byte) (Evidence, error)
	AvailableUsers(context.Context, platformtx.TenantScope, string, int) (json.RawMessage, error)
}
