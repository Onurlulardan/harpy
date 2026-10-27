package api

import "context"

type AirGapRequestArtifact struct {
	OperationID    string
	ProposalSHA256 string
	Bytes          []byte
	ExpiresAt      string
}

type AirGapAckImport struct {
	OperationID string
	AckBytes    []byte
}

type AirGapAuthority interface {
	PrepareInitialExport(context.Context, InitialAuthorityRequest) (AirGapRequestArtifact, error)
	PrepareChangeExport(context.Context, ProtectedChangeRequest) (AirGapRequestArtifact, error)
	PrepareSignerRotationExport(context.Context, SignerRotationRequest) (AirGapRequestArtifact, error)
	ImportAcknowledgement(context.Context, AirGapAckImport) (AuthorityReceipt, error)
}
