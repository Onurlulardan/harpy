package postgres

import (
	"strings"
	"testing"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/google/uuid"
)

func TestDatabaseJobChildInheritsClaimAuthorityAndCatalogPolicy(t *testing.T) {
	_, claim := validLifecycleClaim(t, jobsv1.EffectDatabaseTransactional)
	policy := childJobPolicy(t, "forms.reconcile", "forms", jobsv1.EffectExternal)
	catalog := childJobCatalog(t, policy)
	child := validDatabaseJobChild(claim, policy)
	value, resolved, err := newDatabaseJobChildEnqueue(claim, child, catalog)
	if err != nil {
		t.Fatal(err)
	}
	tenantID, tenantScoped := claim.TenantID()
	if !tenantScoped || value.TenantID == nil || *value.TenantID != tenantID.String() ||
		value.Authority.ScopeID != tenantID.String() ||
		value.Authority.GroupID != claim.authority.groupID ||
		value.Authority.GroupGeneration != claim.authority.groupGeneration ||
		value.Authority.Domain != claim.authority.domain ||
		value.Authority.CutoverGeneration != claim.authority.generation ||
		value.MaximumAttempts != policy.MaximumAttempts || resolved != policy ||
		value.EffectIdempotencyKey == nil ||
		*value.EffectIdempotencyKey != child.EffectIdempotencyKey {
		t.Fatalf("child enqueue did not inherit authority: %#v", value)
	}
}

func TestDatabaseJobChildFailsClosedBeforePersistence(t *testing.T) {
	_, claim := validLifecycleClaim(t, jobsv1.EffectDatabaseTransactional)
	external := childJobPolicy(t,
		"forms.reconcile", "forms", jobsv1.EffectExternal)
	database := childJobPolicy(t,
		"forms.followup", "forms", jobsv1.EffectDatabaseTransactional)
	sameKind := childJobPolicy(t,
		claim.Kind(), "forms", jobsv1.EffectExternal)
	tests := map[string]struct {
		policy jobsv1.KindPolicy
		mutate func(*platformasync.DatabaseJobChildInput)
	}{
		"same kind": {sameKind, func(*platformasync.DatabaseJobChildInput) {}},
		"same job id": {external, func(value *platformasync.DatabaseJobChildInput) {
			value.JobID = claim.JobID()
		}},
		"stale run at": {external, func(value *platformasync.DatabaseJobChildInput) {
			value.RunAt = claim.StartedAt().Truncate(time.Microsecond).Add(-time.Microsecond)
		}},
		"missing external key": {external, func(value *platformasync.DatabaseJobChildInput) {
			value.EffectIdempotencyKey = ""
		}},
		"database effect key": {database, func(value *platformasync.DatabaseJobChildInput) {
			value.EffectIdempotencyKey = "effect:must-not-exist"
		}},
		"wrong owner": {
			childJobPolicy(t, "forms.reconcile", "billing", jobsv1.EffectExternal),
			func(*platformasync.DatabaseJobChildInput) {},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			child := validDatabaseJobChild(claim, test.policy)
			test.mutate(&child)
			if _, _, err := newDatabaseJobChildEnqueue(
				claim, child, childJobCatalog(t, test.policy),
			); err == nil {
				t.Fatal("invalid child enqueue was accepted")
			}
		})
	}
	unknown := validDatabaseJobChild(claim, external)
	if _, _, err := newDatabaseJobChildEnqueue(
		claim, unknown, childJobCatalog(t),
	); err == nil {
		t.Fatal("unregistered child policy was accepted")
	}
}

func childJobPolicy(
	t *testing.T,
	kind string,
	owner string,
	effect jobsv1.EffectClass,
) jobsv1.KindPolicy {
	t.Helper()
	policy, err := jobsv1.NewKindPolicy(kind, 1, owner, effect, 6)
	if err != nil {
		t.Fatal(err)
	}
	return policy
}

func childJobCatalog(t *testing.T, policies ...jobsv1.KindPolicy) jobsv1.Catalog {
	t.Helper()
	catalog, err := jobsv1.NewCatalog(policies)
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func validDatabaseJobChild(
	claim JobClaim,
	policy jobsv1.KindPolicy,
) platformasync.DatabaseJobChildInput {
	return platformasync.DatabaseJobChildInput{
		JobID: uuid.Must(uuid.NewV7()), Kind: policy.Kind,
		PayloadRevision:  policy.PayloadRevision,
		RunAt:            claim.StartedAt().Truncate(time.Microsecond).Add(time.Microsecond),
		PayloadReference: "workflow-child/" + claim.JobID().String(),
		PayloadSHA256:    strings.Repeat("d", 64), DedupeKey: "workflow:child:1",
		EffectIdempotencyKey: "workflow:effect:1", Priority: 100,
	}
}
