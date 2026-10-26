package postgres

import (
	"context"
	"errors"
	"testing"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	"github.com/google/uuid"
)

func TestClaimOutboxRejectsForgedInputsBeforeTransaction(t *testing.T) {
	t.Parallel()
	database := &workerTestDatabase{}
	repository := newWorkerTestRepository()
	repository.database = database
	authority := validOutboxAuthority(t, repository)
	catalog := validPublisherCatalog(t)
	tests := []struct {
		name      string
		authority OutboxAuthority
		publisher uuid.UUID
		limit     int
		catalog   eventsv1.ProducerCatalog
		want      error
	}{
		{name: "authority", publisher: uuid.New(), limit: 1, catalog: catalog, want: ErrOutboxAuthorityInvalid},
		{name: "publisher", authority: authority, limit: 1, catalog: catalog, want: ErrPublishClaimInvalid},
		{name: "zero limit", authority: authority, publisher: uuid.New(), catalog: catalog, want: ErrPublishClaimLimit},
		{name: "large limit", authority: authority, publisher: uuid.New(), limit: 51, catalog: catalog, want: ErrPublishClaimLimit},
		{name: "empty catalog", authority: authority, publisher: uuid.New(), limit: 1, want: ErrPublishClaimInvalid},
	}
	for _, test := range tests {
		_, err := repository.ClaimOutbox(
			context.Background(), test.authority, test.publisher,
			test.limit, test.catalog,
		)
		if !errors.Is(err, test.want) {
			t.Fatalf("%s error = %v", test.name, err)
		}
	}
	if database.beginCalls != 0 {
		t.Fatalf("invalid claims began %d transactions", database.beginCalls)
	}
}

func TestClaimOutboxAcceptsExactBatchBoundaries(t *testing.T) {
	t.Parallel()
	database := &workerTestDatabase{}
	repository := newWorkerTestRepository()
	repository.database = database
	authority := validOutboxAuthority(t, repository)
	for _, limit := range []int{1, 50} {
		_, err := repository.ClaimOutbox(
			context.Background(), authority, uuid.New(), limit,
			validPublisherCatalog(t),
		)
		if !errors.Is(err, errWorkerTestDatabase) {
			t.Fatalf("limit %d error = %v", limit, err)
		}
	}
	if database.beginCalls != 2 {
		t.Fatalf("valid claims began %d transactions", database.beginCalls)
	}
}
