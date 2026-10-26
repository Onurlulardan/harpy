package adoption

import (
	"context"
	_ "embed"
	"time"

	adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/adoption"
)

//go:embed sql/tenant/retire.sql
var retireSQL string

func (storage *store) Retire(ctx context.Context, routeKey string, occurredAt time.Time) error {
	if routeKey == "" || !canonicalTime(occurredAt) {
		return adoptionapp.ErrInvalid
	}
	if !storage.usable(ctx, true) {
		return unavailable()
	}
	var route string
	if err := storage.queryRow(ctx, retireSQL, routeKey, occurredAt).Scan(&route); err != nil {
		if storage.notFound(err) {
			return adoptionapp.ErrBlocked
		}
		return unavailable()
	}
	if route != routeKey {
		return unavailable()
	}
	return nil
}
