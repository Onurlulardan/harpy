package maintenance

import (
	"context"
	"fmt"
)

func (store *store) readDirectoryRepairPage(
	ctx context.Context, values []any,
) ([]directorySourceRow, error) {
	rows, err := store.database.Query(ctx, readDirectoryPageSQL, values...)
	if err != nil {
		return nil, fmt.Errorf("read directory repair page: %w", err)
	}
	defer rows.Close()
	result := make([]directorySourceRow, 0, 250)
	wipe := func() {
		for index := range result {
			result[index].wipe()
		}
	}
	for rows.Next() {
		var source directorySourceRow
		if err := rows.Scan(
			&source.id, &source.userID, &source.kind, &source.normalized,
			&source.normalizedSHA, &source.state, &source.sourceRevision,
		); err != nil {
			wipe()
			return nil, fmt.Errorf("scan directory repair source: %w", err)
		}
		if !source.valid() || len(result) == 250 {
			source.wipe()
			wipe()
			return nil, ErrRuntimeInvalid
		}
		result = append(result, source)
	}
	if err := rows.Err(); err != nil {
		wipe()
		return nil, fmt.Errorf("iterate directory repair page: %w", err)
	}
	return result, nil
}
