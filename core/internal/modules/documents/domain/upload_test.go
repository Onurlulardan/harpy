package domain

import (
	"crypto/sha256"
	"errors"
	"testing"
)

func TestUploadRetryRequiresCleanupAndIncrementsGeneration(t *testing.T) {
	upload := validUpload(t)
	first, err := upload.BeginAttempt()
	if err != nil || first.AttemptGeneration != 1 || first.State != UploadReceiving {
		t.Fatalf("first attempt = %+v, %v", first, err)
	}
	retryable, err := first.RetryAfterTransient()
	if err != nil || !retryable.CleanupRequired {
		t.Fatalf("retryable = %+v, %v", retryable, err)
	}
	if _, err = retryable.BeginAttempt(); !errors.Is(err, ErrUploadTransition) {
		t.Fatalf("attempt without cleanup = %v", err)
	}
	receipt := sha256.Sum256([]byte("object-generation-cleanup"))
	retryable, err = retryable.ConfirmCleanup(receipt)
	if err != nil {
		t.Fatal(err)
	}
	second, err := retryable.BeginAttempt()
	if err != nil || second.AttemptGeneration != 2 {
		t.Fatalf("second attempt = %+v, %v", second, err)
	}
}

func TestUploadReadyCASAndReplayAreDeterministic(t *testing.T) {
	upload := validUpload(t)
	receiving, _ := upload.BeginAttempt()
	scanning, err := receiving.CompleteStream(
		upload.ExpectedBytes, upload.ExpectedContentSHA256,
	)
	if err != nil || scanning.State != UploadScanning {
		t.Fatalf("scanning = %+v, %v", scanning, err)
	}
	wrong := mustID(t, "018f0000-0000-7000-8000-000000000099")
	conflicted, err := scanning.AcceptReady(wrong)
	if !errors.Is(err, ErrUploadConflict) || !conflicted.CleanupRequired {
		t.Fatalf("conflict = %+v, %v", conflicted, err)
	}
	ready, err := scanning.AcceptReady(upload.ExpectedCurrentVersionID)
	if err != nil || ready.State != UploadReady {
		t.Fatalf("ready = %+v, %v", ready, err)
	}
	replayed, err := ready.AcceptReady(upload.ExpectedCurrentVersionID)
	if err != nil || replayed != ready {
		t.Fatalf("ready replay = %+v, %v", replayed, err)
	}
}

func validUpload(t *testing.T) Upload {
	t.Helper()
	request := sha256.Sum256([]byte("upload-request"))
	content := sha256.Sum256([]byte("content"))
	upload, err := NewUpload(UploadInput{
		ID:                       mustID(t, "018f0000-0000-7000-8000-000000000001"),
		TenantID:                 mustID(t, "11111111-1111-4111-8111-111111111111"),
		ActorID:                  mustID(t, "22222222-2222-4222-8222-222222222222"),
		DocumentID:               mustID(t, "018f0000-0000-7000-8000-000000000002"),
		CandidateVersionID:       mustID(t, "018f0000-0000-7000-8000-000000000003"),
		ExpectedCurrentVersionID: mustID(t, "018f0000-0000-7000-8000-000000000004"),
		RequestSHA256:            request, ExpectedContentSHA256: content, ExpectedBytes: 7,
	})
	if err != nil {
		t.Fatal(err)
	}
	return upload
}

func mustID(t *testing.T, value string) ID {
	t.Helper()
	id, err := ParseID(value)
	if err != nil {
		t.Fatal(err)
	}
	return id
}
