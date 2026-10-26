package postgres

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"
	"time"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func validPublisherCatalog(t *testing.T) eventsv1.ProducerCatalog {
	t.Helper()
	policy, err := eventsv1.NewProducerPolicy(
		"workflow.changed", 1, "workflow", strings.Repeat("b", 64),
	)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := eventsv1.NewProducerCatalog([]eventsv1.ProducerPolicy{policy})
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func validOutboxAuthority(t *testing.T, repository *workerRepository) OutboxAuthority {
	t.Helper()
	authority, err := newOutboxAuthority(
		repository.identity, "tenant", uuid.New(), "operations", 3, "workflow", 7,
	)
	if err != nil {
		t.Fatal(err)
	}
	return authority
}

func validPublishClaimRow(authority OutboxAuthority, publisherID uuid.UUID) publishClaimRow {
	payload := []byte(`{"value":1}`)
	digest := sha256.Sum256(payload)
	heartbeatAt := time.Now().UTC().Truncate(time.Microsecond)
	return publishClaimRow{
		eventID: uuid.New(), tenantID: pgtype.UUID{
			Bytes: [16]byte(authority.scopeID), Valid: true,
		},
		scopeKind: authority.scopeKind, scopeID: authority.scopeID,
		groupID: authority.groupID, groupGeneration: authority.groupGeneration,
		domain: authority.domain, generation: authority.generation,
		writerOrigin: "v2", contractVersion: eventsv1.EnvelopeContractVersion,
		eventType: "workflow.changed", eventVersion: 1,
		aggregateType: "workflow.instance", aggregateID: uuid.New(),
		aggregateVersion: 4, eventSequence: 2,
		occurredAt: heartbeatAt.Add(-time.Minute), actorID: uuid.New(),
		correlationID: "corr.publisher.1", causationID: uuid.New(),
		idempotencyKey: "event-1", businessKey: "effect-1", payload: payload,
		payloadSHA256: hex.EncodeToString(digest[:]), schemaSHA256: strings.Repeat("b", 64),
		deliveryAttempts: 1, leaseOwnerID: publisherID, leaseGeneration: 1,
		heartbeatAt: heartbeatAt, leaseUntil: heartbeatAt.Add(publishLeaseDuration),
		nextAttemptAt: heartbeatAt.Add(-time.Second),
	}
}

func validPublishClaim(
	t *testing.T,
	repository *workerRepository,
) PublishClaim {
	t.Helper()
	authority := validOutboxAuthority(t, repository)
	publisherID := uuid.New()
	claim, err := newPublishClaim(
		validPublishClaimRow(authority, publisherID), authority,
		publisherID, validPublisherCatalog(t),
	)
	if err != nil {
		t.Fatal(err)
	}
	return claim
}
