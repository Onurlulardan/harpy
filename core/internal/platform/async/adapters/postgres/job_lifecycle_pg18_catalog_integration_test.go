//go:build integration

package postgres

import (
	"context"
	"errors"
	"testing"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/google/uuid"
)

func assertJobCatalogClaimFence(
	t *testing.T,
	ctx context.Context,
	fixture jobLifecyclePGFixture,
) {
	t.Helper()
	base := lifecyclePolicy(
		t, "workflow.execute", 1, jobsv1.EffectDatabaseTransactional, 8,
	)
	supported := []jobsv1.KindPolicy{base}
	if err := fixture.store.AssertJobCatalog(
		ctx, lifecyclePolicies(t, supported),
	); err != nil {
		t.Fatalf("exact startup catalog: %v", err)
	}
	mismatches := make([]uuid.UUID, 0, 4)
	mismatches = append(mismatches, insertCatalogJob(
		t, ctx, fixture, "workflow.removed", 1,
		jobsv1.EffectDatabaseTransactional, 8,
	))
	assertCatalogIncompatible(t, ctx, fixture, lifecyclePolicies(t, supported))
	supported = append(supported, lifecyclePolicy(
		t, "workflow.removed", 1, jobsv1.EffectDatabaseTransactional, 8,
	))
	assertCatalogCompatible(t, ctx, fixture, supported)

	mismatches = append(mismatches, insertCatalogJob(
		t, ctx, fixture, "workflow.revisioned", 2,
		jobsv1.EffectDatabaseTransactional, 8,
	))
	expected := append(append([]jobsv1.KindPolicy(nil), supported...), lifecyclePolicy(
		t, "workflow.revisioned", 1, jobsv1.EffectDatabaseTransactional, 8,
	))
	assertCatalogIncompatible(t, ctx, fixture, lifecyclePolicies(t, expected))
	supported = append(supported, lifecyclePolicy(
		t, "workflow.revisioned", 2, jobsv1.EffectDatabaseTransactional, 8,
	))
	assertCatalogCompatible(t, ctx, fixture, supported)

	mismatches = append(mismatches, insertCatalogJob(
		t, ctx, fixture, "workflow.effect", 1, jobsv1.EffectExternal, 8,
	))
	expected = append(append([]jobsv1.KindPolicy(nil), supported...), lifecyclePolicy(
		t, "workflow.effect", 1, jobsv1.EffectDatabaseTransactional, 8,
	))
	assertCatalogIncompatible(t, ctx, fixture, lifecyclePolicies(t, expected))
	supported = append(supported, lifecyclePolicy(
		t, "workflow.effect", 1, jobsv1.EffectExternal, 8,
	))
	assertCatalogCompatible(t, ctx, fixture, supported)

	mismatches = append(mismatches, insertCatalogJob(
		t, ctx, fixture, "workflow.maximum", 1,
		jobsv1.EffectDatabaseTransactional, 7,
	))
	expected = append(append([]jobsv1.KindPolicy(nil), supported...), lifecyclePolicy(
		t, "workflow.maximum", 1, jobsv1.EffectDatabaseTransactional, 8,
	))
	assertCatalogIncompatible(t, ctx, fixture, lifecyclePolicies(t, expected))
	supported = append(supported, lifecyclePolicy(
		t, "workflow.maximum", 1, jobsv1.EffectDatabaseTransactional, 7,
	))
	assertCatalogCompatible(t, ctx, fixture, supported)

	exact := insertCatalogJob(
		t, ctx, fixture, "workflow.execute", 1,
		jobsv1.EffectDatabaseTransactional, 8,
	)
	claims, err := fixture.store.ClaimJobs(
		ctx, fixture.authority, uuid.New(), 1,
		lifecycleCatalog(t, jobsv1.EffectDatabaseTransactional, 8),
	)
	if err != nil || len(claims) != 1 || claims[0].JobID() != exact {
		t.Fatalf("catalog-fenced claims = %#v, %v", claims, err)
	}
	var untouched int
	if err := fixture.owner.QueryRowContext(ctx, `
SELECT count(*) FROM platform.jobs
WHERE id = ANY($1::uuid[]) AND state='queued' AND attempts=0`, mismatches,
	).Scan(&untouched); err != nil || untouched != len(mismatches) {
		t.Fatalf("unsupported jobs untouched=%d error=%v", untouched, err)
	}
}

func assertCatalogIncompatible(
	t *testing.T,
	ctx context.Context,
	fixture jobLifecyclePGFixture,
	catalog jobsv1.Catalog,
) {
	t.Helper()
	if err := fixture.store.AssertJobCatalog(
		ctx, catalog,
	); !errors.Is(err, ErrJobCatalogIncompatible) {
		t.Fatalf("incompatible startup catalog error = %v", err)
	}
}

func assertCatalogCompatible(
	t *testing.T,
	ctx context.Context,
	fixture jobLifecyclePGFixture,
	policies []jobsv1.KindPolicy,
) {
	t.Helper()
	if err := fixture.store.AssertJobCatalog(
		ctx, lifecyclePolicies(t, policies),
	); err != nil {
		t.Fatalf("compatible startup catalog: %v", err)
	}
}

func lifecyclePolicy(
	t *testing.T,
	kind string,
	revision int,
	effect jobsv1.EffectClass,
	maximumAttempts int,
) jobsv1.KindPolicy {
	t.Helper()
	policy, err := jobsv1.NewKindPolicy(
		kind, revision, "workflow", effect, maximumAttempts,
	)
	if err != nil {
		t.Fatal(err)
	}
	return policy
}

func lifecyclePolicies(
	t *testing.T,
	policies []jobsv1.KindPolicy,
) jobsv1.Catalog {
	t.Helper()
	catalog, err := jobsv1.NewCatalog(policies)
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func insertCatalogJob(
	t *testing.T,
	ctx context.Context,
	fixture jobLifecyclePGFixture,
	kind string,
	revision int,
	effect jobsv1.EffectClass,
	maximumAttempts int,
) uuid.UUID {
	t.Helper()
	jobID := uuid.New()
	var effectKey any
	if effect == jobsv1.EffectExternal {
		effectKey = "effect:catalog:" + jobID.String()
	}
	const statement = `
INSERT INTO platform.jobs (
 id,tenant_id,scope_kind,scope_id,authority_group_id,authority_group_generation,
 authority_domain,cutover_generation,writer_origin,kind,payload_revision,payload_ref,
 payload_sha256,dedupe_key,effect_class,effect_idempotency_key,priority,run_at,state,
 attempts,max_attempts,lease_generation
) VALUES (
 $1,$2,'tenant',$2,'operations',1,'workflow',1,'v2',$3,$4,$5,repeat('a',64),
 $6,$7,$8,100,pg_catalog.clock_timestamp()-interval '1 second','queued',0,$9,0
)`
	_, err := fixture.runtime.Exec(
		ctx, statement, jobID, asyncTenantA, kind, revision,
		"payload/"+jobID.String(), "dedupe:"+jobID.String(),
		string(effect), effectKey, maximumAttempts,
	)
	if err != nil {
		t.Fatalf("insert catalog job: %v", err)
	}
	return jobID
}
