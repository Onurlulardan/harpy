package serviceprincipal

import (
	"time"

	"github.com/google/uuid"
)

type GrantState string

const (
	GrantActive  GrantState = "active"
	GrantRevoked GrantState = "revoked"
)

type GrantRevision struct {
	GrantID          uuid.UUID
	Revision         int64
	PrincipalID      uuid.UUID
	Capability       CapabilityKey
	Scope            ResourceScope
	ValidFrom        time.Time
	ValidUntil       time.Time
	State            GrantState
	ManifestSHA256   [32]byte
	PreviousRevision int64
}

func (grant GrantRevision) Validate() error {
	emptyManifest := [32]byte{}
	if grant.GrantID == uuid.Nil || grant.PrincipalID == uuid.Nil ||
		grant.Revision < 1 || grant.PreviousRevision < 0 ||
		(grant.Revision == 1) != (grant.PreviousRevision == 0) ||
		(grant.Revision > 1 && grant.PreviousRevision != grant.Revision-1) ||
		grant.ValidFrom.IsZero() ||
		(!grant.ValidUntil.IsZero() && !grant.ValidUntil.After(grant.ValidFrom)) ||
		(grant.State != GrantActive && grant.State != GrantRevoked) ||
		grant.ManifestSHA256 == emptyManifest {
		return ErrInvalidGrant
	}
	if _, err := ParseCapabilityKey(string(grant.Capability)); err != nil {
		return err
	}
	return grant.Scope.Validate()
}
