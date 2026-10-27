package postgres

import (
	"fmt"
	"time"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	"github.com/google/uuid"
)

type PublishClaim struct {
	authority       OutboxAuthority
	eventID         uuid.UUID
	envelope        eventsv1.Envelope
	publisherID     uuid.UUID
	deliveryAttempt int
	leaseGeneration int64
	lease           PublishLease
}

func newPublishClaim(
	row publishClaimRow,
	authority OutboxAuthority,
	publisherID uuid.UUID,
	catalog eventsv1.ProducerCatalog,
) (PublishClaim, error) {
	tenantID, tenantValid := publishTenant(row)
	if !authority.valid(authority.owner) || row.eventID == uuid.Nil ||
		row.scopeKind != authority.scopeKind || row.scopeID != authority.scopeID ||
		row.groupID != authority.groupID ||
		row.groupGeneration != authority.groupGeneration ||
		row.domain != authority.domain || row.generation != authority.generation ||
		(!tenantValid && !(row.scopeKind == "global" && !row.tenantID.Valid)) ||
		publisherID == uuid.Nil || row.leaseOwnerID != publisherID ||
		row.deliveryAttempts < 1 || int64(row.deliveryAttempts) != row.leaseGeneration ||
		row.nextAttemptAt.After(row.heartbeatAt) {
		return PublishClaim{}, ErrPublishClaimInvalid
	}
	lease, err := newPublishLease(row.heartbeatAt, row.leaseUntil)
	if err != nil {
		return PublishClaim{}, err
	}
	envelope, err := makePublishEnvelope(row, tenantID, tenantValid, catalog)
	if err != nil {
		return PublishClaim{}, fmt.Errorf("reconstruct claimed event: %w", err)
	}
	return PublishClaim{
		authority: authority, eventID: row.eventID,
		envelope: envelope, publisherID: publisherID,
		deliveryAttempt: int(row.deliveryAttempts),
		leaseGeneration: row.leaseGeneration, lease: lease,
	}, nil
}

func (claim PublishClaim) valid(repository *workerRepository) bool {
	return repository.valid() && claim.authority.valid(repository.identity) &&
		claim.eventID != uuid.Nil && claim.publisherID != uuid.Nil &&
		claim.deliveryAttempt > 0 && claim.leaseGeneration == int64(claim.deliveryAttempt) &&
		claim.envelope.EventID == claim.eventID.String() &&
		!claim.lease.heartbeatAt.IsZero() &&
		claim.lease.until.Equal(claim.lease.heartbeatAt.Add(publishLeaseDuration))
}

func publishTenant(row publishClaimRow) (uuid.UUID, bool) {
	if !row.tenantID.Valid {
		return uuid.Nil, false
	}
	tenantID := uuid.UUID(row.tenantID.Bytes)
	return tenantID, row.scopeKind == "tenant" && tenantID == row.scopeID
}

func makePublishEnvelope(
	row publishClaimRow,
	tenantID uuid.UUID,
	tenantValid bool,
	catalog eventsv1.ProducerCatalog,
) (eventsv1.Envelope, error) {
	var tenant *string
	if tenantValid {
		value := tenantID.String()
		tenant = &value
	}
	return eventsv1.NewEnvelope(eventsv1.Envelope{
		ContractVersion: row.contractVersion, EventID: row.eventID.String(),
		EventType: row.eventType, EventVersion: int(row.eventVersion), TenantID: tenant,
		ProducerAuthority: eventsv1.Authority{
			ScopeKind: eventsv1.ScopeKind(row.scopeKind), ScopeID: row.scopeID.String(),
			Domain: row.domain, CutoverGeneration: row.generation,
			WriterOrigin: eventsv1.WriterOrigin(row.writerOrigin),
			GroupID:      row.groupID, GroupGeneration: row.groupGeneration,
		},
		AggregateType: row.aggregateType, AggregateID: row.aggregateID.String(),
		AggregateVersion: row.aggregateVersion, EventSequence: row.eventSequence,
		OccurredAt: row.occurredAt.UTC().Format(time.RFC3339Nano),
		ActorID:    row.actorID.String(), CorrelationID: row.correlationID,
		CausationID: row.causationID.String(), IdempotencyKey: row.idempotencyKey,
		BusinessEffectIdempotencyKey: row.businessKey,
		PayloadSchemaSHA256:          row.schemaSHA256, PayloadSHA256: row.payloadSHA256,
		Payload: row.payload,
	}, catalog)
}

func (claim PublishClaim) Authority() OutboxAuthority { return claim.authority }
func (claim PublishClaim) Envelope() eventsv1.Envelope {
	copy := claim.envelope
	copy.Payload = claim.envelope.PayloadClone()
	if claim.envelope.TenantID != nil {
		tenantID := *claim.envelope.TenantID
		copy.TenantID = &tenantID
	}
	return copy
}
func (claim PublishClaim) PublisherID() uuid.UUID { return claim.publisherID }
func (claim PublishClaim) DeliveryAttempt() int   { return claim.deliveryAttempt }
func (claim PublishClaim) LeaseGeneration() int64 { return claim.leaseGeneration }
func (claim PublishClaim) Lease() PublishLease    { return claim.lease }
