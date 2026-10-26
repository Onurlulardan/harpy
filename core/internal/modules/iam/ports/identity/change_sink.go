package identity

import (
	"context"

	identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
)

type ChangeSink[Scope any] interface {
	WriteChange(context.Context, Scope, identitydomain.Change) error
}

type ChangeAuditStore[Scope any] interface {
	WriteAudit(
		context.Context, Scope, identitydomain.Change,
	) (identitydomain.RecordedChange, error)
}
