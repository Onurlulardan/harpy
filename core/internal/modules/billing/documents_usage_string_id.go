package billing

import (
	"errors"

	usageapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/usage"
	"github.com/google/uuid"
)

func documentsUsageStringV7(value string) (uuid.UUID, error) {
	parsed, err := uuid.Parse(value)
	if err != nil || !documentsUsageV7(parsed) || parsed.String() != value {
		return uuid.Nil, usageapp.ErrRequest
	}
	return parsed, nil
}

func documentsUsageAllowed(err error) (bool, error) {
	if errors.Is(err, ErrDocumentsUsageDenied) {
		return false, nil
	}
	return err == nil, err
}
