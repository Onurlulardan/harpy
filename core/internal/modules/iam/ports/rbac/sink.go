package rbac

import (
	"context"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
)

type ChangeSink[Scope any] interface {
	WriteChange(context.Context, Scope, rbacdomain.Change) error
}

type AuditStore[Scope any] interface {
	WriteAudit(context.Context, Scope, rbacdomain.Change) error
}
