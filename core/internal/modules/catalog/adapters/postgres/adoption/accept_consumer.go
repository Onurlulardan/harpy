package adoption

import (
	"context"
	_ "embed"
	"time"

	adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/adoption"
)

//go:embed sql/tenant/acknowledge.sql
var acknowledgeSQL string

func (storage *store) AcceptConsumer(
	ctx context.Context, value adoptionapp.ConsumerAck, occurredAt time.Time,
) error {
	if value.RouteKey == "" || value.ConsumerID == "" ||
		!digestPattern.MatchString(value.EvidenceSHA256) || !canonicalTime(occurredAt) {
		return adoptionapp.ErrInvalid
	}
	if !storage.usable(ctx, true) {
		return unavailable()
	}
	var route, consumer, digest string
	if err := storage.queryRow(ctx, acknowledgeSQL, value.RouteKey, value.ConsumerID,
		value.EvidenceSHA256, occurredAt).Scan(&route, &consumer, &digest); err != nil {
		if storage.notFound(err) {
			return adoptionapp.ErrConflict
		}
		return unavailable()
	}
	if route != value.RouteKey || consumer != value.ConsumerID || digest != value.EvidenceSHA256 {
		return adoptionapp.ErrConflict
	}
	return nil
}
