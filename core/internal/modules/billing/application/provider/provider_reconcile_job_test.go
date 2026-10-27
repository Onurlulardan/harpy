package providerapp

import (
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestProviderCommandReconciliationJobUsesStableEffectIdentity(t *testing.T) {
	commandID := uuid.MustParse("018f1000-0000-7000-8000-000000000011")
	runAt := time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC)
	job, err := NewProviderCommandReconciliationJob(
		"manual", commandID, strings.Repeat("a", 64), runAt,
	)
	if err != nil {
		t.Fatal(err)
	}
	if job.Policy() != ProviderReconciliationPolicy() ||
		job.PayloadReference() != "billing-provider-reconciliation/manual/"+
			commandID.String() || job.DedupeKey() != job.EffectKey() ||
		!strings.HasPrefix(job.EffectKey(), "billing:provider:reconcile:") {
		t.Fatalf("reconciliation job = %#v", job)
	}
}
