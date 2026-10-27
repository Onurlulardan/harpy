package attestation

import (
	"bytes"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestWriteNewCanonicalPublishesCompleteImmutableTarget(t *testing.T) {
	path := filepath.Join(t.TempDir(), "receipt.json")
	if err := WriteNewCanonical(path, []byte(`{"b":2,"a":1}`)); err != nil {
		t.Fatal(err)
	}
	written, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(written) != `{"a":1,"b":2}` {
		t.Fatalf("published bytes = %s", written)
	}
	if err := WriteNewCanonical(path, []byte(`{"a":9}`)); err == nil {
		t.Fatal("existing receipt was replaceable")
	}
	unchanged, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(unchanged) != string(written) {
		t.Fatal("failed replacement changed receipt")
	}
}

func TestWriteNewCanonicalHasOneConcurrentWinner(t *testing.T) {
	path := filepath.Join(t.TempDir(), "receipt.json")
	documents := [][]byte{[]byte(`{"winner":1}`), []byte(`{"winner":2}`)}
	errors := make(chan error, len(documents))
	start := make(chan struct{})
	var wait sync.WaitGroup
	for _, document := range documents {
		wait.Add(1)
		go func(value []byte) {
			defer wait.Done()
			<-start
			errors <- WriteNewCanonical(path, value)
		}(document)
	}
	close(start)
	wait.Wait()
	close(errors)
	successes := 0
	for err := range errors {
		if err == nil {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("concurrent successes = %d", successes)
	}
	written, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(written) != `{"winner":1}` && string(written) != `{"winner":2}` {
		t.Fatalf("partial concurrent publication: %s", written)
	}
}

func TestWriteNewCanonicalRejectsRelativeTarget(t *testing.T) {
	if err := WriteNewCanonical("receipt.json", []byte(`{"safe":true}`)); err == nil {
		t.Fatal("relative receipt path accepted")
	}
}

func TestWriteNewCanonicalBoundedUsesArtifactLimit(t *testing.T) {
	document := append([]byte(`{"value":"`),
		bytes.Repeat([]byte{'a'}, maximumDocumentBytes)...)
	document = append(document, []byte(`"}`)...)
	path := filepath.Join(t.TempDir(), "large-manifest.json")
	if err := WriteNewCanonical(path, document); err == nil {
		t.Fatal("generic attestation writer accepted oversized document")
	}
	if err := WriteNewCanonicalBounded(path, document, len(document)); err != nil {
		t.Fatalf("artifact-bounded publication failed: %v", err)
	}
	written, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(written, document) {
		t.Fatal("artifact-bounded publication changed canonical bytes")
	}
}
