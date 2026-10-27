package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Row interface{ Scan(...any) error }
type QueryRow func(context.Context, string, ...any) Row

type Store struct {
	queryRow QueryRow
}

func NewStore(queryRow QueryRow) *Store {
	return &Store{queryRow: queryRow}
}

func (store *Store) WarmQuery(
	ctx context.Context, queryID, versionID uuid.UUID, manifest [32]byte,
) error {
	if !store.valid(ctx) || queryID == uuid.Nil || versionID == uuid.Nil ||
		manifest == [32]byte{} {
		return fmt.Errorf("Analytics cache warm input is invalid")
	}
	var returned []byte
	var documentBytes, dependencies int64
	err := store.queryRow(ctx, cacheWarmSQL,
		queryID, versionID, manifest[:]).Scan(
		&returned, &documentBytes, &dependencies)
	if err != nil {
		return fmt.Errorf("warm Analytics query metadata: %w", err)
	}
	if len(returned) != 32 || documentBytes < 6 || dependencies < 1 {
		return fmt.Errorf("Analytics query metadata differs")
	}
	return nil
}

func (store *Store) VerifyRevalidated(
	ctx context.Context, queryID, versionID uuid.UUID, manifest [32]byte,
) error {
	if !store.valid(ctx) || queryID == uuid.Nil || versionID == uuid.Nil ||
		manifest == [32]byte{} {
		return fmt.Errorf("Analytics revalidation input is invalid")
	}
	var returned []byte
	var revision int64
	err := store.queryRow(ctx, queryRevalidateSQL,
		queryID, versionID, manifest[:]).Scan(&returned, &revision)
	if err != nil {
		return fmt.Errorf("verify Analytics revalidation: %w", err)
	}
	if len(returned) != 32 || revision < 1 {
		return fmt.Errorf("Analytics revalidation result differs")
	}
	return nil
}

func (store *Store) ApplyPublicationRepair(
	ctx context.Context, releaseID uuid.UUID, generation uint64, observedAt time.Time,
) error {
	if !store.valid(ctx) || releaseID == uuid.Nil || generation < 1 ||
		observedAt.IsZero() || observedAt.Location() != time.UTC {
		return fmt.Errorf("Analytics publication repair input is invalid")
	}
	var repaired bool
	err := store.queryRow(ctx, publicationRepairSQL,
		releaseID, generation, observedAt).Scan(&repaired)
	if err != nil {
		return fmt.Errorf("repair Analytics publication: %w", err)
	}
	if !repaired {
		return fmt.Errorf("Analytics publication was not repaired")
	}
	return nil
}

func (store *Store) valid(ctx context.Context) bool {
	return store != nil && store.queryRow != nil && ctx != nil && ctx.Err() == nil
}
