package providerapp

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestProviderCommandJobSealsCanonicalExternalEffectIdentity(t *testing.T) {
	commandID := uuid.MustParse("018f1000-0000-7000-8000-000000000001")
	runAt := time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC)
	job, err := NewProviderCommandJob(
		"manual", commandID,
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		runAt,
	)
	if err != nil {
		t.Fatal(err)
	}
	identity := "billing:provider:command:manual:" + commandID.String()
	if job.Policy() != ProviderCommandPolicy() ||
		job.PayloadReference() != "billing-provider-command/manual/"+commandID.String() ||
		job.DedupeKey() != identity || job.EffectKey() != identity ||
		!job.RunAt().Equal(runAt) {
		t.Fatalf("provider command job = %#v", job)
	}
}

func TestProviderCommandJobRejectsNonCanonicalIdentity(t *testing.T) {
	_, err := NewProviderCommandJob(
		"Manual", uuid.New(),
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		time.Now().UTC().Truncate(time.Microsecond),
	)
	if err == nil {
		t.Fatal("invalid provider command job accepted")
	}
}
