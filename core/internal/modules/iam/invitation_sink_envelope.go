package iam

import (
	"time"

	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/invitation"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (sink *InvitationChangeSink) tenantEnvelope(
	change invitation.AuditChange,
	scope platformtx.TenantScope,
	authority platformtx.OwnerAuthority,
) (eventsv1.Envelope, error) {
	return sink.envelope(change, scope.TenantID(), eventsv1.Authority{
		ScopeKind: eventsv1.ScopeTenant, ScopeID: authority.ScopeID().String(),
		Domain: authority.Domain(), CutoverGeneration: authority.Generation(),
		WriterOrigin: eventsv1.WriterV2, GroupID: authority.GroupID(),
		GroupGeneration: authority.GroupGeneration(),
	})
}

func (sink *InvitationChangeSink) credentialEnvelope(
	change invitation.AuditChange,
	scope platformtx.CredentialTenantScope,
) (eventsv1.Envelope, error) {
	return sink.envelope(change, scope.TenantID(), eventsv1.Authority{
		ScopeKind: eventsv1.ScopeTenant, ScopeID: scope.TenantID().String(),
		Domain: "iam", CutoverGeneration: scope.IAMGeneration(),
		WriterOrigin: eventsv1.WriterV2, GroupID: "iam_system_resources",
		GroupGeneration: scope.AuthorityGroupGeneration(),
	})
}

func (sink *InvitationChangeSink) envelope(
	change invitation.AuditChange,
	tenantID interface{ String() string },
	authority eventsv1.Authority,
) (eventsv1.Envelope, error) {
	payload, digest, err := invitationEventPayload(change)
	if err != nil {
		return eventsv1.Envelope{}, err
	}
	_, producers, err := sink.catalogs.Snapshot()
	if err != nil {
		return eventsv1.Envelope{}, err
	}
	operationID := change.OperationID.String()
	tenant := tenantID.String()
	value := eventsv1.Envelope{
		EventID: operationID, EventType: "iam.invitation.changed", EventVersion: 1,
		TenantID: &tenant, ProducerAuthority: authority,
		AggregateType: "iam.invitation", AggregateID: change.InvitationID.String(),
		AggregateVersion: change.ResultRevision, EventSequence: change.ResultRevision,
		OccurredAt: sink.now().UTC().Format(time.RFC3339Nano),
		ActorID:    change.ActorID.String(), CorrelationID: operationID,
		CausationID: operationID, IdempotencyKey: operationID,
		BusinessEffectIdempotencyKey: operationID,
		PayloadSchemaSHA256:          invitationPayloadSchemaSHA256,
		PayloadSHA256:                digest, Payload: payload,
	}
	return eventsv1.NewEnvelope(value, producers)
}
