package ports

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

type Clock interface{ Now() time.Time }

type IDSource interface {
	New(time.Time) (domain.ID, error)
}
