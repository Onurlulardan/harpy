package schemaexecutor

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/admission"
)

type drainPersistence struct {
	record    ClaimRecord
	admission *admission.Gate
	claimed   bool
}

func (persistence *drainPersistence) RecoverExpired(
	context.Context,
) (bool, error) {
	return false, nil
}

func (persistence *drainPersistence) Claim(
	_ context.Context,
	request ClaimRequest,
) (ClaimRecord, bool, error) {
	if persistence.claimed {
		return ClaimRecord{}, false, nil
	}
	persistence.claimed = true
	persistence.record.Lease.LeaseID = request.LeaseID
	persistence.record.Lease.ExecutorID = request.ExecutorID
	return persistence.record, true, nil
}

func (*drainPersistence) MarkRunning(context.Context, Lease) error {
	return nil
}

func (*drainPersistence) Heartbeat(context.Context, HeartbeatRequest) error {
	return nil
}

func (persistence *drainPersistence) Apply(
	context.Context,
	ApplyRequest,
) error {
	persistence.admission.BeginDrain()
	return nil
}

func (*drainPersistence) RecordFailure(context.Context, FailureRequest) error {
	return nil
}

func TestRunFinishesInFlightStepThenReturnsOnDrain(t *testing.T) {
	now := testClock()
	admission := &admission.Gate{}
	admission.Open()
	persistence := &drainPersistence{
		record: testClaimRecord(t, now), admission: admission,
	}
	executor, err := New(Dependencies{
		Persistence: persistence, ExecutorID: testExecutorID,
		NewLeaseID: func() (string, error) { return testLeaseID, nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := executor.Run(context.Background(), admission); err != nil {
		t.Fatal(err)
	}
	if !persistence.claimed || !admission.Draining() {
		t.Fatal("executor did not finish the in-flight step before drain")
	}
}
