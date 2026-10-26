package billing

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"

	providerapp "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/provider"
	"github.com/google/uuid"
)

func TestProviderPaymentEventIsExactMinimalSchema(t *testing.T) {
	result := providerFinalizationFixture(
		providerapp.FinalizeSucceeded, providerapp.CommandSucceeded,
	)
	event, err := providerPaymentEvent(
		uuid.MustParse("018f1000-0000-7000-8000-000000000061"),
		result, providerOutcomeIDs(),
	)
	if err != nil {
		t.Fatal(err)
	}
	want := "{\"provider_command_id\":\"" + result.CommandID.String() +
		"\",\"provider_command_revision\":2}"
	digest := sha256.Sum256([]byte(want))
	if string(event.Payload) != want ||
		event.PayloadSHA256 != hex.EncodeToString(digest[:]) ||
		event.PayloadSchemaSHA256 != providerPaymentEventSchemaSHA256 ||
		event.AggregateType != "billing.payment" ||
		event.AggregateID != result.CommandID ||
		event.AggregateVersion != result.Revision ||
		event.EventSequence != result.Revision {
		t.Fatalf("payment event = %#v", event)
	}
	for _, forbidden := range []string{
		"secret", "receipt", "provider_reference", "error",
	} {
		if strings.Contains(string(event.Payload), forbidden) {
			t.Fatalf("payment event leaked %q: %s", forbidden, event.Payload)
		}
	}
}
