package v1bridge

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
	"github.com/google/uuid"
)

type bridgeTestScope struct{ tenant, actor uuid.UUID }

func (value bridgeTestScope) TenantID() uuid.UUID { return value.tenant }
func (value bridgeTestScope) ActorID() uuid.UUID  { return value.actor }
func (bridgeTestScope) OperationID() string       { return "workflow.notification.intent" }

type bridgeTestStore struct {
	authority Authority
	records   map[string]Intent
	jobs      []Job
}

func (store *bridgeTestStore) LockAuthority(context.Context, string) (Authority, error) {
	return store.authority, nil
}
func (store *bridgeTestStore) FindIntent(_ context.Context, producer string,
	hash [32]byte, recipient [16]byte,
) (Intent, bool, error) {
	value, found := store.records[producer+string(hash[:])+string(recipient[:])]
	return value, found, nil
}
func (store *bridgeTestStore) InsertIntent(_ context.Context, value Intent) error {
	key := value.Producer + string(value.EffectKeyHash[:]) + string(value.RecipientID[:])
	store.records[key] = value
	return nil
}
func (store *bridgeTestStore) Enqueue(_ context.Context, value Job) error {
	store.jobs = append(store.jobs, value)
	return nil
}

type bridgeTestBinder struct{ store *bridgeTestStore }

func (binder bridgeTestBinder) BindBridge(context.Context, api.Scope) (Store, error) {
	return binder.store, nil
}

type bridgeTestFinal struct{ calls int }

func (final *bridgeTestFinal) RecordIntent(context.Context, api.Scope,
	api.IntentRequest,
) (api.IntentResult, error) {
	final.calls++
	return api.IntentResult{Outcome: "created"}, nil
}

type bridgeTestClock struct{ now time.Time }

func (clock bridgeTestClock) Now() time.Time { return clock.now }

type bridgeTestIDs struct{ id domain.ID }

func (source bridgeTestIDs) New(time.Time) (domain.ID, error) { return source.id, nil }

func TestRecordIntentUsesOnlyOneWriterAndReplaysV1Intent(t *testing.T) {
	now := time.Date(2026, 7, 18, 16, 0, 0, 0, time.UTC)
	jobID, _ := domain.ParseID("019b0000-0000-7000-8000-000000000061")
	store := &bridgeTestStore{authority: v1BridgeAuthority(), records: map[string]Intent{}}
	final := &bridgeTestFinal{}
	router, err := NewRouter(bridgeTestBinder{store}, final,
		bridgeTestClock{now}, bridgeTestIDs{jobID})
	if err != nil {
		t.Fatal(err)
	}
	scope, request := bridgeRequest(t)
	first, err := router.RecordIntent(t.Context(), scope, request)
	if err != nil || first.MigrationJobID != api.ID(jobID) ||
		first.Outcome != "migration_queued" || final.calls != 0 || len(store.jobs) != 1 {
		t.Fatalf("first=%+v jobs=%d final=%d err=%v", first, len(store.jobs), final.calls, err)
	}
	second, err := router.RecordIntent(t.Context(), scope, request)
	if err != nil || second != (api.IntentResult{MigrationJobID: api.ID(jobID),
		Replay: true, Outcome: "migration_queued"}) || len(store.jobs) != 1 {
		t.Fatalf("replay=%+v jobs=%d err=%v", second, len(store.jobs), err)
	}
	request.NotificationType = "task.changed"
	if _, err = router.RecordIntent(t.Context(), scope, request); !errors.Is(err, api.ErrEffectKeyReused) {
		t.Fatalf("effect reuse error=%v", err)
	}
}

func TestRecordIntentSelectsV2AndRejectsFencedAuthority(t *testing.T) {
	jobID, _ := domain.ParseID("019b0000-0000-7000-8000-000000000062")
	store := &bridgeTestStore{authority: Authority{ProducerWriter: WriterV2,
		ProducerState: "v2", ProducerGeneration: 4, NotificationWriter: WriterV2,
		NotificationState: "v2", NotificationGeneration: 6}, records: map[string]Intent{}}
	final := &bridgeTestFinal{}
	router, _ := NewRouter(bridgeTestBinder{store}, final,
		bridgeTestClock{time.Now().UTC()}, bridgeTestIDs{jobID})
	scope, request := bridgeRequest(t)
	result, err := router.RecordIntent(t.Context(), scope, request)
	if err != nil || result.Outcome != "created" || final.calls != 1 || len(store.jobs) != 0 {
		t.Fatalf("V2 result=%+v final=%d jobs=%d err=%v", result, final.calls, len(store.jobs), err)
	}
	store.authority.NotificationWriter = ""
	store.authority.NotificationState = "fenced"
	if _, err = router.RecordIntent(t.Context(), scope, request); !errors.Is(err, api.ErrUnavailable) || final.calls != 1 {
		t.Fatalf("fenced error=%v final=%d", err, final.calls)
	}
}

func v1BridgeAuthority() Authority {
	return Authority{ProducerWriter: WriterV2, ProducerState: "v2", ProducerGeneration: 4,
		NotificationWriter: WriterV1, NotificationState: "v1", NotificationGeneration: 5}
}

func bridgeRequest(t *testing.T) (bridgeTestScope, api.IntentRequest) {
	t.Helper()
	tenant, actor, recipient := uuid.New(), uuid.New(), uuid.New()
	data := []api.TemplateValue{{Key: "task", Kind: api.ValueString, String: "A-42"}}
	digest, _ := api.CanonicalDataDigest(data)
	return bridgeTestScope{tenant: tenant, actor: actor}, api.IntentRequest{
		Producer: "workflow", NotificationType: "task.ready",
		RecipientPrincipalID: api.ID(recipient), TemplateVersionID: api.ID(uuid.New()), LocaleInput: "tr",
		EffectKey: "workflow:instance:node", Category: api.CategoryWorkflowAction,
		Channels: []api.Channel{api.ChannelInApp, api.ChannelEmail}, Data: data,
		DataDigest: digest, CorrelationID: api.ID(uuid.New()), CausationID: api.ID(uuid.New())}
}
