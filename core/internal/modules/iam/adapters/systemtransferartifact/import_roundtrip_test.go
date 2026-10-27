package systemtransferartifact

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	"github.com/google/uuid"
)

func TestImportSourceSealsAnalyzesAndAuthenticatesExactPlaintext(t *testing.T) {
	runtime, _, _ := newRuntimeFixture(t)
	tenantID, actorID, sourceID := fixtureV7(t), fixtureV7(t), fixtureV7(t)
	plaintext := []byte("email,active\none@example.test,true\n")
	var artifact systemtransfer.SourceArtifact
	record, err := runtime.SealImportSource(context.Background(), importRequest(
		tenantID, actorID, sourceID, plaintext), func(_ context.Context,
		value systemtransfer.SourceArtifact) (systemtransfer.SourceRecord, error) {
		artifact = value
		return sourceRecord(sourceID, value, false), nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(record.Analysis.Columns) != 2 || len(record.Analysis.Samples) != 1 ||
		record.Analysis.Samples[0]["email"] != "one@example.test" {
		t.Fatalf("unexpected analysis: %#v", record.Analysis)
	}
	reader, err := runtime.OpenExactImportSource(context.Background(),
		systemtransfer.ExactImportSource{TenantID: tenantID, ActorID: actorID,
			SourceID: sourceID, Resource: publicresource.Users, Format: "csv",
			Artifact: artifact})
	if err != nil {
		t.Fatal(err)
	}
	decoded, readErr := io.ReadAll(reader)
	closeErr := reader.Close()
	if readErr != nil || closeErr != nil || !bytes.Equal(decoded, plaintext) {
		t.Fatalf("plaintext mismatch: %q, read=%v close=%v", decoded, readErr, closeErr)
	}
}

func TestImportSourceReplayDeletesOnlyRedundantExactVersion(t *testing.T) {
	runtime, imports, _ := newRuntimeFixture(t)
	tenantID, actorID, sourceID := fixtureV7(t), fixtureV7(t), fixtureV7(t)
	existingID := fixtureV7(t)
	plaintext := []byte("email\none@example.test\n")
	record, err := runtime.SealImportSource(context.Background(), importRequest(
		tenantID, actorID, sourceID, plaintext), func(_ context.Context,
		artifact systemtransfer.SourceArtifact) (systemtransfer.SourceRecord, error) {
		return sourceRecord(existingID, artifact, true), nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if record.SourceID != existingID || !record.Replay || imports.deletes != 1 ||
		len(imports.objects) != 0 {
		t.Fatalf("replay cleanup mismatch: record=%#v deletes=%d objects=%d",
			record, imports.deletes, len(imports.objects))
	}
}

func TestImportSourceRegistrationFailureCleanupIsOutcomeAware(t *testing.T) {
	t.Run("definite rejection deletes", func(t *testing.T) {
		runtime, imports, _ := newRuntimeFixture(t)
		_, err := runtime.SealImportSource(context.Background(), importRequest(
			fixtureV7(t), fixtureV7(t), fixtureV7(t), []byte("email\na@b.test\n")),
			func(context.Context, systemtransfer.SourceArtifact) (
				systemtransfer.SourceRecord, error) {
				return systemtransfer.SourceRecord{}, systemtransfer.ErrControlConflict
			})
		if !errors.Is(err, systemtransfer.ErrControlConflict) || imports.deletes != 1 {
			t.Fatalf("definite failure = %v, deletes=%d", err, imports.deletes)
		}
	})
	t.Run("unknown commit retains", func(t *testing.T) {
		runtime, imports, _ := newRuntimeFixture(t)
		_, err := runtime.SealImportSource(context.Background(), importRequest(
			fixtureV7(t), fixtureV7(t), fixtureV7(t), []byte("email\na@b.test\n")),
			func(context.Context, systemtransfer.SourceArtifact) (
				systemtransfer.SourceRecord, error) {
				return systemtransfer.SourceRecord{}, systemtransfer.ErrControlUnavailable
			})
		if !errors.Is(err, errOutcomeUnknown) || imports.deletes != 0 ||
			len(imports.objects) != 1 {
			t.Fatalf("unknown failure = %v, deletes=%d objects=%d",
				err, imports.deletes, len(imports.objects))
		}
	})
}

func importRequest(tenantID, actorID, sourceID uuid.UUID,
	body []byte) systemtransfer.SourceSealRequest {
	return systemtransfer.SourceSealRequest{TenantID: tenantID, ActorID: actorID,
		SourceID: sourceID, Resource: publicresource.Users, Format: "csv",
		IdempotencySHA256: [32]byte{1}, UploadedAt: fixtureArtifactTime,
		Body: bytes.NewReader(body), MaximumBytes: int64(len(body))}
}

func sourceRecord(id uuid.UUID, artifact systemtransfer.SourceArtifact,
	replay bool) systemtransfer.SourceRecord {
	return systemtransfer.SourceRecord{SourceID: id, Resource: "users", Format: "csv",
		PlaintextBytes:  artifact.PlaintextBytes,
		PlaintextSHA256: artifact.PlaintextSHA256,
		UploadedAt:      artifact.UploadedAt, Replay: replay}
}

func fixtureV7(t *testing.T) uuid.UUID {
	t.Helper()
	value, err := uuid.NewV7()
	if err != nil {
		t.Fatal(err)
	}
	return value
}
