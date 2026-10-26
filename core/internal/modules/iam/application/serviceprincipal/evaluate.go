package serviceprincipal

import (
	"time"

	"github.com/google/uuid"
)

type AuthorizationRequest struct {
	Capability    CapabilityKey
	Scope         ResourceScope
	GrantID       uuid.UUID
	GrantRevision int64
	Manifest      [32]byte
	At            time.Time
}

func Evaluate(
	principal Principal,
	grant GrantRevision,
	request AuthorizationRequest,
) error {
	if principal.Validate() != nil || grant.Validate() != nil ||
		principal.State != PrincipalActive || grant.PrincipalID != principal.ID {
		return ErrGrantInactive
	}
	if grant.State != GrantActive || request.At.IsZero() ||
		request.At.Before(grant.ValidFrom) ||
		(!grant.ValidUntil.IsZero() && !request.At.Before(grant.ValidUntil)) {
		return ErrGrantInactive
	}
	if request.GrantID != grant.GrantID || request.GrantRevision != grant.Revision ||
		request.Manifest != grant.ManifestSHA256 || request.Capability != grant.Capability ||
		!request.Scope.Equal(grant.Scope) {
		return ErrGrantMismatch
	}
	return nil
}
