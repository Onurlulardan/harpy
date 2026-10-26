package domain

import (
	"crypto/sha256"
	"errors"
	"testing"
)

func TestAdoptionAcceptanceRequiresExactBytesMerkleAndNoPendingEffects(t *testing.T) {
	root := sha256.Sum256([]byte("root"))
	run := AdoptionRun{State: AdoptionQuiesced, SourceCount: 3, TargetCount: 3,
		SourceBytes: 99, TargetBytes: 99, SourceMerkleSHA256: root, TargetMerkleSHA256: root,
		FinalLSN: 100, AppliedLSN: 100, AcceptanceSHA256: sha256.Sum256([]byte("accept"))}
	accepted, err := run.Advance(AdoptionAccepted)
	if err != nil || accepted.State != AdoptionAccepted {
		t.Fatalf("accepted = %+v, %v", accepted, err)
	}
	run.PendingObjectCount = 1
	if _, err = run.Advance(AdoptionAccepted); !errors.Is(err, ErrAdoptionTransition) {
		t.Fatalf("pending object error = %v", err)
	}
}

func TestAdoptionCannotSkipShadowAndReconcile(t *testing.T) {
	run := AdoptionRun{State: AdoptionInventory}
	if _, err := run.Advance(AdoptionAccepted); !errors.Is(err, ErrAdoptionTransition) {
		t.Fatalf("skipped lifecycle error = %v", err)
	}
}
