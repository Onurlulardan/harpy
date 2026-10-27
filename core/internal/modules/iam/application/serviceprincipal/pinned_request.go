package serviceprincipal

import "github.com/google/uuid"

type PinnedRequest struct {
	PrincipalID   uuid.UUID
	Capability    CapabilityKey
	Scope         ResourceScope
	GrantID       uuid.UUID
	GrantRevision int64
	Manifest      [32]byte
}

func (request PinnedRequest) Validate() error {
	if request.PrincipalID == uuid.Nil || request.GrantID == uuid.Nil || request.GrantRevision < 1 ||
		request.Manifest == [32]byte{} {
		return ErrInvalidGrant
	}
	if _, err := ParseCapabilityKey(string(request.Capability)); err != nil {
		return err
	}
	return request.Scope.Validate()
}
