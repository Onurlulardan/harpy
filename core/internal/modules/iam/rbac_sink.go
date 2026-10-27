package iam

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	rbacchange "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/rbacchange"
	rbacapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/rbacchange"
	rbacports "github.com/Onurlulardan/harpy/core/internal/modules/iam/ports/rbac"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const rbacPayloadSchemaSHA256 = "ec25ad29fe3139a6db7b0ee837d3b9a28371867a671c88322a1bfa14dca64657"

type RBACClock func() time.Time

type RBACChangeSink struct {
	audit    rbacports.AuditStore[platformtx.TenantScope]
	catalogs platformasync.WriteCatalogs
	now      RBACClock
}

func NewRBACChangeSink(catalogs platformasync.WriteCatalogs, now RBACClock) (*RBACChangeSink, error) {
	if now == nil {
		return nil, rbacapp.ErrSink
	}
	if _, _, err := catalogs.Snapshot(); err != nil {
		return nil, rbacapp.ErrSink
	}
	audit, err := rbacchange.NewStore()
	if err != nil {
		return nil, err
	}
	return &RBACChangeSink{audit: audit, catalogs: catalogs, now: now}, nil
}

func (sink *RBACChangeSink) WriteChange(ctx context.Context, scope platformtx.TenantScope,
	change rbacapp.Change) error {
	if sink == nil || sink.audit == nil || sink.now == nil || change.Validate() != nil {
		return rbacapp.ErrSink
	}
	authority, err := platformtx.MintOwnerAuthority(scope, "iam")
	if err != nil || !authority.OwnsScope(scope) {
		return rbacapp.ErrSink
	}
	if err := sink.audit.WriteAudit(ctx, scope, change); err != nil {
		return err
	}
	writer, err := asyncpostgres.Bind(scope, authority, sink.catalogs)
	if err != nil {
		return errors.Join(rbacapp.ErrSink, err)
	}
	envelope, err := sink.envelope(change, scope, authority)
	if err != nil {
		return errors.Join(rbacapp.ErrSink, err)
	}
	result, err := writer.EnqueueEvent(ctx, envelope)
	if err != nil || result.ID != change.OperationID.String() {
		return errors.Join(rbacapp.ErrSink, err)
	}
	return nil
}

func (sink *RBACChangeSink) envelope(change rbacapp.Change, scope platformtx.TenantScope,
	authority platformtx.OwnerAuthority) (eventsv1.Envelope, error) {
	payload, digest, err := rbacEventPayload(change)
	if err != nil {
		return eventsv1.Envelope{}, err
	}
	_, producers, err := sink.catalogs.Snapshot()
	if err != nil {
		return eventsv1.Envelope{}, err
	}
	tenantID := scope.TenantID().String()
	operationID := change.OperationID.String()
	value := eventsv1.Envelope{
		EventID: operationID, EventType: "iam.rbac.changed", EventVersion: 1,
		TenantID: &tenantID, ProducerAuthority: rbacEventAuthority(authority),
		AggregateType: rbacAggregateType(change.CommandKey),
		AggregateID:   change.AggregateID.String(), AggregateVersion: change.Revision,
		EventSequence: change.Revision, OccurredAt: sink.now().UTC().Format(time.RFC3339Nano),
		ActorID: change.ActorID.String(), CorrelationID: operationID, CausationID: operationID,
		IdempotencyKey: operationID, BusinessEffectIdempotencyKey: operationID,
		PayloadSchemaSHA256: rbacPayloadSchemaSHA256, PayloadSHA256: digest, Payload: payload,
	}
	return eventsv1.NewEnvelope(value, producers)
}

func rbacEventPayload(change rbacapp.Change) ([]byte, string, error) {
	raw, err := json.Marshal(struct {
		Clock    int64  `json:"clock"`
		Command  string `json:"command"`
		Result   string `json:"result"`
		Revision int64  `json:"revision"`
	}{change.Clock, change.CommandKey, change.ResultCode, change.Revision})
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

var _ rbacports.ChangeSink[platformtx.TenantScope] = (*RBACChangeSink)(nil)

func rbacAggregateType(commandKey string) string {
	if index := strings.LastIndexByte(commandKey, '.'); index > 0 {
		return commandKey[:index]
	}
	return "iam.rbac.change"
}

func rbacEventAuthority(authority platformtx.OwnerAuthority) eventsv1.Authority {
	return eventsv1.Authority{
		ScopeKind: eventsv1.ScopeTenant, ScopeID: authority.ScopeID().String(),
		Domain: authority.Domain(), CutoverGeneration: authority.Generation(),
		WriterOrigin: eventsv1.WriterV2, GroupID: authority.GroupID(),
		GroupGeneration: authority.GroupGeneration(),
	}
}
