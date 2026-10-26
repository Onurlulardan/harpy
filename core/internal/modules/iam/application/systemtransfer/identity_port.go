package systemtransfer

import (
	"context"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type IdentityMatch struct {
	Found    bool
	ID       publicresource.ID
	Revision int64
}

func (match IdentityMatch) Validate() error {
	if match.Found && (!match.ID.Valid() || match.Revision < 1) ||
		!match.Found && (match.ID.Valid() || match.Revision != 0) {
		return ErrInvalidAttempt
	}
	return nil
}

type IdentityLookup interface {
	FindIdentity(
		context.Context,
		platformtx.TenantScope,
		publicresource.ActorEvidence,
		IdentityManifest,
		publicresource.FieldValues,
	) (IdentityMatch, error)
}
