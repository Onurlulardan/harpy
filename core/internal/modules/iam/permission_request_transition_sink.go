package iam

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissionrequest"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const permissionRequestTransitionSchemaSHA256 = "65c815b2163d6f3ad5ee04f6615b1cf92b4d2161bee3796f1e49fd3872c5a399"

func (sink *PermissionRequestDecisionSink) RecordPermissionRequestTransition(
	ctx context.Context,
	scope platformtx.TenantScope,
	record permissionrequest.TransitionRecord,
) error {
	if sink == nil || sink.audit == nil || sink.now == nil || record.Validate() != nil {
		return permissionrequest.ErrDecisionEffects
	}
	authority, err := platformtx.MintOwnerAuthority(scope, "iam")
	if err != nil || !authority.OwnsScope(scope) {
		return permissionrequest.ErrDecisionEffects
	}
	clock, err := sink.audit.WriteTransitionAudit(ctx, scope, record)
	if err != nil {
		return err
	}
	writer, err := asyncpostgres.Bind(scope, authority, sink.catalogs)
	if err != nil {
		return errors.Join(permissionrequest.ErrDecisionEffects, err)
	}
	envelope, err := sink.transitionEnvelope(record, clock, scope, authority)
	if err != nil {
		return errors.Join(permissionrequest.ErrDecisionEffects, err)
	}
	result, err := writer.EnqueueEvent(ctx, envelope)
	if err != nil || result.ID != record.OperationID.String() {
		return errors.Join(permissionrequest.ErrDecisionEffects, err)
	}
	return nil
}

func (sink *PermissionRequestDecisionSink) transitionEnvelope(
	record permissionrequest.TransitionRecord,
	clock int64,
	scope platformtx.TenantScope,
	authority platformtx.OwnerAuthority,
) (eventsv1.Envelope, error) {
	payload, digest, err := permissionRequestTransitionPayload(record, clock)
	if err != nil {
		return eventsv1.Envelope{}, err
	}
	_, producers, err := sink.catalogs.Snapshot()
	if err != nil {
		return eventsv1.Envelope{}, err
	}
	tenantID, operationID := scope.TenantID().String(), record.OperationID.String()
	value := eventsv1.Envelope{
		EventID: operationID, EventType: "iam.permission_request.transitioned", EventVersion: 1,
		TenantID: &tenantID, ProducerAuthority: rbacEventAuthority(authority),
		AggregateType: "iam.permission_request_group", AggregateID: record.GroupID.String(),
		AggregateVersion: record.Revision, EventSequence: record.Revision,
		OccurredAt: sink.now().UTC().Format(time.RFC3339Nano), ActorID: record.ActorID.String(),
		CorrelationID: operationID, CausationID: operationID, IdempotencyKey: operationID,
		BusinessEffectIdempotencyKey: operationID,
		PayloadSchemaSHA256:          permissionRequestTransitionSchemaSHA256,
		PayloadSHA256:                digest, Payload: payload,
	}
	return eventsv1.NewEnvelope(value, producers)
}

func permissionRequestTransitionPayload(
	record permissionrequest.TransitionRecord,
	clock int64,
) ([]byte, string, error) {
	raw, err := json.Marshal(struct {
		AuthorizationClock int64  `json:"authorization_clock"`
		Command            string `json:"command"`
		FromState          string `json:"from_state"`
		ToState            string `json:"to_state"`
		Revision           int64  `json:"revision"`
	}{clock, string(record.Command), string(record.From), string(record.To), record.Revision})
	if err != nil {
		return nil, "", err
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(raw, eventsv1.MaximumCanonicalPayloadBytes)
	if err != nil {
		return nil, "", err
	}
	digest := sha256.Sum256(canonical)
	return canonical, hex.EncodeToString(digest[:]), nil
}
