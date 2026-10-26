package notifications

import (
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/application"
	"github.com/google/uuid"
)

func domainIDFromUUID(value uuid.UUID) (application.ID, error) {
	return application.IDFromUUID(value)
}
