package lifecycle

import "github.com/google/uuid"

type PortalApprovalDecision string

const (
	PortalApprovalApproved PortalApprovalDecision = "approved"
	PortalApprovalRejected PortalApprovalDecision = "rejected"
)

func (value PortalApprovalDecision) Valid() bool {
	return value == PortalApprovalApproved || value == PortalApprovalRejected
}

type PortalSetupStartCommand struct {
	Evidence  MutationEvidence
	BindingID uuid.UUID
	Token     TokenIssue
	Delivery  DeliveryArtifact
}

type PortalSetupCompleteCommand struct {
	Evidence CredentialEvidence
	Proof    CredentialProof
	Salt     [16]byte
	Tag      [32]byte
}

type PortalDecisionCommand struct {
	Evidence         MutationEvidence
	BindingID        uuid.UUID
	ExpectedRevision int64
	Decision         PortalApprovalDecision
}
