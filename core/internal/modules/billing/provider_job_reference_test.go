package billing

import (
	"testing"
	"time"

	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
	"github.com/google/uuid"
)

func TestProviderJobReferenceRequiresCanonicalProviderAndIdentifier(t *testing.T) {
	provider, identifier, err := parseProviderJobReference(
		"billing-provider-command/manual/018f1000-0000-7000-8000-000000000001",
		"billing-provider-command",
	)
	if err != nil || provider != "manual" || identifier.Version() != 7 {
		t.Fatalf("provider reference = %s/%s, %v", provider, identifier, err)
	}
	invalid := []string{
		"billing-provider-command/manual",
		"billing-provider-command/Manual/018f1000-0000-7000-8000-000000000001",
		"billing-provider-command/manual/not-a-uuid",
		"other/manual/018f1000-0000-7000-8000-000000000001",
	}
	for _, value := range invalid {
		if _, _, err := parseProviderJobReference(
			value, "billing-provider-command",
		); err == nil {
			t.Fatalf("invalid provider reference accepted: %s", value)
		}
	}
}

func TestProviderCommandProducerAndWorkerReferenceContractMatch(t *testing.T) {
	commandID := uuid.MustParse("018f1000-0000-7000-8000-000000000001")
	job, err := providerapp.NewProviderCommandJob(
		"manual", commandID,
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		time.Date(2026, 7, 17, 12, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatal(err)
	}
	provider, parsedID, err := parseProviderJobReference(
		job.PayloadReference(), "billing-provider-command",
	)
	if err != nil || provider != "manual" || parsedID != commandID {
		t.Fatalf("provider command reference = %s/%s, %v", provider, parsedID, err)
	}
}
