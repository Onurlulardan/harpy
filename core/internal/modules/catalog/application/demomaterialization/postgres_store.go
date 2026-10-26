package demomaterialization

import (
	"context"
	"fmt"
)

type Row interface {
	Scan(...any) error
}

type QueryRow func(context.Context, string, ...any) Row
type ScopeGuard func(context.Context) bool

type BoundStore struct {
	generation int64
	query      string
	queryRow   QueryRow
	allowed    ScopeGuard
}

func NewBoundStore(
	generation int64,
	query string,
	queryRow QueryRow,
	allowed ScopeGuard,
) (*BoundStore, error) {
	if generation < 1 || query == "" || queryRow == nil || allowed == nil {
		return nil, ErrInvalid
	}
	return &BoundStore{
		generation: generation, query: query,
		queryRow: queryRow, allowed: allowed,
	}, nil
}

func (store *BoundStore) Project(
	ctx context.Context,
	input Projection,
) (Snapshot, error) {
	value := input.Materialization
	if !store.usable(ctx) || input.ProposedReceiptID.Version() != 7 ||
		input.PackageID == [16]byte{} || !validPackageSlug(input.PackageSlug) ||
		input.ReleaseID.Version() != 7 ||
		input.Component.Owner != owner ||
		input.Component.Contract != catalogContract() ||
		input.Component.Generation != store.generation ||
		value.ModuleSlug != input.PackageSlug ||
		value.SourceContract == "" || len(value.Document) == 0 ||
		value.SemanticSHA256 == [32]byte{} ||
		value.FormCount < 1 || value.RowCount < 1 ||
		input.IdempotencyKey == "" || input.RequestSHA256 == [32]byte{} ||
		input.OccurredAt.IsZero() {
		return Snapshot{}, ErrInvalid
	}
	snapshot, err := scanSnapshot(store.queryRow(
		ctx, store.query,
		input.ProposedReceiptID, input.PackageID, input.PackageSlug,
		input.PackageVersion, input.ReleaseID, input.Component.ID,
		input.Component.Generation, input.Component.Contract,
		input.Component.ContentSHA256, value.ModuleSlug,
		value.SourceContract, value.SemanticSHA256[:],
		value.FormCount, value.RowCount, value.Document,
		input.IdempotencyKey, input.RequestSHA256[:], input.OccurredAt,
	))
	if err != nil {
		return Snapshot{}, fmt.Errorf("execute Catalog demo materialization: %w", err)
	}
	return snapshot, nil
}

func (store *BoundStore) usable(ctx context.Context) bool {
	return store != nil && store.generation > 0 && store.query != "" &&
		store.queryRow != nil && store.allowed != nil && store.allowed(ctx)
}

var _ Store = (*BoundStore)(nil)
