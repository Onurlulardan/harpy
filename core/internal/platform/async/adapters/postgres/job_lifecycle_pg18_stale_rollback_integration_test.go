//go:build integration

package postgres

import (
	"context"
	"crypto/sha256"
	"fmt"
	"testing"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/google/uuid"
)

func claimLegacyStaleJob(
	t *testing.T,
	ctx context.Context,
	fixture jobLifecyclePGFixture,
	tenantID uuid.UUID,
) (JobClaim, []byte, string) {
	t.Helper()
	jobID := uuid.New()
	payload := []byte(`{"contract_version":"stale-rollback-test.v1"}`)
	digest := fmt.Sprintf("%x", sha256.Sum256(payload))
	_, err := fixture.runtime.Exec(ctx, `
INSERT INTO platform.jobs (
    id,tenant_id,scope_kind,scope_id,authority_group_id,
    authority_group_generation,authority_domain,cutover_generation,writer_origin,
    kind,payload_revision,payload_ref,payload_sha256,dedupe_key,effect_class,
    priority,run_at,state,attempts,max_attempts,lease_generation
) VALUES (
    $1,$2,'tenant',$2,'operations',1,'workflow',1,'v2',
    'workflow.execute',1,$3,$4,$5,'database_transactional',100,
    pg_catalog.clock_timestamp()-interval '1 second','queued',0,8,0
)`, jobID, tenantID, "payload/"+jobID.String(), digest, "dedupe:"+jobID.String())
	if err != nil {
		t.Fatalf("insert legacy stale job: %v", err)
	}
	claims, err := fixture.store.ClaimJobs(
		ctx, fixture.authority, uuid.New(), 1,
		lifecycleCatalog(t, jobsv1.EffectDatabaseTransactional, 8),
	)
	if err != nil || len(claims) != 1 || claims[0].JobID() != jobID {
		t.Fatalf("claim legacy stale job = %#v, %v", claims, err)
	}
	return claims[0], payload, digest
}

func assertLegacyStaleJobPark(
	t *testing.T,
	ctx context.Context,
	fixture jobLifecyclePGFixture,
	claim JobClaim,
	payload []byte,
	payloadDigest string,
) {
	t.Helper()
	errorDigest := fmt.Sprintf("%x", sha256.Sum256([]byte("JOB_AUTHORITY_STALE")))
	var parked bool
	err := fixture.runtime.QueryRow(ctx, `
SELECT platform.platform_async_park_stale_work_v1(
    $1,$2,$3,$4,$5,$6,$7::jsonb,$8
)`, claim.JobID(), claim.LeaseOwnerID(), claim.LeaseGeneration(), uuid.New(),
		payload, payloadDigest, `{"contract_version":"stale-rollback-test.v1"}`,
		errorDigest,
	).Scan(&parked)
	if err != nil || !parked {
		t.Fatalf("legacy stale park = %t, %v", parked, err)
	}
	var roots int
	err = fixture.owner.QueryRowContext(ctx, `
SELECT count(*) FROM platform.dead_letters
WHERE source_kind='job' AND source_id=$1
  AND reason_code='JOB_AUTHORITY_STALE'`, claim.JobID()).Scan(&roots)
	if err != nil || roots != 1 {
		t.Fatalf("legacy stale evidence = %d, %v", roots, err)
	}
}
