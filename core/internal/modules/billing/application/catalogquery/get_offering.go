package catalogquery

import (
	"context"
	"fmt"
	"time"

	offeringdomain "github.com/Onurlulardan/harpy/core/internal/modules/billing/domain/offering"
)

func BuildOfferingRequest(key string, instant time.Time) (GetRequest, error) {
	value, err := offeringdomain.NewKey(key)
	if err != nil || !validInstant(instant) {
		return GetRequest{}, ErrRequest
	}
	return GetRequest{key: string(value), instant: instant}, nil
}

func (queries *Queries) GetOffering(
	ctx context.Context,
	request GetRequest,
) (OfferingSnapshot, error) {
	key, err := offeringdomain.NewKey(request.key)
	if !queries.usable(ctx) || err != nil || string(key) != request.key ||
		!validInstant(request.instant) {
		return OfferingSnapshot{}, ErrRequest
	}
	record, found, err := queries.store.GetOffering(ctx, GetStoreQuery{
		Key: request.key, Instant: request.instant,
	})
	if err != nil {
		return OfferingSnapshot{}, fmt.Errorf("get public billing offering: %w", err)
	}
	if !found {
		return OfferingSnapshot{}, ErrNotFound
	}
	if record.Key != request.key || !record.Available || !record.Public ||
		record.State != "active" || !activeAt(
		record.EffectiveStart, record.EffectiveEnd, request.instant,
	) {
		return OfferingSnapshot{}, ErrResult
	}
	result, valid := projectOffering(record, true)
	if !valid {
		return OfferingSnapshot{}, ErrResult
	}
	return result, nil
}
