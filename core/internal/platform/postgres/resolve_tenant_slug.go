package postgres

import (
	"context"
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (pool *Pool) ResolveActiveTenantIDBySlug(
	ctx context.Context,
	slug string,
) (uuid.UUID, error) {
	if ctx == nil || !validTenantSlugLookup(slug) {
		return uuid.Nil, ErrTenantMissing
	}
	transaction, err := pool.BeginRuntimeTransaction(ctx, TransactionSnapshotQuery)
	if err != nil {
		return uuid.Nil, err
	}
	tenantID, err := transaction.resolveActiveTenantIDBySlug(ctx, slug)
	if err != nil {
		_ = transaction.Rollback()
		return uuid.Nil, err
	}
	if err := transaction.Commit(ctx); err != nil {
		return uuid.Nil, err
	}
	return tenantID, nil
}

func (transaction *RuntimeTransaction) resolveActiveTenantIDBySlug(
	ctx context.Context,
	slug string,
) (uuid.UUID, error) {
	if !transaction.usable() || !validTenantSlugLookup(slug) {
		return uuid.Nil, ErrTransactionClosed
	}
	var tenantID uuid.UUID
	var status string
	var notDeleted bool
	err := transaction.transaction.QueryRow(ctx, resolveTenantSlugSQL, slug).Scan(
		&tenantID, &status, &notDeleted,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, ErrTenantMissing
	}
	if err != nil {
		return uuid.Nil, err
	}
	if tenantID == uuid.Nil || !notDeleted || status != "ACTIVE" && status != "TRIAL" {
		return uuid.Nil, ErrTenantInactive
	}
	return tenantID, nil
}

func validTenantSlugLookup(slug string) bool {
	if len(slug) == 0 || len(slug) > 100 || !utf8.ValidString(slug) ||
		strings.TrimSpace(slug) != slug {
		return false
	}
	for _, character := range slug {
		if unicode.IsControl(character) {
			return false
		}
	}
	return true
}
