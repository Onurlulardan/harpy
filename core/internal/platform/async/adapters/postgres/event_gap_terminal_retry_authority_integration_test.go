//go:build integration

package postgres

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func makeGapProducerGroupStale(
	t *testing.T,
	ctx context.Context,
	value terminalGapRetryCase,
) (string, int64) {
	t.Helper()
	var jobGroup, producerGroup string
	var generation int64
	err := value.fixture.runtime.QueryRow(ctx, `
SELECT job.authority_group_id,event.authority_group_id,
       event.authority_group_generation
FROM platform.jobs AS job
JOIN platform.outbox AS event ON event.event_id=$2
WHERE job.id=$1`, value.jobID, value.sourceEventID).Scan(
		&jobGroup, &producerGroup, &generation,
	)
	if err != nil {
		t.Fatalf("read gap producer authority: %v", err)
	}
	if jobGroup == producerGroup {
		t.Fatalf("producer-only switch shares consumer group %q", jobGroup)
	}
	setGapProducerGroupGeneration(
		t, ctx, value.fixture, producerGroup, generation+1,
	)
	return producerGroup, generation
}

func setGapProducerGroupGeneration(
	t *testing.T,
	ctx context.Context,
	fixture *eventGapPGFixture,
	groupID string,
	generation int64,
) {
	t.Helper()
	result, err := fixture.owner.ExecContext(ctx, `
UPDATE platform.cutover_groups
SET generation=$1
WHERE scope_kind='tenant' AND scope_id=$2 AND group_id=$3
  AND state='v2' AND writer='v2'`,
		generation, uuid.MustParse(asyncTenantA), groupID,
	)
	if err != nil {
		t.Fatalf("set gap producer group generation: %v", err)
	}
	if affected, err := result.RowsAffected(); err != nil || affected != 1 {
		t.Fatalf("gap producer generation affected %d rows: %v", affected, err)
	}
}
