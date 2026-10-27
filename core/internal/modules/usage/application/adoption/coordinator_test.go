package adoption

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
)

func executeRequest() ExecuteRequest {
	return ExecuteRequest{OperationID: operation, RunID: runID, LeaseOwner: leaseOwner,
		PageLimit: 100, PageBudget: 10}
}

func TestZeroSourceAdoptionCompletesWithoutReadingPages(t *testing.T) {
	source := &memorySource{manifest: manifest(ZeroSource, 0)}
	target := &memoryTarget{}
	coordinator, err := NewCoordinator(source, target, fixedClock{adoptNow})
	if err != nil {
		t.Fatal(err)
	}
	run, err := coordinator.Execute(context.Background(), executeRequest())
	if err != nil || run.Status != StatusCompleted || source.pages != 0 {
		t.Fatalf("zero-source adoption failed: run=%+v pages=%d err=%v", run, source.pages, err)
	}
}

func TestAdoptionFencesOIDCountOrRootDriftBeforeNextPage(t *testing.T) {
	original := manifest(LegacyTrace, 1)
	drifted := original
	drifted.RelationOID++
	source := &memorySource{manifest: original, drift: &drifted, page: legacyPage()}
	target := &memoryTarget{}
	coordinator, _ := NewCoordinator(source, target, fixedClock{adoptNow})
	run, err := coordinator.Execute(context.Background(), executeRequest())
	if !errors.Is(err, domain.ErrSourceDrift) || target.run.Status != StatusFenced ||
		run.CopiedCount != 0 || source.pages != 0 {
		t.Fatalf("source drift was not fenced: run=%+v pages=%d err=%v", run, source.pages, err)
	}
}

func TestAdoptionResumesAfterApplyBeforeAdvanceCrash(t *testing.T) {
	source := &memorySource{manifest: manifest(LegacyTrace, 1), page: legacyPage()}
	target := &memoryTarget{failAdvance: true}
	coordinator, _ := NewCoordinator(source, target, fixedClock{adoptNow})
	first, err := coordinator.Execute(context.Background(), executeRequest())
	if !errors.Is(err, domain.ErrUnavailable) || first.CopiedCount != 0 || len(target.receipts) != 1 {
		t.Fatalf("crash point was not preserved: run=%+v receipts=%d err=%v",
			first, len(target.receipts), err)
	}
	completed, err := coordinator.Execute(context.Background(), executeRequest())
	if err != nil || completed.Status != StatusCompleted || completed.CopiedCount != 1 ||
		len(target.receipts) != 1 || source.pages != 2 {
		t.Fatalf("crash resume failed: run=%+v receipts=%d pages=%d err=%v",
			completed, len(target.receipts), source.pages, err)
	}
}
