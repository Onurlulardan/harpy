package query

import (
	"context"
	"encoding/binary"
	"fmt"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type UpdateRequest struct {
	QueryID                          uuid.UUID
	ExpectedRootRevision             uint64
	ExpectedVersionRevision          uint64
	Name                             string
	QueryDocument, ParameterDocument []byte
	OutputDocument                   []byte
	IdempotencyKey                   string
}

func (application *Application) Update(
	ctx context.Context, scope platformtx.TenantScope, request UpdateRequest,
) (Snapshot, error) {
	queryDocument, queryErr := NewDocument(request.QueryDocument)
	parameters, parameterErr := NewDocument(request.ParameterDocument)
	output, outputErr := NewDocument(request.OutputDocument)
	if !validUUIDv7(request.QueryID) || request.ExpectedRootRevision == 0 ||
		request.ExpectedVersionRevision == 0 || !validName(request.Name) ||
		!validIdempotency(request.IdempotencyKey) || queryErr != nil ||
		parameterErr != nil || outputErr != nil {
		return Snapshot{}, ErrInvalid
	}
	store, identity, err := application.bind(ctx, scope, ScopeCommand,
		OperationUpdate, PermissionUpdate)
	if err != nil {
		return Snapshot{}, err
	}
	return application.updateBound(ctx, scope, store, identity, request,
		queryDocument, parameters, output)
}

func (application *Application) updateBound(
	ctx context.Context, scope platformtx.TenantScope, store Store,
	identity ScopeIdentity, request UpdateRequest,
	queryDocument, parameters, output Document,
) (Snapshot, error) {
	operationID, err := application.ids.NewUUIDv7()
	if err != nil || !validUUIDv7(operationID) {
		return Snapshot{}, ErrUnavailable
	}
	now := application.clock.Now()
	if !validTime(now) {
		return Snapshot{}, ErrUnavailable
	}
	revisions := make([]byte, 16)
	binary.BigEndian.PutUint64(revisions[:8], request.ExpectedRootRevision)
	binary.BigEndian.PutUint64(revisions[8:], request.ExpectedVersionRevision)
	digest := requestDigest([]byte(OperationUpdate), []byte(identity.TenantID.String()),
		[]byte(identity.ActorID.String()), request.QueryID[:], revisions, []byte(request.Name),
		digestBytes(queryDocument), digestBytes(parameters), digestBytes(output))
	decision, err := application.decideMutation(ctx, scope, identity,
		billingapi.MeteredResourceUpdate, operationID, request.QueryID,
		request.IdempotencyKey, digest)
	if err != nil {
		return Snapshot{}, err
	}
	result, replay, err := store.Update(ctx, UpdateMutation{OperationID: decision.OperationID,
		ActorID: identity.ActorID, QueryID: decision.OwnerFactID,
		IdempotencyKey: request.IdempotencyKey, RequestSHA256: digest,
		ExpectedRootRevision:    request.ExpectedRootRevision,
		ExpectedVersionRevision: request.ExpectedVersionRevision, Name: request.Name,
		Query: queryDocument, Parameters: parameters, Output: output, OccurredAt: now})
	if err != nil {
		return Snapshot{}, fmt.Errorf("update analytics query: %w", err)
	}
	if validateSnapshot(result) != nil || result.QueryID != decision.OwnerFactID ||
		replay != decision.Replay ||
		(!replay && (result.RootRevision != request.ExpectedRootRevision+1 ||
			result.VersionRevision != request.ExpectedVersionRevision+1)) {
		return Snapshot{}, ErrUnavailable
	}
	application.invalidate(ctx, request.QueryID)
	return result, nil
}

func digestBytes(value Document) []byte { digest := value.SHA256(); return digest[:] }
