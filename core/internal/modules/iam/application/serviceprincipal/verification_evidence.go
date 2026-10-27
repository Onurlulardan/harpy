package serviceprincipal

import (
	"time"

	"github.com/google/uuid"
)

type VerificationEvidence struct {
	PrincipalID          uuid.UUID
	PrincipalRevision    int64
	AuthorizationVersion int64
	GrantID              uuid.UUID
	GrantRevision        int64
	Capability           CapabilityKey
	Scope                ResourceScope
	ManifestSHA256       [32]byte
	VerifiedAt           time.Time
}

func (evidence VerificationEvidence) Validate() error {
	if evidence.PrincipalID == uuid.Nil || evidence.PrincipalRevision < 1 ||
		evidence.AuthorizationVersion < 1 || evidence.GrantID == uuid.Nil ||
		evidence.GrantRevision < 1 || evidence.ManifestSHA256 == [32]byte{} ||
		evidence.VerifiedAt.IsZero() || evidence.VerifiedAt.Location() != time.UTC ||
		evidence.Scope.Validate() != nil {
		return ErrEvidenceMissing
	}
	_, err := ParseCapabilityKey(string(evidence.Capability))
	return err
}

func verifiedEvidence(
	principal Principal,
	grant GrantRevision,
	verifiedAt time.Time,
) VerificationEvidence {
	return VerificationEvidence{
		PrincipalID: principal.ID, PrincipalRevision: principal.Revision,
		AuthorizationVersion: principal.AuthorizationVersion,
		GrantID:              grant.GrantID, GrantRevision: grant.Revision,
		Capability: grant.Capability, Scope: grant.Scope,
		ManifestSHA256: grant.ManifestSHA256, VerifiedAt: verifiedAt,
	}
}
