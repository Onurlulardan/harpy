package ports

import (
	"context"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

type StaleDeliveryReconciler interface {
	SweepStaleDelivery(context.Context, domain.ID, domain.ID, time.Time) error
}
