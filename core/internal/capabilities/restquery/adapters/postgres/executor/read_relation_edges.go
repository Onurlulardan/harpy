package executor

import (
	"context"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api/execution"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
	"github.com/google/uuid"
)

func readRelationEdges(
	ctx context.Context,
	driver execution.SnapshotDriver,
	policy planner.Policy,
	candidate statement,
	sourceCount int,
) ([]relationEdge, error) {
	if ctx == nil || interfaceIsNil(driver) ||
		candidate.class != planner.StatementClassInclude ||
		sourceCount < 1 || sourceCount > 1000 {
		return nil, ErrContract
	}
	if err := applyTimeout(ctx, driver, policy, candidate.class); err != nil {
		return nil, err
	}
	rows, err := driver.Query(ctx, candidate.sql, candidate.arguments...)
	if err != nil {
		return nil, mapDatabaseError(ctx, policy, candidate.class, err)
	}
	if interfaceIsNil(rows) {
		return nil, ErrContract
	}
	defer rows.Close()
	maximum := sourceCount * 100
	edges := make([]relationEdge, 0, min(maximum, 1024))
	for rows.Next() {
		if len(edges) == maximum {
			return nil, ErrContract
		}
		var source, target uuid.UUID
		var slot int16
		if err := rows.Scan(&source, &target, &slot); err != nil {
			return nil, mapScanError(ctx, policy, candidate.class, err)
		}
		edge := relationEdge{source: source, target: target, slot: slot}
		if edge.validate() != nil {
			return nil, ErrContract
		}
		edges = append(edges, edge)
	}
	if err := rows.Err(); err != nil {
		return nil, mapDatabaseError(ctx, policy, candidate.class, err)
	}
	return edges, nil
}
