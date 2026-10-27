package systemtransfer

import (
	"bytes"
	"context"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

const ExportOutcomeUnknownCode = "JOB_OUTCOME_UNKNOWN"

type ExportArtifactBeginEvidence struct {
	EffectID      uuid.UUID
	ObjectKey     string
	RequestSHA256 [32]byte
	StartedAt     time.Time
	Draft         ExportArtifactDraftEvidence
}

func (evidence ExportArtifactBeginEvidence) Validate() error {
	if !exportV7(evidence.EffectID) || len(evidence.ObjectKey) < 1 ||
		len(evidence.ObjectKey) > 512 || !utf8.ValidString(evidence.ObjectKey) ||
		evidence.RequestSHA256 == [32]byte{} || !exportArtifactTime(evidence.StartedAt) ||
		evidence.Draft.Validate() != nil {
		return ErrExportArtifactTampered
	}
	return nil
}

type ExportArtifactDraftEvidence struct {
	KeyID                          string
	HeaderSHA256, CiphertextSHA256 [32]byte
	CiphertextBytes                int64
	WrappedDEK                     []byte
	WrapNonce                      [12]byte
}

func (draft ExportArtifactDraftEvidence) Validate() error {
	if len(draft.KeyID) < 1 || len(draft.KeyID) > 128 ||
		!utf8.ValidString(draft.KeyID) || draft.HeaderSHA256 == [32]byte{} ||
		draft.CiphertextSHA256 == [32]byte{} || draft.CiphertextBytes < 1 ||
		draft.CiphertextBytes > 2148532224 || len(draft.WrappedDEK) < 32 ||
		len(draft.WrappedDEK) > 1024 || draft.WrapNonce == [12]byte{} {
		return ErrExportArtifactTampered
	}
	return nil
}

type ExportArtifactEffectStart struct {
	Attempt         ExportAttempt
	EffectID        uuid.UUID
	EffectReceiptID uuid.UUID
	ObjectKey       string
	RequestSHA256   [32]byte
	StartedAt       time.Time
	Draft           ExportArtifactDraftEvidence
}

func (start ExportArtifactEffectStart) Validate() error {
	if start.Attempt.Validate() != nil || start.EffectID != start.Attempt.AttemptID ||
		!exportV7(start.EffectReceiptID) || start.EffectReceiptID == start.EffectID ||
		(ExportArtifactBeginEvidence{EffectID: start.EffectID,
			ObjectKey: start.ObjectKey, RequestSHA256: start.RequestSHA256,
			StartedAt: start.StartedAt, Draft: start.Draft}).Validate() != nil {
		return ErrExportArtifactTampered
	}
	return nil
}

type ExportOutcomeUnknownRecord struct {
	Effect        ExportArtifactEffectStart
	Artifact      ExportArtifactMetadata
	ArtifactKnown bool
	At            time.Time
	Code          string
}

func (record ExportOutcomeUnknownRecord) Validate() error {
	if record.Effect.Validate() != nil || record.Code != ExportOutcomeUnknownCode ||
		!exportArtifactTime(record.At) || record.At.Before(record.Effect.StartedAt) {
		return ErrExportWorkerInvalid
	}
	if record.ArtifactKnown {
		return exportArtifactMatches(record.Effect, record.Artifact)
	}
	if !emptyExportArtifactMetadata(record.Artifact) {
		return ErrExportArtifactTampered
	}
	return nil
}

type ExportArtifactReconciliationState string

const (
	ExportArtifactMissing ExportArtifactReconciliationState = "missing"
	ExportArtifactSealed  ExportArtifactReconciliationState = "sealed"
	ExportArtifactUnknown ExportArtifactReconciliationState = "unknown"
)

type ExportArtifactReconciliation struct {
	State             ExportArtifactReconciliationState
	Artifact          ExportArtifactMetadata
	ProviderRequestID string
	EvidenceSHA256    [32]byte
	CheckedAt         time.Time
}

func (result ExportArtifactReconciliation) Validate(
	effect ExportArtifactEffectStart,
) error {
	if effect.Validate() != nil || !exportArtifactTime(result.CheckedAt) ||
		result.CheckedAt.Before(effect.StartedAt) {
		return ErrExportWorkerInvalid
	}
	switch result.State {
	case ExportArtifactSealed:
		if len(result.ProviderRequestID) < 1 ||
			len(result.ProviderRequestID) > 512 || !utf8.ValidString(result.ProviderRequestID) ||
			result.EvidenceSHA256 == [32]byte{} {
			return ErrExportArtifactTampered
		}
		return exportArtifactMatches(effect, result.Artifact)
	case ExportArtifactMissing:
		if !emptyExportArtifactMetadata(result.Artifact) {
			return ErrExportArtifactTampered
		}
		if len(result.ProviderRequestID) < 1 ||
			len(result.ProviderRequestID) > 512 || !utf8.ValidString(result.ProviderRequestID) ||
			result.EvidenceSHA256 == [32]byte{} {
			return ErrExportArtifactTampered
		}
		return nil
	case ExportArtifactUnknown:
		if !emptyExportArtifactMetadata(result.Artifact) ||
			(result.ProviderRequestID == "") != (result.EvidenceSHA256 == [32]byte{}) {
			return ErrExportArtifactTampered
		}
		if result.ProviderRequestID != "" && (len(result.ProviderRequestID) > 512 ||
			!utf8.ValidString(result.ProviderRequestID)) {
			return ErrExportArtifactTampered
		}
		return nil
	default:
		return ErrExportWorkerInvalid
	}
}

func emptyExportArtifactMetadata(metadata ExportArtifactMetadata) bool {
	return metadata.ObjectKey == "" && metadata.ObjectVersion == "" &&
		metadata.KeyID == "" && metadata.ProviderRequestID == "" &&
		metadata.EffectID == uuid.Nil && metadata.EffectReceiptID == uuid.Nil &&
		metadata.HeaderSHA256 == [32]byte{} &&
		metadata.CiphertextSHA256 == [32]byte{} &&
		metadata.RequestSHA256 == [32]byte{} && metadata.ResultSHA256 == [32]byte{} &&
		metadata.CiphertextBytes == 0 && len(metadata.WrappedDEK) == 0 &&
		metadata.WrapNonce == [12]byte{} && metadata.AttemptGeneration == 0 &&
		metadata.SealedAt.IsZero()
}

func exportDraftMatchesMetadata(
	draft ExportArtifactDraftEvidence,
	metadata ExportArtifactMetadata,
) bool {
	return metadata.KeyID == draft.KeyID &&
		metadata.HeaderSHA256 == draft.HeaderSHA256 &&
		metadata.CiphertextSHA256 == draft.CiphertextSHA256 &&
		metadata.CiphertextBytes == draft.CiphertextBytes &&
		bytes.Equal(metadata.WrappedDEK, draft.WrappedDEK) &&
		metadata.WrapNonce == draft.WrapNonce
}

type ExportArtifactOutcomeReconciler interface {
	ReconcileExportArtifact(
		context.Context,
		ExportArtifactEffectStart,
	) (ExportArtifactReconciliation, error)
}
