package customerlifecycle

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

type fixtureWorkerStore struct {
	claim     ClaimedOperation
	completed string
	failed    string
}

func (store *fixtureWorkerStore) Claim(context.Context, uuid.UUID, time.Time, time.Duration) (ClaimedOperation, error) {
	return store.claim, nil
}
func (store *fixtureWorkerStore) CompleteOffboard(_ context.Context, claim ClaimedOperation, _ time.Time) (Operation, error) {
	store.completed = OperationOffboard
	return claim.Operation, nil
}
func (store *fixtureWorkerStore) CompletePurge(_ context.Context, claim ClaimedOperation, _ ExecutionReceipt, _ time.Time) (Operation, error) {
	store.completed = OperationPurge
	return claim.Operation, nil
}
func (store *fixtureWorkerStore) Fail(_ context.Context, _ ClaimedOperation, code string, _ time.Time) error {
	store.failed = code
	return nil
}

type fixtureExecutor struct {
	calls   int
	receipt ExecutionReceipt
	err     error
}

func (executor *fixtureExecutor) Purge(context.Context, ClaimedOperation) (ExecutionReceipt, error) {
	executor.calls++
	return executor.receipt, executor.err
}

func TestWorkerOffboardNeverCallsDestructionExecutor(t *testing.T) {
	store := &fixtureWorkerStore{claim: ClaimedOperation{Operation: Operation{Kind: OperationOffboard}}}
	executor := &fixtureExecutor{}
	worker, err := NewWorker(WorkerDependencies{Store: store, Executor: executor, Clock: fixtureClock{fixtureNow}})
	if err != nil {
		t.Fatal(err)
	}
	_, err = worker.RunOnce(context.Background(), fixtureActorID)
	if err != nil || store.completed != OperationOffboard || executor.calls != 0 {
		t.Fatalf("err=%v completed=%s calls=%d", err, store.completed, executor.calls)
	}
}

func TestWorkerPurgeRequiresExecutorReceiptBeforeCompletion(t *testing.T) {
	store := &fixtureWorkerStore{claim: ClaimedOperation{
		Operation: Operation{Kind: OperationPurge}, SchemaName: "tenant_fixture", Proof: validProof(),
	}}
	executor := &fixtureExecutor{receipt: ExecutionReceipt{
		ID:     uuid.MustParse("99999999-9999-4999-8999-999999999999"),
		SHA256: [32]byte{9}, CreatedAt: fixtureNow,
	}}
	worker, _ := NewWorker(WorkerDependencies{Store: store, Executor: executor, Clock: fixtureClock{fixtureNow}})
	_, err := worker.RunOnce(context.Background(), fixtureActorID)
	if err != nil || store.completed != OperationPurge || executor.calls != 1 || store.failed != "" {
		t.Fatalf("err=%v completed=%s calls=%d failed=%s", err, store.completed, executor.calls, store.failed)
	}
}

func TestWorkerFailsClosedOnInvalidPurgeProof(t *testing.T) {
	store := &fixtureWorkerStore{claim: ClaimedOperation{
		Operation: Operation{Kind: OperationPurge}, SchemaName: "tenant_fixture",
	}}
	worker, _ := NewWorker(WorkerDependencies{
		Store: store, Executor: &fixtureExecutor{}, Clock: fixtureClock{fixtureNow},
	})
	_, err := worker.RunOnce(context.Background(), fixtureActorID)
	if !errors.Is(err, ErrPurgeProof) || store.failed != "PURGE_PROOF_INVALID" {
		t.Fatalf("err=%v failed=%s", err, store.failed)
	}
}
