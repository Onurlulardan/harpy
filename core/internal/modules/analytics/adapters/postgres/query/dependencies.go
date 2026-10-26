package query

import (
	"context"

	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
	"github.com/google/uuid"
)

func (storage *store) Dependencies(
	ctx context.Context, queryID, versionID uuid.UUID,
) ([]queryapp.DependencySnapshot, error) {
	if !storage.usable(ctx) || !validUUIDv7(queryID) || !validUUIDv7(versionID) {
		return nil, queryapp.ErrInvalid
	}
	rows, err := storage.query(ctx, dependenciesSQL, queryID, versionID)
	if err != nil {
		return nil, mapError(err, storage.notFound)
	}
	defer rows.Close()
	result := make([]queryapp.DependencySnapshot, 0, 4)
	for rows.Next() {
		value, scanErr := scanDependency(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		result = append(result, value)
	}
	if rows.Err() != nil || len(result) == 0 || len(result) > 32 {
		return nil, queryapp.ErrUnavailable
	}
	return result, nil
}
