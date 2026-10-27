package profilebackfill

import (
	"context"
	"crypto/sha256"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

const MaximumPageSize = 250

type ReadQuery struct {
	ResourceKind    string
	SourceSchemaID  uuid.UUID
	AfterResourceID uuid.UUID
	Limit           int
	Evidence        Evidence
}

func (query ReadQuery) Validate() error {
	if (query.ResourceKind != "user" && query.ResourceKind != "organization" &&
		query.ResourceKind != "role") || !stableID(query.SourceSchemaID) ||
		query.Limit < 1 || query.Limit > MaximumPageSize ||
		query.AfterResourceID != uuid.Nil && !stableID(query.AfterResourceID) ||
		query.Evidence.Validate() != nil {
		return ErrInvalid
	}
	return nil
}

type SourceResource struct {
	ID              uuid.UUID
	SchemaVersionID uuid.UUID
	Revision        int64
	Document        []byte
	SHA256          [sha256.Size]byte
}

func (resource SourceResource) Validate() error {
	if !stableID(resource.ID) || !stableID(resource.SchemaVersionID) ||
		resource.Revision < 1 || len(resource.Document) < 2 ||
		resource.SHA256 == [sha256.Size]byte{} {
		return ErrInvalid
	}
	return nil
}

type Receipt struct {
	CommandKey         string
	RequestSHA256      [sha256.Size]byte
	ResultID           uuid.UUID
	TargetSchemaID     uuid.UUID
	Revision           int64
	AuthorizationClock int64
}

func (receipt Receipt) Validate() error {
	if (receipt.CommandKey != "iam.profile_schema.publish" &&
		receipt.CommandKey != "iam.profile_backfill.plan_create") ||
		receipt.RequestSHA256 == [sha256.Size]byte{} || !stableID(receipt.ResultID) ||
		receipt.Revision < 1 || receipt.AuthorizationClock < 1 ||
		receipt.CommandKey == "iam.profile_backfill.plan_create" &&
			!stableID(receipt.TargetSchemaID) {
		return ErrInvalid
	}
	return nil
}

type SourcePort interface {
	Read(context.Context, platformtx.TenantScope, ReadQuery) ([]SourceResource, error)
	Lookup(
		context.Context,
		platformtx.TenantScope,
		Evidence,
	) (Receipt, bool, error)
}

type Reader struct {
	port SourcePort
}

func NewReader(port SourcePort) (*Reader, error) {
	if interfaceNil(port) {
		return nil, ErrUnavailable
	}
	return &Reader{port: port}, nil
}

func (reader *Reader) Read(
	ctx context.Context,
	scope platformtx.TenantScope,
	query ReadQuery,
) ([]SourceResource, error) {
	if reader == nil || reader.port == nil || query.Validate() != nil ||
		!sourceScopeUsable(ctx, scope, query.Evidence) {
		return nil, ErrInvalid
	}
	resources, err := reader.port.Read(ctx, scope, query)
	if err != nil {
		return nil, err
	}
	if len(resources) > query.Limit {
		return nil, ErrUnavailable
	}
	for _, resource := range resources {
		if resource.Validate() != nil {
			return nil, ErrUnavailable
		}
	}
	return resources, nil
}

func (reader *Reader) Lookup(
	ctx context.Context,
	scope platformtx.TenantScope,
	evidence Evidence,
) (Receipt, bool, error) {
	if reader == nil || reader.port == nil || evidence.Validate() != nil ||
		!sourceScopeUsable(ctx, scope, evidence) {
		return Receipt{}, false, ErrInvalid
	}
	receipt, found, err := reader.port.Lookup(ctx, scope, evidence)
	if err != nil || !found {
		return Receipt{}, found, err
	}
	if receipt.Validate() != nil {
		return Receipt{}, false, ErrUnavailable
	}
	return receipt, true, nil
}

func sourceScopeUsable(
	ctx context.Context,
	scope platformtx.TenantScope,
	evidence Evidence,
) bool {
	return ctx != nil && ctx.Err() == nil && !platformtx.NetworkAllowed(ctx) &&
		platformtx.ContextOwnsScope(ctx, scope) && scope.ActorID() == evidence.ActorID &&
		platformtx.RequireActive(scope, platformtx.Command) == nil
}
