package analytics

import (
	"context"
	"testing"
	"time"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	retentionapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/retention"
	"github.com/google/uuid"
)

type analyticsRetentionStoreFixture struct {
	outcome retentionapp.Outcome
}

func (store *analyticsRetentionStoreFixture) Execute(
	context.Context, retentionapp.Claim, uuid.UUID,
) (retentionapp.Outcome, error) {
	return store.outcome, nil
}

func TestAnalyticsRetentionCommandPersistsSuccessor(t *testing.T) {
	nextID := uuid.MustParse("018f0000-0000-7000-8000-000000000506")
	reference, digest, err := retentionapp.Reference(analyticsJobTenant)
	if err != nil {
		t.Fatal(err)
	}
	store := &analyticsRetentionStoreFixture{outcome: retentionapp.Outcome{
		CheckpointRevision: 1,
		Successor: retentionapp.Successor{JobID: nextID,
			RunAt:            time.Date(2026, 7, 20, 3, 0, 0, 0, time.UTC),
			PayloadReference: reference, PayloadSHA256: digest,
			DedupeKey: "analytics:retention:" + nextID.String()},
	}}
	commands, err := retentionapp.NewCommands(store,
		func() (uuid.UUID, error) { return nextID, nil })
	if err != nil {
		t.Fatal(err)
	}
	command := &analyticsRetentionCommand{commands: commands}
	jobOutcome, err := command.Execute(t.Context(), newAnalyticsJobClaim(
		asyncmanifestv1.AnalyticsRetentionJobKind, reference))
	if err != nil || !jobOutcome.Valid() {
		t.Fatalf("retention command outcome valid=%v err=%v", jobOutcome.Valid(), err)
	}
	successor, found := jobOutcome.Successor()
	if !found || successor.JobID != nextID || successor.PayloadSHA256 != digest {
		t.Fatalf("retention successor=%+v found=%v", successor, found)
	}
}
