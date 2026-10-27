package maintenance

import "github.com/google/uuid"

type Claim struct {
	JobID            uuid.UUID
	AttemptID        uuid.UUID
	TenantID         uuid.UUID
	Kind             Kind
	PayloadRevision  int
	PayloadReference string
	PayloadSHA256    string
}

func (claim Claim) Validate() error {
	if !stableID(claim.JobID) || !stableID(claim.AttemptID) ||
		!stableID(claim.TenantID) || !claim.Kind.Valid() ||
		claim.PayloadRevision != PayloadRevision {
		return ErrClaimInvalid
	}
	if claim.Kind.Recurring() {
		reference, digest, err := RecurringReference(claim.TenantID, claim.Kind)
		if err != nil || reference != claim.PayloadReference ||
			digest != claim.PayloadSHA256 {
			return ErrClaimInvalid
		}
		return nil
	}
	tenantID, _, _, valid := parseProfileReference(claim.PayloadReference)
	if !valid || tenantID != claim.TenantID ||
		digestReference(claim.PayloadReference) != claim.PayloadSHA256 {
		return ErrClaimInvalid
	}
	return nil
}
