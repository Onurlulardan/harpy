package query

import (
	"context"
	"fmt"

	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/api"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type CreateRequest struct {
	Name              string
	QueryDocument     []byte
	ParameterDocument []byte
	OutputDocument    []byte
	IdempotencyKey    string
}

func (application *Application) Create(
	ctx context.Context, scope platformtx.TenantScope, request CreateRequest,
) (Snapshot, error) {
	queryDocument, queryErr := NewDocument(request.QueryDocument)
	parameters, parameterErr := NewDocument(request.ParameterDocument)
	output, outputErr := NewDocument(request.OutputDocument)
	if !validName(request.Name) || !validIdempotency(request.IdempotencyKey) ||
		queryErr != nil || parameterErr != nil || outputErr != nil {
		return Snapshot{}, ErrInvalid
	}
	store, identity, err := application.bind(ctx, scope, ScopeCommand,
		OperationCreate, PermissionCreate)
	if err != nil {
		return Snapshot{}, err
	}
	return application.createBound(ctx, scope, store, identity, request,
		queryDocument, parameters, output)
}

func (application *Application) createBound(
	ctx context.Context, scope platformtx.TenantScope, store Store,
	identity ScopeIdentity, request CreateRequest,
	queryDocument, parameters, output Document,
) (Snapshot, error) {
	operationID, queryID, versionID, err := application.threeIDs()
	if err != nil {
		return Snapshot{}, err
	}
	now := application.clock.Now()
	if !validTime(now) {
		return Snapshot{}, ErrUnavailable
	}
	digest := requestDigest([]byte(OperationCreate), []byte(identity.TenantID.String()),
		[]byte(identity.ActorID.String()), []byte(request.Name), digestBytes(queryDocument),
		digestBytes(parameters), digestBytes(output))
	decision, err := application.decideMutation(ctx, scope, identity,
		billingapi.MeteredResourceCreate, operationID, queryID,
		request.IdempotencyKey, digest)
	if err != nil {
		return Snapshot{}, err
	}
	result, replay, err := store.Create(ctx, CreateMutation{OperationID: decision.OperationID,
		ActorID: identity.ActorID, QueryID: decision.OwnerFactID, VersionID: versionID,
		IdempotencyKey: request.IdempotencyKey, RequestSHA256: digest, Name: request.Name,
		Query: queryDocument, Parameters: parameters, Output: output, OccurredAt: now})
	if err != nil {
		return Snapshot{}, fmt.Errorf("create analytics query: %w", err)
	}
	if err := validateSnapshot(result); err != nil || result.QueryID != decision.OwnerFactID ||
		replay != decision.Replay || (!replay && result.VersionID != versionID) {
		return Snapshot{}, ErrUnavailable
	}
	return result, nil
}

func (application *Application) threeIDs() (first, second, third uuid.UUID, err error) {
	values := []*uuid.UUID{&first, &second, &third}
	for _, target := range values {
		value, nextErr := application.ids.NewUUIDv7()
		if nextErr != nil || !validUUIDv7(value) {
			return first, second, third, ErrUnavailable
		}
		*target = value
	}
	if first == second || first == third || second == third {
		return first, second, third, ErrUnavailable
	}
	return first, second, third, nil
}
