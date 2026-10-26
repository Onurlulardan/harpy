package postgres

import (
	"context"
	"errors"
	"slices"
	"testing"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
)

func TestCompleteDatabaseJobUsesOnlyBoundTransaction(t *testing.T) {
	t.Parallel()
	_, claim := validLifecycleClaim(t, jobsv1.EffectDatabaseTransactional)
	completed := time.Now().UTC().Add(time.Second)
	database := &boundLifecycleDatabase{claim: claim, completed: completed}
	repository := &repository{
		database: database,
		authority: authorityStamp{
			scopeID: claim.tenantID, groupID: claim.authority.groupID,
			groupGeneration: claim.authority.groupGeneration,
			domain:          claim.authority.domain, generation: claim.authority.generation,
		},
	}
	persisted, err := repository.CompleteDatabaseJob(context.Background(), claim)
	if err != nil || !persisted.Equal(completed) {
		t.Fatalf("bound completion = %v, %v", persisted, err)
	}
	if !slices.Equal(database.statements, []string{
		lockJobClaimSQL, completeJobAttemptSQL, completeDatabaseJobSQL,
	}) {
		t.Fatalf("bound completion statements = %d", len(database.statements))
	}
}

func TestCompleteDatabaseJobRejectsCrossTenantBeforeSQL(t *testing.T) {
	t.Parallel()
	_, claim := validLifecycleClaim(t, jobsv1.EffectDatabaseTransactional)
	database := &boundLifecycleDatabase{claim: claim}
	repository := &repository{database: database, authority: authorityStamp{}}
	if _, err := repository.CompleteDatabaseJob(
		context.Background(), claim,
	); !errors.Is(err, ErrAuthorityMismatch) {
		t.Fatalf("cross-tenant completion error = %v", err)
	}
	if len(database.statements) != 0 {
		t.Fatal("cross-tenant completion reached SQL")
	}
}
