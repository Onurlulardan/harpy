package adoption

import (
	"context"
	_ "embed"
	"time"

	adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/adoption"
)

//go:embed sql/tenant/record_hit.sql
var recordHitSQL string

func (storage *store) RecordHit(ctx context.Context, request adoptionapp.RouteRequest,
	occurredAt time.Time,
) error {
	if request.RouteKey == "" || request.RequiredConsumers < 0 ||
		!canonicalTime(request.RemoveAfter) || !canonicalTime(occurredAt) ||
		!request.RemoveAfter.After(occurredAt) {
		return adoptionapp.ErrInvalid
	}
	if !storage.usable(ctx, true) {
		return unavailable()
	}
	var route string
	if err := storage.queryRow(ctx, recordHitSQL, request.RouteKey,
		request.RequiredConsumers, request.RemoveAfter, occurredAt).Scan(&route); err != nil ||
		route != request.RouteKey {
		return unavailable()
	}
	return nil
}
