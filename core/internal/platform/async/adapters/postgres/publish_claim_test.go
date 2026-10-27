package postgres

import (
	"errors"
	"strings"
	"testing"
	"time"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	"github.com/google/uuid"
)

func TestPublishClaimReconstructsCatalogValidatedImmutableEnvelope(t *testing.T) {
	t.Parallel()
	repository := newWorkerTestRepository()
	authority := validOutboxAuthority(t, repository)
	publisherID := uuid.New()
	row := validPublishClaimRow(authority, publisherID)
	claim, err := newPublishClaim(row, authority, publisherID, validPublisherCatalog(t))
	if err != nil {
		t.Fatal(err)
	}
	first := claim.Envelope()
	first.EventType = "tampered.event"
	first.Payload[0] = '['
	*first.TenantID = uuid.NewString()
	second := claim.Envelope()
	if second.EventType != "workflow.changed" || string(second.Payload) != `{"value":1}` ||
		second.TenantID == nil || *second.TenantID != authority.scopeID.String() ||
		claim.PublisherID() != publisherID || claim.DeliveryAttempt() != 1 ||
		claim.LeaseGeneration() != 1 ||
		!claim.Lease().Until().Equal(claim.Lease().HeartbeatAt().Add(60*time.Second)) {
		t.Fatalf("publish claim mutated: %#v", claim)
	}
}

func TestPublishClaimRejectsCatalogAndLeaseMismatch(t *testing.T) {
	t.Parallel()
	repository := newWorkerTestRepository()
	authority := validOutboxAuthority(t, repository)
	publisherID := uuid.New()
	row := validPublishClaimRow(authority, publisherID)
	policy, err := eventsv1.NewProducerPolicy(
		"workflow.changed", 1, "workflow", strings.Repeat("c", 64),
	)
	if err != nil {
		t.Fatal(err)
	}
	wrongCatalog, err := eventsv1.NewProducerCatalog([]eventsv1.ProducerPolicy{policy})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := newPublishClaim(
		row, authority, publisherID, wrongCatalog,
	); err == nil {
		t.Fatal("schema-mismatched catalog accepted")
	}
	row = validPublishClaimRow(authority, publisherID)
	row.leaseUntil = row.heartbeatAt.Add(59 * time.Second)
	if _, err := newPublishClaim(
		row, authority, publisherID, validPublisherCatalog(t),
	); !errors.Is(err, ErrPublishClaimInvalid) {
		t.Fatalf("invalid lease error = %v", err)
	}
}
