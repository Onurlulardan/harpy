package dashboardexportartifact

import (
	"bytes"
	"context"
	"crypto/sha256"
	"io"
	"testing"
	"time"

	fcartv1 "github.com/Onurlulardan/harpy/contracts/fcart/v1"
	artifactobject "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactobject/api"
	exportapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboardexport"
	"github.com/google/uuid"
)

func TestPlaintextEvidenceWriterBoundsAndHashesStream(t *testing.T) {
	t.Parallel()
	var destination bytes.Buffer
	writer := newPlaintextEvidenceWriter(&destination, 5)
	if written, err := writer.Write([]byte("hello")); err != nil || written != 5 {
		t.Fatalf("Write() = %d, %v", written, err)
	}
	if _, err := writer.Write([]byte("!")); err != errPlaintextLimit {
		t.Fatalf("over-limit Write() error = %v", err)
	}
	length, digest := writer.evidence()
	want := sha256.Sum256([]byte("hello"))
	if length != 5 || digest != want || destination.String() != "hello" {
		t.Fatalf("evidence = %d/%x/%q", length, digest, destination.String())
	}
}

func TestOpenStreamsAndVerifiesFCARTPlaintext(t *testing.T) {
	t.Parallel()
	plaintext := bytes.Repeat([]byte("analytics-stream\n"), 1024)
	runtime, stored := streamingArtifactFixture(t, plaintext)
	reader, err := runtime.Open(context.Background(), stored)
	if err != nil {
		t.Fatal(err)
	}
	opened, err := io.ReadAll(reader)
	closeErr := reader.Close()
	if err != nil || closeErr != nil || !bytes.Equal(opened, plaintext) {
		t.Fatalf("streamed FCART = %d bytes, read=%v close=%v",
			len(opened), err, closeErr)
	}
}

func TestOpenReportsPlaintextEvidenceMismatchAtStreamEnd(t *testing.T) {
	t.Parallel()
	runtime, stored := streamingArtifactFixture(t, []byte("analytics-integrity"))
	stored.Draft.PlaintextSHA256[0] ^= 0xff
	reader, err := runtime.Open(context.Background(), stored)
	if err != nil {
		t.Fatal(err)
	}
	_, err = io.ReadAll(reader)
	_ = reader.Close()
	if err == nil {
		t.Fatal("streamed FCART accepted mismatched plaintext evidence")
	}
}

func streamingArtifactFixture(t *testing.T, plaintext []byte) (
	*Runtime, exportapp.StoredContent,
) {
	t.Helper()
	tenantID := uuid.MustParse("019b1000-0000-7000-8000-000000000001")
	exportID := uuid.MustParse("019b1000-0000-7000-8000-000000000061")
	kek, err := fcartv1.NewKEK("analytics-kek-1", bytes.Repeat([]byte{0x41}, 32))
	if err != nil {
		t.Fatal(err)
	}
	keys, err := fcartv1.NewEnvelopeKeyring(kek, nil,
		bytes.NewReader(bytes.Repeat([]byte{0x42}, 128)))
	if err != nil {
		t.Fatal(err)
	}
	identity := fcartv1.KeyIdentity{TenantID: tenantID,
		ArtifactID: exportID, AttemptGeneration: 1}
	key, envelope, err := keys.Generate(context.Background(), identity)
	if err != nil {
		t.Fatal(err)
	}
	header, err := fcartv1.NewHeader(exportID, 1, fcartv1.FormatCSV)
	if err != nil {
		t.Fatal(err)
	}
	var ciphertext bytes.Buffer
	encrypted, err := fcartv1.Encrypt(&ciphertext, bytes.NewReader(plaintext), header,
		key, fcartv1.NewCryptographicRandomSource())
	clear(key[:])
	if err != nil {
		t.Fatal(err)
	}
	plainDigest := sha256.Sum256(plaintext)
	draft := exportapp.ArtifactDraft{AttemptGeneration: 1,
		ObjectKey: "analytics/v1/export/1.fcart", Envelope: envelope,
		HeaderSHA256: encrypted.HeaderSHA256, CiphertextSHA256: encrypted.CiphertextSHA256,
		PlaintextSHA256: plainDigest, CiphertextBytes: int64(encrypted.CiphertextBytes),
		PlaintextBytes: int64(len(plaintext)), Rows: 1,
		ContentType: exportapp.FormatCSV.ContentType(), Filename: "dashboard.csv"}
	objects := &streamingObjectStore{body: ciphertext.Bytes()}
	return &Runtime{objects: objects, keys: keys, prefix: "analytics"},
		exportapp.StoredContent{Pin: exportapp.Pin{TenantID: tenantID,
			ExportID: exportID, Format: exportapp.FormatCSV}, Draft: draft,
			ObjectVersion: "version-1", ReadyAt: time.Now().UTC()}
}

type streamingObjectStore struct{ body []byte }

func (store *streamingObjectStore) OpenExact(
	context.Context, artifactobject.Expected,
) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(store.body)), nil
}
func (*streamingObjectStore) Upload(
	context.Context, artifactobject.Upload,
) (artifactobject.UploadResult, error) {
	return artifactobject.UploadResult{}, artifactobject.ErrUnavailable
}
func (*streamingObjectStore) DeleteExact(
	context.Context, artifactobject.Identity,
) (artifactobject.DeleteResult, error) {
	return artifactobject.DeleteResult{}, artifactobject.ErrUnavailable
}
func (*streamingObjectStore) ReconcileExact(
	context.Context, artifactobject.ReconcileExpected,
) (artifactobject.ReconcileResult, error) {
	return artifactobject.ReconcileResult{}, artifactobject.ErrUnavailable
}
