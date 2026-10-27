package resource

import (
	"context"
	"errors"
	"testing"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type changeSinkStub struct {
	calls  int
	change rbacdomain.Change
	err    error
}

func (sink *changeSinkStub) WriteChange(_ context.Context, _ platformtx.TenantScope,
	change rbacdomain.Change) error {
	sink.calls++
	sink.change = change
	return sink.err
}

func TestFinishFailsClosedWhenRequiredSinkFails(t *testing.T) {
	sinkFailure := errors.New("sink failed")
	sink := &changeSinkStub{err: sinkFailure}
	commands := &Commands{sink: sink}
	result := rbacdomain.MutationResult{ID: rbacdomain.ID{1}, Revision: 2, AuthorizationClock: 3}
	evidence := rbacdomain.MutationEvidence{
		OperationID: rbacdomain.ID{2}, RequestSHA256: [32]byte{3}, ActorID: rbacdomain.ID{4},
		RequiredPermissionID: rbacdomain.ID{5}, ExpectedTokenVersion: 1,
		ExpectedAuthorizationVersion: 1,
	}
	if _, err := commands.finish(context.Background(), platformtx.TenantScope{}, evidence,
		result, "iam.rbac.resource.update", "iam.rbac.resource.updated"); !errors.Is(err, rbacdomain.ErrSink) {
		t.Fatalf("sink failure error = %v", err)
	}
	if sink.calls != 1 || sink.change.AggregateID != result.ID {
		t.Fatal("required change sink did not receive the exact aggregate")
	}
}

func TestFinishSkipsDuplicateSinkWriteForReplay(t *testing.T) {
	sink := &changeSinkStub{}
	commands := &Commands{sink: sink}
	result := rbacdomain.MutationResult{
		ID: rbacdomain.ID{1}, Revision: 2, AuthorizationClock: 3, Replayed: true,
	}
	returned, err := commands.finish(context.Background(), platformtx.TenantScope{},
		rbacdomain.MutationEvidence{}, result, "ignored", "ignored")
	if err != nil || returned != result {
		t.Fatalf("replay result = %#v, error = %v", returned, err)
	}
	if sink.calls != 0 {
		t.Fatal("replayed mutation emitted a duplicate change")
	}
}
