package usage

import (
	"context"
	"testing"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/google/uuid"
)

type retentionIDs struct{ value uuid.UUID }

func (ids retentionIDs) NewV7() (uuid.UUID, error) { return ids.value, nil }

type retentionFacadeRow struct{ deleted int }

func (row retentionFacadeRow) Scan(destinations ...any) error {
	*destinations[0].(*int) = row.deleted
	return nil
}

type retentionFacadeDatabase struct{ deleted int }

func (database retentionFacadeDatabase) QueryRow(
	context.Context,
	string,
	...any,
) PostgresRow {
	return retentionFacadeRow{deleted: database.deleted}
}

func TestRetentionPolicyContributionAndOwnerBinding(t *testing.T) {
	policy := RetentionJobPolicy()
	if err := ValidateRetentionJobPolicy(policy); err != nil ||
		policy.Kind != RetentionJobKind || policy.OwnerDomain != RetentionAuthorityDomain ||
		policy.EffectClass != jobsv1.EffectDatabaseTransactional ||
		len(RetentionJobPolicies()) != 1 {
		t.Fatalf("usage retention policy=%+v err=%v", policy, err)
	}
	drift := policy
	drift.MaximumAttempts = 3
	if err := ValidateRetentionJobPolicy(drift); err == nil {
		t.Fatal("usage retention policy drift was accepted")
	}
	nextID := uuid.MustParse("019c1000-0000-7000-8000-000000000003")
	binding, err := NewRetentionJobBinding(retentionIDs{nextID})
	if err != nil || binding.Owner() != RetentionAuthorityDomain {
		t.Fatalf("usage retention binding owner=%q err=%v", binding.Owner(), err)
	}
	commands, err := binding.Bind(retentionFacadeDatabase{})
	if err != nil || commands == nil {
		t.Fatalf("bind usage retention command=%v err=%v", commands, err)
	}
}

func TestRetentionBindingRejectsMissingDependencies(t *testing.T) {
	if _, err := NewRetentionJobBinding(nil); err == nil {
		t.Fatal("nil retention ID generator was accepted")
	}
	binding, _ := NewRetentionJobBinding(retentionIDs{
		uuid.MustParse("019c1000-0000-7000-8000-000000000003"),
	})
	var database *retentionFacadeDatabase
	if _, err := binding.Bind(database); err == nil {
		t.Fatal("nil retention database was accepted")
	}
}

func TestRetentionFacadeExecutesGenerationFencedClaim(t *testing.T) {
	environmentID := uuid.MustParse("30000000-0000-4000-8000-000000000003")
	cycleAt := time.Date(2026, 8, 2, 2, 0, 0, 0, time.UTC)
	authority, err := NewRetentionAuthority(RetentionAuthorityInput{
		EnvironmentID: environmentID, ScopeKind: RetentionScopeKind,
		ScopeID: environmentID, GroupID: RetentionAuthorityGroup,
		GroupGeneration: 3, Domain: RetentionAuthorityDomain,
		DomainGeneration: 5, WriterOrigin: RetentionWriterOrigin,
	})
	if err != nil {
		t.Fatal(err)
	}
	reference, digest, _ := RetentionReference(environmentID, cycleAt)
	binding, _ := NewRetentionJobBinding(retentionIDs{
		uuid.MustParse("019c1000-0000-7000-8000-000000000003"),
	})
	commands, _ := binding.Bind(retentionFacadeDatabase{deleted: 4})
	outcome, err := commands.Execute(context.Background(), RetentionJobClaim{
		JobID:     uuid.MustParse("019c1000-0000-7000-8000-000000000001"),
		AttemptID: uuid.MustParse("019c1000-0000-7000-8000-000000000002"),
		Authority: authority, Kind: RetentionJobKind,
		PayloadRevision: RetentionPayloadRevision, PayloadReference: reference,
		PayloadSHA256: digest, LeaseGeneration: 1, StartedAt: cycleAt,
	})
	if err != nil || outcome.DeletedCount != 4 || outcome.HasBacklog {
		t.Fatalf("retention facade outcome=%+v err=%v", outcome, err)
	}
}
