package v1bridge

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"reflect"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

type Router struct {
	binder Binder
	final  api.NotificationIntentPort
	clock  ports.Clock
	ids    ports.IDSource
}

func NewRouter(binder Binder, final api.NotificationIntentPort, clock ports.Clock,
	ids ports.IDSource,
) (*Router, error) {
	if bridgeNil(binder) || bridgeNil(final) || bridgeNil(clock) || bridgeNil(ids) {
		return nil, ErrConfiguration
	}
	return &Router{binder: binder, final: final, clock: clock, ids: ids}, nil
}

func (router *Router) RecordIntent(ctx context.Context, scope api.Scope,
	request api.IntentRequest,
) (api.IntentResult, error) {
	if router == nil || ctx == nil || scope == nil || request.Validate() != nil {
		return api.IntentResult{}, api.ErrIntentInvalid
	}
	store, err := router.binder.BindBridge(ctx, scope)
	if err != nil || store == nil {
		return api.IntentResult{}, api.ErrUnavailable
	}
	authority, err := store.LockAuthority(ctx, request.Producer)
	if err != nil || !authority.valid() {
		return api.IntentResult{}, api.ErrUnavailable
	}
	if authority.NotificationWriter == WriterV2 {
		return router.final.RecordIntent(ctx, scope, request)
	}
	return router.recordV1(ctx, scope, store, authority, request)
}

func (router *Router) recordV1(ctx context.Context, scope api.Scope, store Store,
	authority Authority, request api.IntentRequest,
) (api.IntentResult, error) {
	payload, canonicalDigest, err := canonicalPayload(request)
	if err != nil {
		return api.IntentResult{}, err
	}
	recipient, err := domain.IDFromBytes(request.RecipientPrincipalID[:])
	if err != nil {
		return api.IntentResult{}, api.ErrIntentInvalid
	}
	effectHash := sha256.Sum256([]byte(request.EffectKey))
	existing, found, err := store.FindIntent(ctx, request.Producer, effectHash,
		request.RecipientPrincipalID)
	if err != nil {
		return api.IntentResult{}, api.ErrUnavailable
	}
	if found {
		return replayResult(existing, authority, canonicalDigest)
	}
	now := router.clock.Now().UTC()
	jobID, err := router.ids.New(now)
	if err != nil || now.IsZero() {
		return api.IntentResult{}, api.ErrUnavailable
	}
	record := Intent{Producer: request.Producer, EffectKeyHash: effectHash,
		CanonicalDigest: canonicalDigest, PayloadDigest: sha256.Sum256(payload),
		RecipientID: recipient, JobID: jobID,
		ProducerGeneration:     authority.ProducerGeneration,
		NotificationGeneration: authority.NotificationGeneration,
		Payload:                payload, State: StateQueued, CreatedAt: now, UpdatedAt: now, Revision: 1}
	if err = store.InsertIntent(ctx, record); err != nil {
		return api.IntentResult{}, api.ErrUnavailable
	}
	tenantUUID, tenantScoped := tenantIdentity(scope)
	if !tenantScoped {
		return api.IntentResult{}, api.ErrIntentInvalid
	}
	tenant, err := domain.IDFromBytes(tenantUUID[:])
	if err != nil {
		return api.IntentResult{}, api.ErrIntentInvalid
	}
	reference := "notification-v1-bridge://tenant/" + tenant.String() + "/intent/" + jobID.String()
	job := Job{ID: jobID, TenantID: tenant, Producer: request.Producer,
		ProducerGeneration:     authority.ProducerGeneration,
		NotificationGeneration: authority.NotificationGeneration,
		PayloadReference:       reference, PayloadDigestHex: hex.EncodeToString(record.PayloadDigest[:]),
		EffectKey: "notifications.v1:" + hex.EncodeToString(effectHash[:]), RunAt: now}
	if err = store.Enqueue(ctx, job); err != nil {
		return api.IntentResult{}, api.ErrUnavailable
	}
	return api.IntentResult{MigrationJobID: api.ID(jobID),
		Outcome: "migration_queued"}, nil
}

func replayResult(existing Intent, authority Authority,
	digest [32]byte,
) (api.IntentResult, error) {
	if existing.CanonicalDigest != digest ||
		existing.ProducerGeneration != authority.ProducerGeneration ||
		existing.NotificationGeneration != authority.NotificationGeneration {
		return api.IntentResult{}, api.ErrEffectKeyReused
	}
	outcome := string(existing.State)
	if existing.State == StateQueued {
		outcome = "migration_queued"
	}
	return api.IntentResult{MigrationJobID: api.ID(existing.JobID),
		Replay: true, Outcome: outcome}, nil
}

func bridgeNil(value any) bool {
	if value == nil {
		return true
	}
	raw := reflect.ValueOf(value)
	return raw.Kind() == reflect.Pointer && raw.IsNil()
}

var _ api.NotificationIntentPort = (*Router)(nil)
