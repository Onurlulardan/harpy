package systemtransfer

import (
	"time"

	"github.com/google/uuid"
)

type SourceArtifact struct {
	ObjectKey, ObjectVersion, KeyID, ProviderRequestID string
	HeaderSHA256, CiphertextSHA256, PlaintextSHA256    [32]byte
	CiphertextBytes, PlaintextBytes                    int64
	WrappedDEK                                         []byte
	WrapNonce                                          [12]byte
	UploadedAt                                         time.Time
}

func (value SourceArtifact) Validate() error {
	if value.ObjectKey == "" || value.ObjectVersion == "" || value.KeyID == "" ||
		value.ProviderRequestID == "" || value.HeaderSHA256 == [32]byte{} ||
		value.CiphertextSHA256 == [32]byte{} || value.PlaintextSHA256 == [32]byte{} ||
		value.CiphertextBytes < 1 || value.CiphertextBytes > 538968064 ||
		value.PlaintextBytes < 1 || value.PlaintextBytes > 536870912 ||
		len(value.WrappedDEK) < 32 || len(value.WrappedDEK) > 1024 ||
		value.WrapNonce == [12]byte{} || !controlTime(value.UploadedAt) {
		return ErrControlInvalid
	}
	return nil
}

type SourceWrite struct {
	Evidence             ControlEvidence
	SourceID, ReceiptID  uuid.UUID
	IdempotencyKeySHA256 [32]byte
	Format               string
	Artifact             SourceArtifact
}

func (value SourceWrite) Validate() error {
	if value.Evidence.Validate() != nil || !controlV7(value.SourceID) ||
		!controlV7(value.ReceiptID) || value.IdempotencyKeySHA256 == [32]byte{} ||
		(value.Format != "csv" && value.Format != "jsonl") ||
		value.Artifact.Validate() != nil {
		return ErrControlInvalid
	}
	return nil
}

type SourceRecord struct {
	SourceID        uuid.UUID
	Resource        string
	Format          string
	PlaintextBytes  int64
	PlaintextSHA256 [32]byte
	UploadedAt      time.Time
	Replay          bool
	Analysis        SourceAnalysis
}
