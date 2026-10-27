package catalogquery

import (
	"context"
	"fmt"
	"time"

	offeringdomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/offering"
)

func BuildOfferingListRequest(limit int, cursor string, instant time.Time) (ListRequest, error) {
	after, err := decodeCursor(cursor)
	if err != nil || limit < 1 || limit > MaximumPageSize || !validInstant(instant) {
		return ListRequest{}, ErrRequest
	}
	return ListRequest{limit: limit, after: after, instant: instant}, nil
}

func (queries *Queries) ListOfferings(
	ctx context.Context,
	request ListRequest,
) (OfferingPage, error) {
	if !queries.usable(ctx) || !validListRequest(request) {
		return OfferingPage{}, ErrRequest
	}
	records, err := queries.store.ListOfferings(ctx, ListStoreQuery{
		AfterKey: request.after, Limit: request.limit + 1, Instant: request.instant,
	})
	if err != nil {
		return OfferingPage{}, fmt.Errorf("list public billing offerings: %w", err)
	}
	if len(records) > request.limit+1 {
		return OfferingPage{}, ErrResult
	}
	items := make([]OfferingSnapshot, len(records))
	previous := request.after
	for index, record := range records {
		if record.Key <= previous || !record.Available || !record.Public ||
			record.State != "active" || !activeAt(
			record.EffectiveStart, record.EffectiveEnd, request.instant,
		) {
			return OfferingPage{}, ErrResult
		}
		item, valid := projectOffering(record, true)
		if !valid {
			return OfferingPage{}, ErrResult
		}
		items[index], previous = item, record.Key
	}
	next := ""
	if len(items) > request.limit {
		items = items[:request.limit]
		next = encodeCursor(items[len(items)-1].Key())
	}
	return OfferingPage{items: items, nextCursor: next}, nil
}

func validListRequest(request ListRequest) bool {
	if request.limit < 1 || request.limit > MaximumPageSize ||
		!validInstant(request.instant) {
		return false
	}
	if request.after == "" {
		return true
	}
	key, err := offeringdomain.NewKey(request.after)
	return err == nil && string(key) == request.after
}
