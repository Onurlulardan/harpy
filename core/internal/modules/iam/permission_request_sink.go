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
	permissionpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/permissionrequest"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissionrequest"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const permissionRequestPayloadSchemaSHA256 = "f42bbb6b9274fc239afed496f6553340cc514a5bd8eb5c1549d002fe16e4997d"

type PermissionRequestClock func() time.Time

type PermissionRequestDecisionSink struct {
	audit    permissionrequest.DecisionAuditStore
	catalogs platformasync.WriteCatalogs
	now      PermissionRequestClock
}

func NewPermissionRequestDecisionSink(
	catalogs platformasync.WriteCatalogs,
	now PermissionRequestClock,
) (*PermissionRequestDecisionSink, error) {
	if now == nil {
		return nil, permissionrequest.ErrDecisionEffects
	}
	if _, _, err := catalogs.Snapshot(); err != nil {
		return nil, permissionrequest.ErrDecisionEffects
	}
	audit, err := permissionpostgres.NewPermissionRequestDecisionAuditStore()
	if err != nil {
		return nil, err
	}
	return &PermissionRequestDecisionSink{audit: audit, catalogs: catalogs, now: now}, nil
}

func (sink *PermissionRequestDecisionSink) RecordPermissionRequestDecision(
	ctx context.Context,
	scope platformtx.TenantScope,
	record permissionrequest.DecisionRecord,
) error {
	if sink == nil || sink.audit == nil || sink.now == nil || record.Validate() != nil {
		return permissionrequest.ErrDecisionEffects
	}
	authority, err := platformtx.MintOwnerAuthority(scope, "iam")
	if err != nil || !authority.OwnsScope(scope) {
		return permissionrequest.ErrDecisionEffects
	}
	if err := sink.audit.WriteDecisionAudit(ctx, scope, record); err != nil {
		return err
	}
	writer, err := asyncpostgres.Bind(scope, authority, sink.catalogs)
	if err != nil {
		return errors.Join(permissionrequest.ErrDecisionEffects, err)
	}
	envelope, err := sink.envelope(record, scope, authority)
	if err != nil {
		return errors.Join(permissionrequest.ErrDecisionEffects, err)
	}
	result, err := writer.EnqueueEvent(ctx, envelope)
	if err != nil || result.ID != record.OperationID.String() {
		return errors.Join(permissionrequest.ErrDecisionEffects, err)
	}
	return nil
}

func (sink *PermissionRequestDecisionSink) envelope(
	record permissionrequest.DecisionRecord,
	scope platformtx.TenantScope,
	authority platformtx.OwnerAuthority,
) (eventsv1.Envelope, error) {
	payload, digest, err := permissionRequestEventPayload(record)
	if err != nil {
		return eventsv1.Envelope{}, err
	}
	_, producers, err := sink.catalogs.Snapshot()
	if err != nil {
		return eventsv1.Envelope{}, err
	}
	tenantID, operationID := scope.TenantID().String(), record.OperationID.String()
	value := eventsv1.Envelope{
		EventID: operationID, EventType: "iam.permission_request.decided", EventVersion: 1,
		TenantID: &tenantID, ProducerAuthority: rbacEventAuthority(authority),
		AggregateType: "iam.permission_request_group", AggregateID: record.GroupID.String(),
		AggregateVersion: record.Result.Revision, EventSequence: record.Result.Revision,
		OccurredAt: sink.now().UTC().Format(time.RFC3339Nano), ActorID: record.ReviewerID.String(),
		CorrelationID: operationID, CausationID: operationID, IdempotencyKey: operationID,
		BusinessEffectIdempotencyKey: operationID,
		PayloadSchemaSHA256:          permissionRequestPayloadSchemaSHA256,
		PayloadSHA256:                digest, Payload: payload,
	}
	return eventsv1.NewEnvelope(value, producers)
}

func permissionRequestEventPayload(
	record permissionrequest.DecisionRecord,
) ([]byte, string, error) {
	raw, err := json.Marshal(struct {
		AuthorizationClock int64  `json:"authorization_clock"`
		Decision           string `json:"decision"`
		Result             string `json:"result"`
		Revision           int64  `json:"revision"`
	}{record.Result.AuthorizationClock, string(record.Decision), record.Result.Code, record.Result.Revision})
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

var _ permissionrequest.DecisionEffects = (*PermissionRequestDecisionSink)(nil)
