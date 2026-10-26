//go:build integration

package postgres

import (
	"context"
	"fmt"
)

func queryRetentionCycle(
	ctx context.Context, querier retentionCycleQuerier,
) (bool, [7]int32, error) {
	var acquired bool
	var counts [7]int32
	err := querier.QueryRow(ctx, retentionCycleCall).Scan(
		&acquired, &counts[0], &counts[1], &counts[2], &counts[3],
		&counts[4], &counts[5], &counts[6],
	)
	return acquired, counts, err
}

func runRetentionCycleCounts(
	ctx context.Context, querier retentionCycleQuerier,
) ([7]int32, error) {
	acquired, counts, err := queryRetentionCycle(ctx, querier)
	if err != nil {
		return counts, err
	}
	if !acquired {
		return counts, fmt.Errorf("retention cycle lock was not acquired")
	}
	return counts, nil
}
