package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
	"github.com/google/uuid"
)

var (
	retentionJobID     = uuid.MustParse("019c1000-0000-7000-8000-000000000001")
	retentionAttemptID = uuid.MustParse("019c1000-0000-7000-8000-000000000002")
	retentionNextID    = uuid.MustParse("019c1000-0000-7000-8000-000000000003")
)

type retentionStoreStub struct {
	deleted  int
	err      error
	requests []RetentionPrune
}

func (store *retentionStoreStub) Prune(
	_ context.Context,
	request RetentionPrune,
) (int, error) {
	store.requests = append(store.requests, request)
	return store.deleted, store.err
}

func TestRetentionExecutePreservesAuthorityAndBoundedBacklog(t *testing.T) {
	cycleAt := time.Date(2026, 8, 2, 2, 0, 0, 0, time.UTC)
	startedAt := cycleAt.Add(5 * time.Minute)
	claim := retentionClaim(t, cycleAt, startedAt)
	store := &retentionStoreStub{deleted: RetentionBatchLimit}
	commands, err := NewRetentionCommands(store, fixedIDs{retentionNextID})
	if err != nil {
		t.Fatal(err)
	}
	outcome, err := commands.Execute(context.Background(), claim)
	if err != nil || !outcome.HasBacklog || outcome.DeletedCount != RetentionBatchLimit ||
		len(store.requests) != 1 {
		t.Fatalf("retention outcome=%+v requests=%d err=%v", outcome, len(store.requests), err)
	}
	request := store.requests[0]
	if request.ReceiptID != retentionJobID || request.EnvironmentID != testEnvironment ||
		request.DeleteBefore != cycleAt.Add(-RetentionAge) ||
		request.BatchLimit != RetentionBatchLimit || request.ObservedAt != cycleAt ||
		request.Authority.GroupGeneration() != 7 ||
		request.Authority.DomainGeneration() != 11 {
		t.Fatalf("retention owner request drifted: %+v", request)
	}
	if outcome.Successor.RunAt != startedAt.Add(time.Second) ||
		outcome.Successor.JobID != retentionNextID {
		t.Fatalf("backlog successor drifted: %+v", outcome.Successor)
	}
	nextEnvironment, nextCycle, err := ParseRetentionReference(
		outcome.Successor.PayloadReference,
	)
	if err != nil || nextEnvironment != testEnvironment || nextCycle != cycleAt {
		t.Fatalf("backlog reference=%s env=%s cycle=%s err=%v",
			outcome.Successor.PayloadReference, nextEnvironment, nextCycle, err)
	}
}

func TestRetentionExecuteReplaysStableReceiptAndSkipsMissedIdleCycles(t *testing.T) {
	cycleAt := time.Date(2026, 7, 28, 2, 0, 0, 0, time.UTC)
	startedAt := time.Date(2026, 8, 2, 3, 0, 0, 0, time.UTC)
	claim := retentionClaim(t, cycleAt, startedAt)
	store := &retentionStoreStub{deleted: 0}
	commands, _ := NewRetentionCommands(store, fixedIDs{retentionNextID})
	first, firstErr := commands.Execute(context.Background(), claim)
	second, secondErr := commands.Execute(context.Background(), claim)
	if firstErr != nil || secondErr != nil || first != second || len(store.requests) != 2 ||
		store.requests[0].ReceiptID != store.requests[1].ReceiptID {
		t.Fatalf("retention replay first=%+v second=%+v err=%v/%v requests=%+v",
			first, second, firstErr, secondErr, store.requests)
	}
	_, nextCycle, err := ParseRetentionReference(first.Successor.PayloadReference)
	if err != nil || !nextCycle.After(startedAt) || first.Successor.RunAt != nextCycle {
		t.Fatalf("idle successor cycle=%s started=%s err=%v", nextCycle, startedAt, err)
	}
}

func TestRetentionExecuteRejectsTamperedAuthorityAndClaim(t *testing.T) {
	cycleAt := time.Date(2026, 8, 2, 2, 0, 0, 0, time.UTC)
	claim := retentionClaim(t, cycleAt, cycleAt)
	store := &retentionStoreStub{}
	commands, _ := NewRetentionCommands(store, fixedIDs{retentionNextID})
	tampered := []byte(claim.PayloadSHA256)
	if tampered[0] == '0' {
		tampered[0] = '1'
	} else {
		tampered[0] = '0'
	}
	claim.PayloadSHA256 = string(tampered)
	if _, err := commands.Execute(context.Background(), claim); !errors.Is(err, domain.ErrInvalid) {
		t.Fatalf("tampered digest error=%v", err)
	}
	if len(store.requests) != 0 {
		t.Fatal("tampered claim reached retention owner store")
	}
	_, err := NewRetentionAuthority(RetentionAuthorityInput{
		EnvironmentID: testEnvironment, ScopeKind: RetentionScopeKind,
		ScopeID: testEnvironment, GroupID: RetentionAuthorityGroup,
		GroupGeneration: 0, Domain: RetentionAuthorityDomain,
		DomainGeneration: 1, WriterOrigin: RetentionWriterOrigin,
	})
	if !errors.Is(err, domain.ErrForbidden) {
		t.Fatalf("zero generation authority error=%v", err)
	}
}

func retentionClaim(t *testing.T, cycleAt, startedAt time.Time) RetentionJobClaim {
	t.Helper()
	authority, err := NewRetentionAuthority(RetentionAuthorityInput{
		EnvironmentID: testEnvironment, ScopeKind: RetentionScopeKind,
		ScopeID: testEnvironment, GroupID: RetentionAuthorityGroup,
		GroupGeneration: 7, Domain: RetentionAuthorityDomain,
		DomainGeneration: 11, WriterOrigin: RetentionWriterOrigin,
	})
	if err != nil {
		t.Fatal(err)
	}
	reference, digest, err := RetentionReference(testEnvironment, cycleAt)
	if err != nil {
		t.Fatal(err)
	}
	return RetentionJobClaim{JobID: retentionJobID, AttemptID: retentionAttemptID,
		Authority: authority, Kind: RetentionJobKind,
		PayloadRevision: RetentionPayloadRevision, PayloadReference: reference,
		PayloadSHA256: digest, LeaseGeneration: 1, StartedAt: startedAt}
}
