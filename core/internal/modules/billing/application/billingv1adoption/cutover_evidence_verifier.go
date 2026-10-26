package billingv1adoptionapp

import "context"

type CutoverEvidenceVerifier interface {
	Verify(context.Context, Operation, ScopePlan) (CutoverEvidence, error)
}
