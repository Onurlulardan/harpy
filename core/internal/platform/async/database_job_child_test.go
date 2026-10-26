package async

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func validDatabaseEffectChild() DatabaseJobChildInput {
	return DatabaseJobChildInput{
		JobID:                uuid.MustParse("019c6a7e-7c00-7000-8000-000000000030"),
		Kind:                 "billing.provider.reconcile",
		PayloadRevision:      1,
		RunAt:                time.Date(2026, 7, 17, 1, 2, 4, 0, time.UTC),
		PayloadReference:     "billing-provider-reconciliation/receipt-1",
		PayloadSHA256:        strings.Repeat("d", 64),
		DedupeKey:            "billing:provider:reconcile:receipt-1",
		EffectIdempotencyKey: "billing:provider:reconcile:account-1",
		Priority:             100,
	}
}

func TestDatabaseJobChildIsBoundedAndSealed(t *testing.T) {
	child := validDatabaseEffectChild()
	outcome, err := CompleteDatabaseJobWithEffects(DatabaseJobEffects{
		Children: []DatabaseJobChildInput{child},
	})
	if err != nil || !outcome.Valid() {
		t.Fatalf("complete child effect = %+v, %v", outcome, err)
	}
	child.Kind = "billing.provider.command"
	children := outcome.Children()
	if len(children) != 1 || children[0].Kind != "billing.provider.reconcile" {
		t.Fatalf("sealed children = %#v", children)
	}
	children[0].Kind = "billing.webhook.process"
	if outcome.Children()[0].Kind != "billing.provider.reconcile" {
		t.Fatal("returned child slice mutated outcome")
	}
}

func TestDatabaseJobChildRejectsDuplicateIdentityAndDedupe(t *testing.T) {
	first := validDatabaseEffectChild()
	second := first
	second.JobID = uuid.MustParse("019c6a7e-7c00-7000-8000-000000000031")
	for name, mutate := range map[string]func(*DatabaseJobChildInput){
		"duplicate id": func(value *DatabaseJobChildInput) {
			value.DedupeKey += ":other"
			value.JobID = first.JobID
		},
		"duplicate dedupe": func(*DatabaseJobChildInput) {},
		"invalid effect key": func(value *DatabaseJobChildInput) {
			value.DedupeKey += ":other"
			value.EffectIdempotencyKey = " padded"
		},
	} {
		t.Run(name, func(t *testing.T) {
			candidate := second
			mutate(&candidate)
			if _, err := CompleteDatabaseJobWithEffects(DatabaseJobEffects{
				Children: []DatabaseJobChildInput{first, candidate},
			}); err == nil {
				t.Fatal("ambiguous child set was accepted")
			}
		})
	}
}

func TestDatabaseJobChildCannotConflictWithOtherEffects(t *testing.T) {
	child := validDatabaseEffectChild()
	cancellation := validDatabaseEffectCancellation()
	cancellation.JobID = child.JobID
	if _, err := CompleteDatabaseJobWithEffects(DatabaseJobEffects{
		Children:      []DatabaseJobChildInput{child},
		Cancellations: []DatabaseJobCancellationInput{cancellation},
	}); err == nil {
		t.Fatal("child cancellation ambiguity was accepted")
	}
	successor := DatabaseJobSuccessorInput{
		JobID: child.JobID, RunAt: child.RunAt,
		PayloadReference: child.PayloadReference,
		PayloadSHA256:    child.PayloadSHA256, DedupeKey: child.DedupeKey,
	}
	if _, err := CompleteDatabaseJobWithEffects(DatabaseJobEffects{
		Successor: &successor, Children: []DatabaseJobChildInput{child},
	}); err == nil {
		t.Fatal("child successor ambiguity was accepted")
	}
}
