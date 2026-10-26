package assignment

import (
	"context"
	"errors"
	"testing"

	rbacdomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/rbac"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type assignmentChangeSinkStub struct {
	calls  int
	scope  platformtx.TenantScope
	change rbacdomain.Change
	err    error
}

func (sink *assignmentChangeSinkStub) WriteChange(
	_ context.Context,
	scope platformtx.TenantScope,
	change rbacdomain.Change,
) error {
	sink.calls++
	sink.scope = scope
	sink.change = change
	return sink.err
}

func TestAssignmentFinishFailsClosedInSameScopeWhenSinkFails(t *testing.T) {
	sink := &assignmentChangeSinkStub{err: errors.New("outbox failed")}
	commands := &Commands{sink: sink}
	scope := platformtx.TenantScope{}
	result := rbacdomain.MutationResult{
		ID: rbacdomain.ID{1}, Revision: 2, AuthorizationClock: 3,
	}
	evidence := rbacdomain.MutationEvidence{
		OperationID: rbacdomain.ID{2}, RequestSHA256: [32]byte{3},
		ActorID: rbacdomain.ID{4}, RequiredPermissionID: rbacdomain.ID{5},
		ExpectedTokenVersion: 1, ExpectedAuthorizationVersion: 1,
	}
	_, err := commands.finish(
		context.Background(), scope, evidence, result,
		"iam.rbac.assignment.create", "iam.rbac.assignment.created",
	)
	if !errors.Is(err, rbacdomain.ErrSink) || sink.calls != 1 ||
		sink.scope != scope || sink.change.AggregateID != result.ID {
		t.Fatalf("sink failure=%v calls=%d change=%#v", err, sink.calls, sink.change)
	}
}

func TestAssignmentFinishDoesNotDuplicateAuditOrOutboxOnReplay(t *testing.T) {
	sink := &assignmentChangeSinkStub{}
	commands := &Commands{sink: sink}
	result := rbacdomain.MutationResult{
		ID: rbacdomain.ID{1}, Revision: 2, AuthorizationClock: 3, Replayed: true,
	}
	returned, err := commands.finish(
		context.Background(), platformtx.TenantScope{},
		rbacdomain.MutationEvidence{}, result, "ignored", "ignored",
	)
	if err != nil || returned != result || sink.calls != 0 {
		t.Fatalf("replay=%#v error=%v sink calls=%d", returned, err, sink.calls)
	}
}
