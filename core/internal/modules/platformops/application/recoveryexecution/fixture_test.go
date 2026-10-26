package recoveryexecution

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformops/ports"
)

const (
	testSetID      = "10000000-0000-4000-8000-000000000001"
	testAttemptID  = "01890f3e-7b10-7cc0-98c4-dc0c0c07398f"
	testLeaseToken = "01890f3e-7b10-7cc0-a8c4-dc0c0c073990"
	testDigest     = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
)

type executionStoreFixture struct {
	claim    ports.RecoveryExecutionClaim
	found    bool
	changed  bool
	putCalls int
}

func (fixture *executionStoreFixture) ClaimNextRecoverySet(
	context.Context, string, string,
) (ports.RecoveryExecutionClaim, bool, error) {
	return fixture.claim, fixture.found, nil
}

func (fixture *executionStoreFixture) UpdateRecoveryLease(
	context.Context, ports.RecoveryExecutionIdentity,
) (bool, error) {
	return fixture.changed, nil
}

func (fixture *executionStoreFixture) StartRecoverySnapshot(
	context.Context, ports.RecoveryExecutionIdentity,
) (bool, error) {
	return fixture.changed, nil
}

func (fixture *executionStoreFixture) StartRecoveryCopy(
	context.Context, ports.RecoveryExecutionIdentity,
) (bool, error) {
	return fixture.changed, nil
}

func (fixture *executionStoreFixture) StartRecoveryVerify(
	context.Context, ports.RecoveryExecutionIdentity,
) (bool, error) {
	return fixture.changed, nil
}

func (fixture *executionStoreFixture) AttestDocumentsHold(
	context.Context, ports.DocumentsHoldAttestation,
) (bool, error) {
	return fixture.changed, nil
}

func (fixture *executionStoreFixture) RecordRecoveryComponent(
	context.Context, ports.RecoveryComponentRecord,
) (bool, error) {
	return fixture.changed, nil
}

func (fixture *executionStoreFixture) PutRecoveryComponents(
	context.Context, ports.RecoveryComponentBatch,
) (bool, error) {
	fixture.putCalls++
	return fixture.changed, nil
}

func (fixture *executionStoreFixture) RejectRecoveryExecution(
	context.Context, ports.RecoveryExecutionFailure,
) (bool, error) {
	return fixture.changed, nil
}

func (fixture *executionStoreFixture) CompleteRecoverySet(
	context.Context, ports.RecoveryCompletionRecord,
) (bool, error) {
	return fixture.changed, nil
}

func testIdentity() ports.RecoveryExecutionIdentity {
	return ports.RecoveryExecutionIdentity{
		RecoverySetID: testSetID, AttemptID: testAttemptID, LeaseToken: testLeaseToken,
	}
}

func testService(fixture *executionStoreFixture) *Service {
	service, _ := New(fixture)
	return service
}
