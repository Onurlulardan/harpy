//go:build integration

package postgres

import (
	"context"
	"errors"
	"strings"
	"testing"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	"github.com/google/uuid"
)

func assertPublisherCatalogFence(
	t *testing.T,
	ctx context.Context,
	fixture publisherPGFixture,
) {
	t.Helper()
	if err := fixture.store.AssertOutboxCatalog(ctx, fixture.catalog); err != nil {
		t.Fatalf("exact live outbox catalog: %v", err)
	}
	seedPublisherEvent(t, ctx, fixture, publisherEventSeed{
		eventID: uuid.New(), tenantID: uuid.MustParse(asyncTenantA),
		aggregateID: uuid.New(), sequence: 1, schemaSHA256: strings.Repeat("c", 64),
		groupGen: 2, domainGen: 2,
	})
	if err := fixture.store.AssertOutboxCatalog(
		ctx, fixture.catalog,
	); !errors.Is(err, ErrOutboxCatalogIncompatible) {
		t.Fatalf("unsupported live outbox catalog error = %v", err)
	}
	empty, err := eventsv1.NewProducerCatalog(nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := fixture.store.AssertOutboxCatalog(
		ctx, empty,
	); !errors.Is(err, ErrOutboxCatalogIncompatible) {
		t.Fatalf("empty live outbox catalog error = %v", err)
	}
}
