package main

import (
	"os"
	"path/filepath"
	"testing"
)

type failingStagingFile struct {
	chmodErr error
	writeErr error
	short    bool
	closed   bool
}

func (file *failingStagingFile) Chmod(os.FileMode) error { return file.chmodErr }
func (file *failingStagingFile) Write(value []byte) (int, error) {
	if file.short {
		return len(value) - 1, file.writeErr
	}
	return len(value), file.writeErr
}
func (file *failingStagingFile) Sync() error  { return nil }
func (file *failingStagingFile) Close() error { file.closed = true; return nil }

func TestPublishNoReplaceIsIdempotentAndRejectsDifferentBytes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "receipt.json")
	document := []byte(`{"contract_version":"receipt.v1"}`)
	if err := publishNoReplace(path, document); err != nil {
		t.Fatalf("initial publication: %v", err)
	}
	if err := publishNoReplace(path, append([]byte(nil), document...)); err != nil {
		t.Fatalf("identical replay: %v", err)
	}
	if err := publishNoReplace(path, []byte(`{"different":true}`)); err == nil {
		t.Fatal("different replay replaced existing artifact")
	}
	stored, err := os.ReadFile(path)
	if err != nil || string(stored) != string(document) {
		t.Fatalf("stored receipt changed: %q err=%v", stored, err)
	}
}

func TestStageArtifactNeverAcceptsChmodOrShortWriteFailure(t *testing.T) {
	chmodFailure := &failingStagingFile{chmodErr: os.ErrPermission}
	if err := stageArtifact(chmodFailure, []byte("document")); err == nil ||
		!chmodFailure.closed {
		t.Fatal("chmod failure did not fail closed")
	}
	shortWrite := &failingStagingFile{short: true}
	if err := stageArtifact(shortWrite, []byte("document")); err == nil ||
		!shortWrite.closed {
		t.Fatal("short write did not fail closed")
	}
}

func TestPublishAdoptionReceiptRecoversMissingSignature(t *testing.T) {
	directory := t.TempDir()
	profile := adoptionInvocation{
		ReceiptPath:          filepath.Join(directory, "receipt.json"),
		ReceiptSignaturePath: filepath.Join(directory, "receipt.sig"),
	}
	signed := signedAdoptionReceipt{
		document:          []byte(`{"receipt":true}`),
		signatureDocument: []byte(`{"signature":true}`),
	}
	if err := publishNoReplace(profile.ReceiptPath, signed.document); err != nil {
		t.Fatal(err)
	}
	if err := publishAdoptionReceipt(profile, signed); err != nil {
		t.Fatalf("recover signature publication: %v", err)
	}
	if _, err := os.Stat(profile.ReceiptSignaturePath); err != nil {
		t.Fatalf("signature was not published: %v", err)
	}
}
