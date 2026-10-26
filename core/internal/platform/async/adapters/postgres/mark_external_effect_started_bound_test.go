package postgres

import (
	"context"
	"slices"
	"testing"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
)

func TestMarkExternalEffectStartedMintsPermitInsideBoundTransaction(t *testing.T) {
	t.Parallel()
	worker, claim := validLifecycleClaim(t, jobsv1.EffectExternal)
	startedAt := time.Now().UTC().Add(time.Millisecond)
	database := &boundLifecycleDatabase{claim: claim, startedAt: startedAt}
	repository := &repository{
		database: database,
		authority: authorityStamp{
			scopeID: claim.tenantID, groupID: claim.authority.groupID,
			groupGeneration: claim.authority.groupGeneration,
			domain:          claim.authority.domain, generation: claim.authority.generation,
		},
	}
	permit, err := repository.MarkExternalEffectStarted(
		context.Background(), claim, validExternalSpec(t),
	)
	if err != nil || !permit.valid(worker.identity) ||
		!permit.StartedAt().Equal(startedAt) {
		t.Fatalf("bound effect permit = %#v, %v", permit, err)
	}
	if !slices.Equal(database.statements, []string{
		lockJobClaimSQL, markExternalJobEffectSQL, markExternalAttemptEffectSQL,
	}) {
		t.Fatalf("bound effect statements = %d", len(database.statements))
	}
}
