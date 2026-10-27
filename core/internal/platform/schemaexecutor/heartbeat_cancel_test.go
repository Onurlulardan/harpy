package schemaexecutor

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/admission"
)

type heartbeatFailurePersistence struct {
	record        ClaimRecord
	heartbeatErr  error
	applyCanceled chan struct{}
}

func (*heartbeatFailurePersistence) RecoverExpired(
	context.Context,
) (bool, error) {
	return false, nil
}

func (persistence *heartbeatFailurePersistence) Claim(
	_ context.Context,
	request ClaimRequest,
) (ClaimRecord, bool, error) {
	persistence.record.Lease.LeaseID = request.LeaseID
	persistence.record.Lease.ExecutorID = request.ExecutorID
	return persistence.record, true, nil
}

func (*heartbeatFailurePersistence) MarkRunning(
	context.Context,
	Lease,
) error {
	return nil
}

func (persistence *heartbeatFailurePersistence) Heartbeat(
	context.Context,
	HeartbeatRequest,
) error {
	return persistence.heartbeatErr
}

func (persistence *heartbeatFailurePersistence) Apply(
	ctx context.Context,
	_ ApplyRequest,
) error {
	<-ctx.Done()
	close(persistence.applyCanceled)
	return ctx.Err()
}

func (*heartbeatFailurePersistence) RecordFailure(
	context.Context,
	FailureRequest,
) error {
	return errors.New("failure must not be recorded after heartbeat loss")
}

func TestHeartbeatLossCancelsInFlightApply(t *testing.T) {
	now := testClock()
	expected := errors.New("lease heartbeat lost")
	persistence := &heartbeatFailurePersistence{
		record: testClaimRecord(t, now), heartbeatErr: expected,
		applyCanceled: make(chan struct{}),
	}
	executor, err := New(Dependencies{
		Persistence: persistence, ExecutorID: testExecutorID,
		NewLeaseID: func() (string, error) { return testLeaseID, nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	executor.heartbeatInterval = time.Millisecond
	gate := &admission.Gate{}
	gate.Open()
	permit, admitted := gate.Admit()
	if !admitted {
		t.Fatal("test admission did not open")
	}
	processed, err := executor.processOne(context.Background(), permit)
	if !processed || !errors.Is(err, expected) {
		t.Fatalf("processed=%v error=%v", processed, err)
	}
	select {
	case <-persistence.applyCanceled:
	default:
		t.Fatal("heartbeat loss did not cancel apply")
	}
}
