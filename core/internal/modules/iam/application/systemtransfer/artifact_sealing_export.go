package systemtransfer

import (
	"context"
	"io"
	"time"
	"unicode/utf8"

	tabulartransfer "github.com/Onurlulardan/harpy/core/internal/capabilities/tabulartransfer/api"
	"github.com/google/uuid"
)

type ExportArtifactSealRequest struct {
	TenantID, ExportID    uuid.UUID
	EffectID              uuid.UUID
	EffectReceiptID       uuid.UUID
	AttemptGeneration     int64
	Format                string
	Plaintext             io.Reader
	MaximumPlaintextBytes int64
}

func (request ExportArtifactSealRequest) Validate() error {
	if request.TenantID == uuid.Nil || !exportV7(request.ExportID) ||
		!exportV7(request.EffectID) || !exportV7(request.EffectReceiptID) ||
		request.EffectID == request.EffectReceiptID ||
		request.AttemptGeneration < 1 || !validExportFormat(request.Format) ||
		request.Plaintext == nil || request.MaximumPlaintextBytes < 1 ||
		request.MaximumPlaintextBytes > tabulartransfer.MaximumDecompressedBytes {
		return ErrExportWorkerInvalid
	}
	return nil
}

type ExportArtifactMetadata struct {
	ObjectKey, ObjectVersion, KeyID string
	ProviderRequestID               string
	EffectID, EffectReceiptID       uuid.UUID
	HeaderSHA256, CiphertextSHA256  [32]byte
	RequestSHA256, ResultSHA256     [32]byte
	CiphertextBytes                 int64
	WrappedDEK                      []byte
	WrapNonce                       [12]byte
	AttemptGeneration               int64
	SealedAt                        time.Time
}

func (metadata ExportArtifactMetadata) Validate() error {
	if len(metadata.ObjectKey) < 1 || len(metadata.ObjectKey) > 512 ||
		len(metadata.ObjectVersion) < 1 || len(metadata.ObjectVersion) > 512 ||
		len(metadata.KeyID) < 1 || len(metadata.KeyID) > 128 ||
		len(metadata.ProviderRequestID) < 1 || len(metadata.ProviderRequestID) > 512 ||
		!utf8.ValidString(metadata.ObjectKey) ||
		!utf8.ValidString(metadata.ObjectVersion) || !utf8.ValidString(metadata.KeyID) ||
		!utf8.ValidString(metadata.ProviderRequestID) ||
		!exportV7(metadata.EffectID) || !exportV7(metadata.EffectReceiptID) ||
		metadata.EffectID == metadata.EffectReceiptID ||
		metadata.HeaderSHA256 == [32]byte{} || metadata.CiphertextSHA256 == [32]byte{} ||
		metadata.RequestSHA256 == [32]byte{} || metadata.ResultSHA256 == [32]byte{} ||
		metadata.CiphertextBytes < 1 || metadata.CiphertextBytes > 2148532224 ||
		len(metadata.WrappedDEK) < 32 || len(metadata.WrappedDEK) > 1024 ||
		metadata.WrapNonce == [12]byte{} ||
		metadata.AttemptGeneration < 1 || !exportArtifactTime(metadata.SealedAt) {
		return ErrExportArtifactTampered
	}
	return nil
}

type ExportArtifactCommit func(context.Context, ExportArtifactMetadata) error

type ExportArtifactBegin func(context.Context, ExportArtifactBeginEvidence) error

type ExportArtifactSealer interface {
	// SealExport must finish local FCART construction before begin. It must not
	// perform S3 or other network I/O unless begin returns nil, and it must call
	// commit only after the provider upload has a complete immutable receipt.
	SealExport(
		context.Context,
		ExportArtifactSealRequest,
		ExportArtifactBegin,
		ExportArtifactCommit,
	) error
}

func exportArtifactTime(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC &&
		value.Nanosecond()%int(time.Microsecond) == 0
}
