package providerapp

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestProviderFinalizeJobSealsDatabaseEffectIdentity(t *testing.T) {
	commandID := uuid.MustParse("018f1000-0000-7000-8000-000000000001")
	runAt := time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC)
	job, err := NewProviderFinalizeJob(
		commandID,
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		runAt,
	)
	if err != nil {
		t.Fatal(err)
	}
	identity := commandID.String()
	if job.Policy() != ProviderFinalizePolicy() ||
		job.PayloadReference() != "billing-provider-finalize/"+identity ||
		job.DedupeKey() != "billing:provider:finalize:"+identity ||
		job.EffectKey() != "" || !job.RunAt().Equal(runAt) {
		t.Fatalf("provider finalize job = %#v", job)
	}
}
