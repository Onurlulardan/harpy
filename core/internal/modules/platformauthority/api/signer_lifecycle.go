package api

import (
	"context"
	"time"
)

type SignerRotationRequest struct {
	OperationID              string
	State                    State
	NewSignerPublicKeySHA256 string
	RecoveryTrustBundle      []byte
	IssuedAt                 time.Time
	Policy                   ExpiryPolicy
}

type SignerFenceRequest struct {
	RegistryBindingID        string
	ExpectedSignerGeneration int64
	ExpectedSignerKID        string
	EvidenceSHA256           string
}

type SignerLifecycle interface {
	PrepareSignerRotation(context.Context, SignerRotationRequest) (PreparedAuthority, error)
	AcknowledgeSignerRotation(context.Context, string) (AuthorityReceipt, error)
	FenceSigner(context.Context, SignerFenceRequest) error
}
