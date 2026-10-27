package query

import (
	"context"
	"encoding/binary"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type PublishRequest struct {
	QueryID               uuid.UUID
	DraftVersionID        uuid.UUID
	ExpectedRootRevision  uint64
	ExpectedDraftRevision uint64
	Dependencies          []DependencyRequest
	RevisionBoundCache    bool
	IdempotencyKey        string
}

func validPublishRequest(request PublishRequest) bool {
	return validUUIDv7(request.QueryID) && validUUIDv7(request.DraftVersionID) &&
		request.ExpectedRootRevision > 0 && request.ExpectedDraftRevision > 0 &&
		len(request.Dependencies) > 0 && len(request.Dependencies) <= 32 &&
		validIdempotency(request.IdempotencyKey)
}

func publishDigest(
	operation string, identity ScopeIdentity, request PublishRequest,
	contract NormalizedContract, manifest [32]byte,
) [32]byte {
	revisions := make([]byte, 16)
	binary.BigEndian.PutUint64(revisions[:8], request.ExpectedRootRevision)
	binary.BigEndian.PutUint64(revisions[8:], request.ExpectedDraftRevision)
	return requestDigest([]byte(operation), []byte(identity.TenantID.String()),
		[]byte(identity.ActorID.String()), request.QueryID[:], request.DraftVersionID[:],
		revisions, digestBytes(contract.Query), digestBytes(contract.Parameters),
		digestBytes(contract.Output), manifest[:])
}

func (application *Application) resolvePublication(
	ctx context.Context, scope platformtx.TenantScope, snapshot Snapshot,
	request PublishRequest,
) (NormalizedContract, []DependencySnapshot, [32]byte, error) {
	resolved, err := application.dependencies.Dependencies(ctx, scope, request.Dependencies)
	if err != nil {
		return NormalizedContract{}, nil, [32]byte{}, ErrOwnerNotV2
	}
	manifest, err := dependencyManifest(request.Dependencies, resolved, request.RevisionBoundCache)
	if err != nil {
		return NormalizedContract{}, nil, [32]byte{}, err
	}
	queryDocument, queryErr := NewDocument(snapshot.QueryDocument)
	parameters, parameterErr := NewDocument(snapshot.ParameterDocument)
	output, outputErr := NewDocument(snapshot.OutputDocument)
	if queryErr != nil || parameterErr != nil || outputErr != nil {
		return NormalizedContract{}, nil, [32]byte{}, ErrUnavailable
	}
	contract, err := application.contracts.Normalize(ctx, scope, queryDocument,
		parameters, output, resolved)
	if err != nil || !contract.Query.Valid() || !contract.Parameters.Valid() ||
		!contract.Output.Valid() {
		return NormalizedContract{}, nil, [32]byte{}, ErrInvalid
	}
	return contract, resolved, manifest.SHA256(), nil
}
