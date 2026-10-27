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
)

type trackedPlaintext struct {
	reader *bytes.Reader
	eof    bool
}

func (value *trackedPlaintext) Read(target []byte) (int, error) {
	read, err := value.reader.Read(target)
	if errors.Is(err, io.EOF) {
		value.eof = true
	}
	return read, err
}

func TestExportSealFencesEffectBeforeUploadAndAuthenticatesDownload(t *testing.T) {
	runtime, _, exports := newRuntimeFixture(t)
	tenantID, exportID := fixtureV7(t), fixtureV7(t)
	effectID, receiptID := fixtureV7(t), fixtureV7(t)
	plaintext := []byte("id,email\n1,one@example.test\n")
	tracked := &trackedPlaintext{reader: bytes.NewReader(plaintext)}
	began := false
	exports.beforeRead = func() {
		if !began {
			t.Fatal("provider upload started before durable effect fence")
		}
	}
	var metadata systemtransfer.ExportArtifactMetadata
	err := runtime.SealExport(context.Background(), systemtransfer.ExportArtifactSealRequest{
		TenantID: tenantID, ExportID: exportID, EffectID: effectID,
		EffectReceiptID: receiptID, AttemptGeneration: 1, Format: "csv",
		Plaintext: tracked, MaximumPlaintextBytes: int64(len(plaintext)),
	}, func(_ context.Context, evidence systemtransfer.ExportArtifactBeginEvidence) error {
		if !tracked.eof || evidence.EffectID != effectID || evidence.RequestSHA256 == [32]byte{} {
			t.Fatalf("invalid begin evidence: %#v eof=%v", evidence, tracked.eof)
		}
		began = true
		return nil
	}, func(_ context.Context, value systemtransfer.ExportArtifactMetadata) error {
		metadata = value
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !began || exports.uploads != 1 || metadata.EffectID != effectID ||
		metadata.EffectReceiptID != receiptID || metadata.RequestSHA256 == [32]byte{} ||
		metadata.ResultSHA256 == [32]byte{} {
		t.Fatalf("effect evidence mismatch: %#v", metadata)
	}
	ready, expires := fixtureArtifactTime, fixtureArtifactTime.Add(time.Hour)
	reader, err := runtime.OpenExport(context.Background(),
		systemtransfer.ExportOpenRequest{TenantID: tenantID,
			Resource: publicresource.Users, ExportID: exportID,
			Artifact: systemtransfer.ExportArtifact{
				ExportRecord: systemtransfer.ExportRecord{ExportID: exportID,
					JobID: fixtureV7(t), Resource: "users", Format: "csv", State: "ready",
					Bytes: int64(len(plaintext)), CreatedAt: fixtureArtifactTime,
					ReadyAt: &ready, ExpiresAt: &expires},
				ObjectKey: metadata.ObjectKey, ObjectVersion: metadata.ObjectVersion,
				KeyID: metadata.KeyID, HeaderSHA256: metadata.HeaderSHA256,
				CiphertextSHA256: metadata.CiphertextSHA256,
				CiphertextBytes:  metadata.CiphertextBytes,
				WrappedDEK:       metadata.WrappedDEK, WrapNonce: metadata.WrapNonce,
				AttemptGeneration: metadata.AttemptGeneration,
			}})
	if err != nil {
		t.Fatal(err)
	}
	decoded, readErr := io.ReadAll(reader)
	closeErr := reader.Close()
	if readErr != nil || closeErr != nil || !bytes.Equal(decoded, plaintext) {
		t.Fatalf("download mismatch: %q read=%v close=%v", decoded, readErr, closeErr)
	}
}

func TestExportSealDoesNotUploadWhenEffectFenceFails(t *testing.T) {
	runtime, _, exports := newRuntimeFixture(t)
	fenceErr := errors.New("effect fence failed")
	err := runtime.SealExport(context.Background(), systemtransfer.ExportArtifactSealRequest{
		TenantID: fixtureV7(t), ExportID: fixtureV7(t), EffectID: fixtureV7(t),
		EffectReceiptID: fixtureV7(t), AttemptGeneration: 1, Format: "jsonl",
		Plaintext: bytes.NewBufferString("{\"id\":1}\n"), MaximumPlaintextBytes: 64,
	}, func(context.Context, systemtransfer.ExportArtifactBeginEvidence) error {
		return fenceErr
	}, func(context.Context, systemtransfer.ExportArtifactMetadata) error {
		t.Fatal("commit called after failed effect fence")
		return nil
	})
	if !errors.Is(err, fenceErr) || exports.uploads != 0 {
		t.Fatalf("seal error=%v uploads=%d", err, exports.uploads)
	}
}

func TestExportCommitFailureRetainsUploadedObjectForReconciliation(t *testing.T) {
	runtime, _, exports := newRuntimeFixture(t)
	commitErr := errors.New("receipt commit unavailable")
	err := runtime.SealExport(context.Background(), systemtransfer.ExportArtifactSealRequest{
		TenantID: fixtureV7(t), ExportID: fixtureV7(t), EffectID: fixtureV7(t),
		EffectReceiptID: fixtureV7(t), AttemptGeneration: 1, Format: "jsonl",
		Plaintext: bytes.NewBufferString("{\"id\":1}\n"), MaximumPlaintextBytes: 64,
	}, func(context.Context, systemtransfer.ExportArtifactBeginEvidence) error { return nil },
		func(context.Context, systemtransfer.ExportArtifactMetadata) error { return commitErr })
	if !errors.Is(err, commitErr) || exports.uploads != 1 || exports.deletes != 0 ||
		len(exports.objects) != 1 {
		t.Fatalf("seal error=%v uploads=%d deletes=%d objects=%d",
			err, exports.uploads, exports.deletes, len(exports.objects))
	}
}
