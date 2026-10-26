package systemtransfer

import (
	"context"
	"errors"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type fixtureBatchScope struct{ tenantID, actorID uuid.UUID }

func (scope fixtureBatchScope) TenantID() uuid.UUID { return scope.tenantID }
func (scope fixtureBatchScope) ActorID() uuid.UUID  { return scope.actorID }
func (scope fixtureBatchScope) IAMScope() platformtx.TenantScope {
	return platformtx.TenantScope{}
}

type fixtureScopes struct{ calls int }

func (scopes *fixtureScopes) InFreshImportBatch(
	ctx context.Context,
	attempt Attempt,
	command func(context.Context, BatchScope) error,
) error {
	scopes.calls++
	return command(ctx, fixtureBatchScope{attempt.TenantID, attempt.ActorID})
}

type fixtureAuthorizer struct {
	tenantID, actorID uuid.UUID
	version           int64
	calls             int
	err               error
}

func (gate *fixtureAuthorizer) AuthorizeImportBatch(
	_ context.Context,
	_ platformtx.TenantScope,
	_ Attempt,
) (publicresource.ActorEvidence, error) {
	gate.calls++
	if gate.err != nil {
		return publicresource.ActorEvidence{}, gate.err
	}
	return publicresource.NewActorEvidence(publicresource.ID(gate.tenantID),
		publicresource.ID(gate.actorID), 3, gate.version)
}

type fixtureIdentityLookup struct {
	match IdentityMatch
	calls int
}

func (lookup *fixtureIdentityLookup) FindIdentity(
	context.Context,
	platformtx.TenantScope,
	publicresource.ActorEvidence,
	IdentityManifest,
	publicresource.FieldValues,
) (IdentityMatch, error) {
	lookup.calls++
	return lookup.match, nil
}

type fixtureInvitation struct {
	result InvitationResult
	inputs []InvitationInput
	err    error
}

func (command *fixtureInvitation) Invite(
	_ context.Context,
	_ platformtx.TenantScope,
	_ publicresource.ActorEvidence,
	input InvitationInput,
) (InvitationResult, error) {
	command.inputs = append(command.inputs, input)
	return command.result, command.err
}

type fixtureReceiptStore struct {
	existing map[int64]RowReceipt
	appended []RowReceipt
	failNext bool
}

func (store *fixtureReceiptStore) ExistingReceipts(
	context.Context,
	Attempt,
	[]int64,
) (map[int64]RowReceipt, error) {
	result := make(map[int64]RowReceipt, len(store.existing))
	for key, value := range store.existing {
		result[key] = value
	}
	return result, nil
}

func (store *fixtureReceiptStore) AppendReceipts(
	_ context.Context,
	_ Attempt,
	receipts []RowReceipt,
	_ int64,
) error {
	if store.failNext {
		store.failNext = false
		return errors.New("receipt commit interrupted")
	}
	store.appended = append(store.appended, receipts...)
	if store.existing == nil {
		store.existing = make(map[int64]RowReceipt)
	}
	for _, receipt := range receipts {
		store.existing[receipt.RowIndex] = receipt
	}
	return nil
}
