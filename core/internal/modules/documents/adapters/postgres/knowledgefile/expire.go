package knowledgefile

import (
	"context"
	"time"

	knowledgeapi "github.com/Onurlulardan/harpy/core/internal/modules/documents/api/knowledgefile"
	knowledgeapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/knowledgefile"
)

func (storage *store) Expire(ctx context.Context, now time.Time,
	limit uint16,
) ([]knowledgeapp.Lease, error) {
	if !storage.usable(ctx) || now.IsZero() || limit < 1 || limit > 100 {
		return nil, knowledgeapi.ErrInvalid
	}
	values, err := storage.query(ctx, expireSQL, now.UTC(), limit)
	if err != nil {
		return nil, knowledgeapi.ErrUnavailable
	}
	defer values.Close()
	result := make([]knowledgeapp.Lease, 0, limit)
	for values.Next() {
		value, scanErr := scanExpiredLease(values, storage.tenantID)
		if scanErr != nil {
			return nil, knowledgeapi.ErrUnavailable
		}
		result = append(result, value)
	}
	if values.Err() != nil {
		return nil, knowledgeapi.ErrUnavailable
	}
	return result, nil
}
