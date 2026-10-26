package catalog

import (
	"context"
	"errors"
	"testing"

	lifecycleapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type resumeScopeFixture struct {
	calls              int
	pending, committed []string
}

func (*resumeScopeFixture) InJobTenant(
	context.Context, uuid.UUID,
	func(context.Context, platformtx.JobTenantScope) error,
) error {
	return errors.New("not used")
}

func (fixture *resumeScopeFixture) InTenant(
	ctx context.Context,
	_ platformtx.ScopeRequest,
	call func(context.Context, platformtx.TenantScope) error,
) error {
	fixture.calls++
	fixture.pending = nil
	err := call(ctx, platformtx.TenantScope{})
	if err == nil {
		fixture.committed = append(fixture.committed, fixture.pending...)
	}
	fixture.pending = nil
	return err
}

func TestResumeFailureRollsBackOwnerEffectsBeforeDurableSettlementScope(t *testing.T) {
	scopes := &resumeScopeFixture{}
	err := runResumeScopes(
		context.Background(), scopes, platformtx.ScopeRequest{},
		func(context.Context, platformtx.TenantScope) error {
			scopes.pending = append(scopes.pending, "owner-a-destructive-effect")
			return &resumeFailure{cause: lifecycleapp.ErrOwnerReceiptStale}
		},
		func(_ context.Context, _ platformtx.TenantScope, cause error) error {
			if len(scopes.committed) != 0 || !errors.Is(cause, lifecycleapp.ErrOwnerReceiptStale) {
				t.Fatalf("first scope leaked=%v cause=%v", scopes.committed, cause)
			}
			scopes.pending = append(scopes.pending, "blocked-outcome-and-settlement")
			return nil
		},
	)
	if err != nil || scopes.calls != 2 ||
		len(scopes.committed) != 1 ||
		scopes.committed[0] != "blocked-outcome-and-settlement" {
		t.Fatalf("scopes=%d committed=%v err=%v", scopes.calls, scopes.committed, err)
	}
}

func TestResumeUnknownFailureDoesNotOpenSettlementScope(t *testing.T) {
	scopes := &resumeScopeFixture{}
	unknown := errors.New("unknown")
	err := runResumeScopes(
		context.Background(), scopes, platformtx.ScopeRequest{},
		func(context.Context, platformtx.TenantScope) error { return unknown },
		func(context.Context, platformtx.TenantScope, error) error {
			t.Fatal("unknown failure reached durable settlement")
			return nil
		},
	)
	if !errors.Is(err, unknown) || scopes.calls != 1 {
		t.Fatalf("scopes=%d err=%v", scopes.calls, err)
	}
}
