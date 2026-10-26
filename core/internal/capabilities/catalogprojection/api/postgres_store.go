package catalogprojection

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
	owner      string
	contract   string
	generation int64
	query      string
	kinds      map[string]bool
	queryRow   QueryRow
	allowed    ScopeGuard
}

func (store *BoundStore) Project(
	ctx context.Context,
	input Projection,
) (Snapshot, error) {
	materialization := input.Materialization
	if !store.usable(ctx) || input.ProposedReceiptID.Version() != 7 ||
		input.PackageID == [16]byte{} || !validPackageSlug(input.PackageSlug) ||
		input.ReleaseID.Version() != 7 ||
		input.Component.Owner != store.owner ||
		input.Component.Contract != store.contract ||
		input.Component.Generation != store.generation ||
		!store.kinds[materialization.Kind] ||
		!validMaterialization(materialization, materialization.Document) ||
		input.IdempotencyKey == "" || input.RequestSHA256 == [32]byte{} ||
		input.OccurredAt.IsZero() {
		return Snapshot{}, ErrInvalid
	}
	snapshot, err := scanSnapshot(store.queryRow(
		ctx, store.query,
		input.ProposedReceiptID, input.PackageID, input.PackageSlug,
		input.PackageVersion,
		input.ReleaseID, input.Component.ID, input.Component.Generation,
		input.Component.Contract, input.Component.ContentSHA256,
		materialization.Kind, materialization.ModuleSlug,
		materialization.Identity,
		materialization.SemanticSHA256[:], materialization.ItemCount,
		materialization.AuxiliaryCount, materialization.Document,
		input.IdempotencyKey, input.RequestSHA256[:], input.OccurredAt,
	))
	if err != nil {
		return Snapshot{}, fmt.Errorf("execute catalog materialization: %w", err)
	}
	return snapshot, nil
}

func (store *BoundStore) usable(ctx context.Context) bool {
	return store != nil && ownerPattern.MatchString(store.owner) &&
		store.contract != "" && store.generation > 0 && store.query != "" &&
		len(store.kinds) > 0 && store.queryRow != nil &&
		store.allowed != nil && store.allowed(ctx)
}

var _ Store = (*BoundStore)(nil)
