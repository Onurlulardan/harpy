package ports

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/domain/artifact"
)

type Store interface {
	LoadBinding(context.Context, string) (BindingRecord, error)
	SaveBinding(context.Context, BindingRecord, *artifact.VerifiedBindingProof) error
	LoadOperation(context.Context, string) (OperationRecord, error)
	PrepareOperation(context.Context, OperationRecord) (OperationRecord, error)
	PersistAcknowledgement(context.Context, string, artifact.VerifiedAck) (OperationRecord, error)
	CommitAcknowledged(context.Context, string) (CommitResult, error)
	ApplySignerRotation(context.Context, SignerRotationRecord) error
	FenceSigner(context.Context, SignerFenceRecord) error
	LoadHead(context.Context, HeadKey) (ExternalHead, error)
}
