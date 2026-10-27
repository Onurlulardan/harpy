//go:build integration

package postgres

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

type publisherEventSeed struct {
	eventID      uuid.UUID
	tenantID     uuid.UUID
	aggregateID  uuid.UUID
	sequence     int64
	schemaSHA256 string
	groupGen     int64
	domainGen    int64
	occurredAt   time.Time
}

func seedPublisherEvent(
	t *testing.T,
	ctx context.Context,
	fixture publisherPGFixture,
	seed publisherEventSeed,
) {
	t.Helper()
	payload := []byte(fmt.Sprintf(`{"sequence":%d}`, seed.sequence))
	digest := sha256.Sum256(payload)
	occurredAt := seed.occurredAt
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC().Add(-time.Minute)
	}
	occurredAt = occurredAt.Truncate(time.Microsecond)
	schemaSHA256 := seed.schemaSHA256
	if schemaSHA256 == "" {
		schemaSHA256 = strings.Repeat("b", 64)
	}
	groupGeneration, domainGeneration := seed.groupGen, seed.domainGen
	if groupGeneration == 0 {
		groupGeneration = 1
	}
	if domainGeneration == 0 {
		domainGeneration = 1
	}
	arguments := []any{
		seed.eventID, seed.tenantID, "tenant", seed.tenantID,
		"operations", groupGeneration, "workflow", domainGeneration, "v2",
		"integration-event.v1", "workflow.changed", 1,
		"workflow.instance", seed.aggregateID, seed.sequence, seed.sequence,
		occurredAt, uuid.MustParse(asyncActor), "corr." + seed.eventID.String()[:8],
		uuid.MustParse(asyncCausation), "event-" + seed.eventID.String(),
		"effect-" + seed.eventID.String(), payload, hex.EncodeToString(digest[:]),
		schemaSHA256,
	}
	result, err := fixture.runtime.Exec(ctx, enqueueEventSQL, arguments...)
	if err != nil {
		t.Fatalf("seed publisher event %s: %v", seed.eventID, err)
	}
	if result.RowsAffected() != 1 {
		t.Fatalf("seed publisher event %s affected %d rows", seed.eventID, result.RowsAffected())
	}
}

func publisherAuthorityFor(
	t *testing.T,
	ctx context.Context,
	fixture publisherPGFixture,
	tenantID uuid.UUID,
) OutboxAuthority {
	t.Helper()
	token, err := fixture.store.BeginOutboxAuthorityPage(256)
	if err != nil {
		t.Fatal(err)
	}
	page, err := fixture.store.ListOutboxAuthorities(ctx, token)
	if err != nil {
		t.Fatal(err)
	}
	for _, authority := range page.Authorities() {
		if authority.scopeKind == "tenant" && authority.scopeID == tenantID &&
			authority.groupID == "operations" && authority.domain == "workflow" {
			return authority
		}
	}
	t.Fatalf("publisher authority for tenant %s was not listed", tenantID)
	return OutboxAuthority{}
}
