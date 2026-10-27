package lifecycle

import (
	"context"
	"errors"
	"testing"
	"time"

	catalogcomponentv2 "github.com/Onurlulardan/harpy/contracts/catalogcomponent/v2"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/componentowner"
	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestPreparePurgeDoesNotCrossDestructiveBarrierWhenSecondOwnerBlocks(t *testing.T) {
	authority := purgeOrderAuthority()
	operationID, _ := packagecontract.ParseStableID(authority.OperationID)
	firstID, _ := packagecontract.ParseComponentID(
		"019fa200-0000-7000-8000-000000000011",
	)
	secondID, _ := packagecontract.ParseComponentID(
		"019fa200-0000-7000-8000-000000000012",
	)
	operation := catalogdomain.Operation{
		ID: operationID, Kind: catalogdomain.OperationPurge,
		State: catalogdomain.OperationApplying, Revision: 2,
		Plan: catalogdomain.Plan{
			Components: []catalogdomain.ComponentPin{
				{ComponentID: firstID, Owner: "analytics", OwnerGeneration: 1,
					Contract: "analytics.catalog.v2", ContentSHA256: digestA},
				{ComponentID: secondID, Owner: "catalog", OwnerGeneration: 1,
					Contract: "catalog.catalog.v2", ContentSHA256: digestA},
			},
		},
	}
	first := &purgeOrderOwner{}
	second := &purgeOrderOwner{blocked: true}
	store := &purgeOrderStore{}
	application := &Application{owners: purgeOrderDirectory{
		"analytics": first, "catalog": second,
	}}
	_, _, _, err := application.preparePurgeDestruction(
		context.Background(), platformtx.TenantScope{}, store,
		operation, authority, authority.CoolingUntil,
	)
	if !errors.Is(err, ErrDependencyBlocked) {
		t.Fatalf("preflight blocker error=%v", err)
	}
	if first.preflight != 1 || second.preflight != 1 ||
		first.begin != 0 || second.begin != 0 ||
		first.finalize != 0 || second.finalize != 0 || store.changes != 0 {
		t.Fatalf("calls first=%+v second=%+v destructive writes=%d",
			first, second, store.changes)
	}
}

func TestPreparePurgeResumeSkipsRetiredOnlyPreflightAndReturnsFutureRetry(t *testing.T) {
	authority := purgeOrderAuthority()
	operationID, _ := packagecontract.ParseStableID(authority.OperationID)
	componentID, _ := packagecontract.ParseComponentID(
		"019fa200-0000-7000-8000-000000000011",
	)
	started := authority.RequestedAt
	operation := catalogdomain.Operation{
		ID: operationID, Kind: catalogdomain.OperationPurge,
		State: catalogdomain.OperationApplying, Revision: 3,
		DestructiveStarted: &started,
		Plan: catalogdomain.Plan{Components: []catalogdomain.ComponentPin{{
			ComponentID: componentID, Owner: "forms", OwnerGeneration: 1,
			Contract: "forms.catalog-definition.v1", ContentSHA256: digestA,
		}}},
	}
	forms := &purgeOrderOwner{beginWaiting: true}
	application := &Application{owners: purgeOrderDirectory{"forms": forms}}
	_, retryAt, ready, err := application.preparePurgeDestruction(
		context.Background(), platformtx.TenantScope{}, &purgeOrderStore{},
		operation, authority, authority.RequestedAt.Add(time.Second),
	)
	if err != nil || ready || !retryAt.After(authority.RequestedAt) ||
		forms.preflight != 0 || forms.begin != 1 {
		t.Fatalf("retry/ready/error/calls = %s/%t/%v/%+v",
			retryAt, ready, err, forms)
	}
}

func TestCollectPurgeStateSeparatesReadyBeginFromWaitingFinalize(t *testing.T) {
	authority := purgeOrderAuthority()
	command := componentowner.PurgeCommand{Authority: authority,
		Component: componentowner.Component{
			ID: "019fa200-0000-7000-8000-000000000011", Owner: "forms",
			Generation: 1, Contract: "forms.catalog-definition.v1",
			ContentSHA256: digestA,
		}, IdempotencyKey: "collect-state-0001"}
	ready := purgeOrderResult(command, catalogcomponentv2.StatePurgePending)
	retryAt, waiting, blocked := collectPurgeState(time.Time{}, false, false, ready)
	if waiting || blocked || !retryAt.IsZero() {
		t.Fatalf("ready begin was treated as waiting: %s/%t/%t",
			retryAt, waiting, blocked)
	}
	pending := ready
	pending.Ready, pending.CompletedAt = false, time.Time{}
	pending.RetryAt = authority.RequestedAt.Add(time.Minute)
	pending.ResultSHA256 = catalogcomponentv2.ResultSHA256(pending)
	retryAt, waiting, blocked = collectPurgeState(time.Time{}, false, false, pending)
	if !waiting || blocked || retryAt != pending.RetryAt {
		t.Fatalf("waiting finalize was treated as ready: %s/%t/%t",
			retryAt, waiting, blocked)
	}
}
