package retention

import "github.com/google/uuid"

type Claim struct {
	JobID            uuid.UUID
	AttemptID        uuid.UUID
	TenantID         uuid.UUID
	Kind             string
	PayloadRevision  int
	PayloadReference string
	PayloadSHA256    string
}

func (claim Claim) Check() error {
	if claim.JobID == uuid.Nil || claim.JobID.Version() != uuid.Version(7) ||
		claim.AttemptID == uuid.Nil || claim.AttemptID.Version() != uuid.Version(7) ||
		claim.TenantID == uuid.Nil || claim.Kind != JobKind ||
		claim.PayloadRevision != PayloadRevision {
		return ErrClaimInvalid
	}
	reference, digest, err := Reference(claim.TenantID)
	if err != nil || claim.PayloadReference != reference || claim.PayloadSHA256 != digest {
		return ErrClaimInvalid
	}
	return nil
}
