//go:build integration

package systemtransferartifact

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	"github.com/google/uuid"
)

func sealMinIOImport(t *testing.T, runtime *Runtime, tenantID, actorID,
	sourceID uuid.UUID, plaintext []byte,
) systemtransfer.SourceArtifact {
	t.Helper()
	request := systemtransfer.SourceSealRequest{TenantID: tenantID, ActorID: actorID,
		SourceID: sourceID, Resource: publicresource.Users, Format: "csv",
		IdempotencySHA256: [32]byte{1}, UploadedAt: minIOArtifactClock{}.Now(),
		Body: bytes.NewReader(plaintext), MaximumBytes: int64(len(plaintext))}
	var artifact systemtransfer.SourceArtifact
	_, err := runtime.SealImportSource(t.Context(), request,
		func(_ context.Context, value systemtransfer.SourceArtifact) (
			systemtransfer.SourceRecord, error,
		) {
			artifact = value
			return systemtransfer.SourceRecord{SourceID: sourceID, Resource: "users",
				Format: "csv", PlaintextBytes: value.PlaintextBytes,
				PlaintextSHA256: value.PlaintextSHA256,
				UploadedAt:      value.UploadedAt}, nil
		})
	if err != nil {
		t.Fatal(err)
	}
	return artifact
}

func readMinIOImport(t *testing.T, runtime *Runtime, tenantID, actorID,
	sourceID uuid.UUID, artifact systemtransfer.SourceArtifact,
) ([]byte, error) {
	t.Helper()
	reader, err := runtime.OpenExactImportSource(t.Context(),
		systemtransfer.ExactImportSource{TenantID: tenantID, ActorID: actorID,
			SourceID: sourceID, Resource: publicresource.Users, Format: "csv",
			Artifact: artifact})
	if err != nil {
		return nil, err
	}
	value, readErr := io.ReadAll(reader)
	return value, errors.Join(readErr, reader.Close())
}

func sealMinIOExport(t *testing.T, runtime *Runtime, tenantID, exportID uuid.UUID,
	plaintext []byte,
) systemtransfer.ExportArtifactMetadata {
	t.Helper()
	var metadata systemtransfer.ExportArtifactMetadata
	err := runtime.SealExport(t.Context(), systemtransfer.ExportArtifactSealRequest{
		TenantID: tenantID, ExportID: exportID, EffectID: fixtureV7(t),
		EffectReceiptID: fixtureV7(t), AttemptGeneration: 1, Format: "jsonl",
		Plaintext: bytes.NewReader(plaintext), MaximumPlaintextBytes: int64(len(plaintext)),
	}, func(_ context.Context, evidence systemtransfer.ExportArtifactBeginEvidence) error {
		return evidence.Validate()
	}, func(_ context.Context, value systemtransfer.ExportArtifactMetadata) error {
		metadata = value
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return metadata
}

func readMinIOExport(t *testing.T, runtime *Runtime, tenantID, exportID uuid.UUID,
	plaintextBytes int64, metadata systemtransfer.ExportArtifactMetadata,
) ([]byte, error) {
	t.Helper()
	ready, expires := minIOArtifactClock{}.Now(), minIOArtifactClock{}.Now().Add(time.Hour)
	reader, err := runtime.OpenExport(t.Context(), systemtransfer.ExportOpenRequest{
		TenantID: tenantID, Resource: publicresource.Users, ExportID: exportID,
		Artifact: systemtransfer.ExportArtifact{ExportRecord: systemtransfer.ExportRecord{
			ExportID: exportID, JobID: fixtureV7(t), Resource: "users", Format: "jsonl",
			State: "ready", Bytes: plaintextBytes, CreatedAt: ready,
			ReadyAt: &ready, ExpiresAt: &expires}, ObjectKey: metadata.ObjectKey,
			ObjectVersion: metadata.ObjectVersion, KeyID: metadata.KeyID,
			HeaderSHA256:     metadata.HeaderSHA256,
			CiphertextSHA256: metadata.CiphertextSHA256,
			CiphertextBytes:  metadata.CiphertextBytes,
			WrappedDEK:       metadata.WrappedDEK, WrapNonce: metadata.WrapNonce,
			AttemptGeneration: metadata.AttemptGeneration}})
	if err != nil {
		return nil, err
	}
	value, readErr := io.ReadAll(reader)
	return value, errors.Join(readErr, reader.Close())
}
