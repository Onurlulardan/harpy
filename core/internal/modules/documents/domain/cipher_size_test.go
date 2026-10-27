package domain

import "testing"

func TestCiphertextUpperBoundAccountsForEveryFCDOC1Chunk(t *testing.T) {
	one, err := CiphertextUpperBound(1)
	if err != nil || one != 111 {
		t.Fatalf("one-byte bound = %d, %v", one, err)
	}
	twoChunks, err := CiphertextUpperBound(cipherChunkBytes + 1)
	if err != nil || twoChunks != cipherHeaderBytes+cipherChunkBytes+1+2*cipherFrameBytes {
		t.Fatalf("two-chunk bound = %d, %v", twoChunks, err)
	}
}

func TestArchiveCiphertextUpperBoundSupportsTenGiBArtifact(t *testing.T) {
	got, err := ArchiveCiphertextUpperBound(MaximumArchivePlaintext)
	chunks := (MaximumArchivePlaintext + cipherChunkBytes - 1) / cipherChunkBytes
	want := cipherHeaderBytes + MaximumArchivePlaintext + chunks*cipherFrameBytes
	if err != nil || got != want {
		t.Fatalf("archive ciphertext bytes = %d, %v; want %d", got, err, want)
	}
}
