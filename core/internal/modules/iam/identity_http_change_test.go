package iam

import (
	"context"
	"errors"
	"testing"

	changeapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/change"
	identityfacade "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type identityMutationSinkStub struct {
	calls  int
	change changeapp.Change
	err    error
}

func (sink *identityMutationSinkStub) WriteChange(
	_ context.Context,
	_ platformtx.TenantScope,
	change changeapp.Change,
) error {
	sink.calls++
	sink.change = change
	return sink.err
}

func TestFinishIdentityMutationSuppressesReplay(t *testing.T) {
	sink := &identityMutationSinkStub{}
	application := &identityHTTPApplication{changes: sink}
	err := application.finishIdentityMutation(
		context.Background(), platformtx.TenantScope{}, identityChangeEvidence(),
		uuid.MustParse("aa000000-0000-4000-8000-000000000003"),
		"user", "iam.identity.user.create", true,
	)
	if err != nil || sink.calls != 0 {
		t.Fatalf("replay emitted change: calls=%d error=%v", sink.calls, err)
	}
}

func TestFinishIdentityMutationPropagatesSinkFailure(t *testing.T) {
	sink := &identityMutationSinkStub{err: errors.New("outbox unavailable")}
	application := &identityHTTPApplication{changes: sink}
	err := application.finishIdentityMutation(
		context.Background(), platformtx.TenantScope{}, identityChangeEvidence(),
		uuid.MustParse("aa000000-0000-4000-8000-000000000003"),
		"user", "iam.identity.user.create", false,
	)
	if !errors.Is(err, changeapp.ErrSink) || sink.calls != 1 ||
		sink.change.CommandKey != "iam.identity.user.create" {
		t.Fatalf("sink failure was not propagated: calls=%d change=%+v error=%v",
			sink.calls, sink.change, err)
	}
}

func identityChangeEvidence() identityfacade.MutationEvidence {
	return identityfacade.MutationEvidence{
		OperationID:   identityfacade.IdentityID(uuid.MustParse("aa000000-0000-4000-8000-000000000001")),
		ActorID:       identityfacade.IdentityID(uuid.MustParse("aa000000-0000-4000-8000-000000000002")),
		PermissionID:  identityfacade.IdentityID(uuid.MustParse("aa000000-0000-4000-8000-000000000004")),
		RequestSHA256: [32]byte{1}, ExpectedTokenVersion: 1,
		ExpectedAuthorizationVersion: 1,
	}
}
