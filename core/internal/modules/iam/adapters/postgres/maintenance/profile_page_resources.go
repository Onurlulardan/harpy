package maintenance

import (
	"context"
	"crypto/sha256"
	"fmt"

	"github.com/google/uuid"
)

func (store *store) readProfilePageResources(
	ctx context.Context, values []any, page profilePage,
) ([]profileResource, error) {
	rows, err := store.database.Query(
		ctx, readProfileResourcesSQL,
		append(values, page.planID, page.pageNumber)...,
	)
	if err != nil {
		return nil, fmt.Errorf("read profile resources: %w", err)
	}
	defer rows.Close()
	result := make([]profileResource, 0, page.expectedCount)
	wipe := func() {
		for index := range result {
			result[index].wipe()
		}
	}
	for rows.Next() {
		var resource profileResource
		if err := rows.Scan(
			&resource.id, &resource.actualSchemaID, &resource.revision,
			&resource.document, &resource.sha, &resource.preconverted,
			&resource.plannedTarget,
		); err != nil || resource.id == uuid.Nil ||
			resource.actualSchemaID == uuid.Nil || resource.revision < 1 ||
			len(resource.sha) != sha256.Size || len(result) == 250 ||
			(resource.preconverted && (len(resource.document) != 0 ||
				len(resource.plannedTarget) != sha256.Size)) ||
			(!resource.preconverted && (len(resource.document) == 0 ||
				len(resource.plannedTarget) != 0)) {
			resource.wipe()
			wipe()
			return nil, ErrRuntimeInvalid
		}
		result = append(result, resource)
	}
	if err := rows.Err(); err != nil {
		wipe()
		return nil, fmt.Errorf("iterate profile resources: %w", err)
	}
	return result, nil
}
