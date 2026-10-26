//go:build integration

package postgres

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func assertPublisherStaleGeneration(
	t *testing.T,
	ctx context.Context,
	fixture publisherPGFixture,
) {
	t.Helper()
	tenantID := uuid.MustParse(asyncTenantA)
	seedPublisherEvent(t, ctx, fixture, publisherEventSeed{
		eventID: uuid.New(), tenantID: tenantID,
		aggregateID: uuid.New(), sequence: 1,
	})
	authority := publisherAuthorityFor(t, ctx, fixture, tenantID)
	claims, err := fixture.store.ClaimOutbox(
		ctx, authority, uuid.New(), 1, fixture.catalog,
	)
	if err != nil || len(claims) != 1 {
		t.Fatalf("claim generation event = %#v, %v", claims, err)
	}
	setPublisherDomainGeneration(t, ctx, fixture, tenantID, 2)
	if _, err := fixture.store.ConfirmPublish(
		ctx, claims[0], strings.Repeat("4", 64),
	); !errors.Is(err, ErrOutboxAuthorityStale) {
		t.Fatalf("stale authority confirmation error = %v", err)
	}
	setPublisherDomainGeneration(t, ctx, fixture, tenantID, 1)
	confirmPublisherClaim(t, ctx, fixture, claims[0], "5")
}

func setPublisherDomainGeneration(
	t *testing.T,
	ctx context.Context,
	fixture publisherPGFixture,
	tenantID uuid.UUID,
	generation int64,
) {
	t.Helper()
	result, err := fixture.owner.ExecContext(ctx, `
UPDATE platform.domain_cutovers
SET generation = $1
WHERE scope_kind = 'tenant' AND scope_id = $2 AND domain = 'workflow'`,
		generation, tenantID)
	if err != nil {
		t.Fatalf("set publisher authority generation: %v", err)
	}
	if affected, err := result.RowsAffected(); err != nil || affected != 1 {
		t.Fatalf("generation update affected %d rows: %v", affected, err)
	}
}
