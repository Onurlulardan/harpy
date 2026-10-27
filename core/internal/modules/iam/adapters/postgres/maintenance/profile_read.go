package maintenance

import (
	"context"
	"fmt"

	maintenanceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/maintenance"
	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/google/uuid"
)

func (store *store) readProfilePage(
	ctx context.Context,
	claim maintenanceapp.Claim,
) (profilePage, error) {
	values, err := claimValues(claim)
	if err != nil {
		return profilePage{}, err
	}
	var page profilePage
	var resource, converter, state string
	var after *uuid.UUID
	err = store.database.QueryRow(ctx, readProfilePageSQL, values...).Scan(
		&page.planID, &resource, &page.sourceSchemaID, &page.targetSchemaID,
		&converter, &state, &page.pageNumber, &after, &page.throughID,
		&page.expectedCount, &page.sourceSHA, &page.targetSHA,
	)
	if err != nil {
		return profilePage{}, fmt.Errorf("read profile backfill page: %w", err)
	}
	page.resource = profileapp.ResourceKind(resource)
	if after != nil {
		value := *after
		page.afterID = &value
	}
	if converter != "typed_map_v1" ||
		(state != "planned" && state != "running") || !page.valid() {
		return profilePage{}, ErrRuntimeInvalid
	}
	return page, nil
}

func (store *store) readProfileMappings(
	ctx context.Context,
	claim maintenanceapp.Claim,
	planID uuid.UUID,
) ([]profileMapping, error) {
	values, err := claimValues(claim)
	if err != nil {
		return nil, err
	}
	rows, err := store.database.Query(
		ctx, readProfileMappingsSQL, append(values, planID)...,
	)
	if err != nil {
		return nil, fmt.Errorf("read profile mappings: %w", err)
	}
	defer rows.Close()
	result := make([]profileMapping, 0)
	sources, targets := map[uuid.UUID]struct{}{}, map[uuid.UUID]struct{}{}
	for rows.Next() {
		var mapping profileMapping
		if err := rows.Scan(
			&mapping.sourceID, &mapping.targetID, &mapping.kind, &mapping.semanticSHA,
		); err != nil || !mapping.valid() {
			return nil, ErrRuntimeInvalid
		}
		if _, exists := sources[mapping.sourceID]; exists {
			return nil, ErrRuntimeInvalid
		}
		if _, exists := targets[mapping.targetID]; exists {
			return nil, ErrRuntimeInvalid
		}
		sources[mapping.sourceID], targets[mapping.targetID] = struct{}{}, struct{}{}
		result = append(result, mapping)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate profile mappings: %w", err)
	}
	if len(result) == 0 {
		return nil, ErrRuntimeInvalid
	}
	return result, nil
}
