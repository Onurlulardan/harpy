package v1cutover

import (
	"context"
	"crypto/sha256"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/api"
	"github.com/google/uuid"
)

type cutoverScope struct{ tenant, actor uuid.UUID }

func (value cutoverScope) TenantID() uuid.UUID { return value.tenant }
func (value cutoverScope) ActorID() uuid.UUID  { return value.actor }
func (cutoverScope) OperationID() string       { return "notifications.cutover.test" }

type cutoverBinderFake struct{ store *cutoverStoreFake }

func (value cutoverBinderFake) BindCutover(context.Context, api.Scope) (Store, error) {
	return value.store, nil
}

type cutoverStoreFake struct{ authority Authority }

func (store *cutoverStoreFake) CreateAuthority(_ context.Context, origin Origin) (Authority, error) {
	writer := "v1"
	if origin == OriginFresh {
		writer = "v2"
	}
	store.authority = Authority{Writer: writer, State: "writable", Generation: 1, RowVersion: 2}
	return store.authority, nil
}
func (store *cutoverStoreFake) PauseV1(context.Context, Authority) (Authority, error) {
	store.authority.State, store.authority.RowVersion = "quiescing", store.authority.RowVersion+1
	return store.authority, nil
}
func (*cutoverStoreFake) RecordEvidence(_ context.Context, value Inventory) (Evidence, error) {
	return Evidence{ID: uuid.MustParse("30000000-0000-4000-8000-000000000001"),
		Inventory: value, ExpiresAt: value.ObservedAt.Add(5 * time.Minute)}, nil
}
func (store *cutoverStoreFake) DisableV1(_ context.Context, _ Authority, evidence Evidence) (Authority, error) {
	if !evidence.Inventory.Zero() {
		return Authority{}, ErrNotReady
	}
	store.authority.State, store.authority.RowVersion = "fenced", store.authority.RowVersion+1
	return store.authority, nil
}
func (store *cutoverStoreFake) ActivateV2(context.Context, Authority) (Authority, error) {
	store.authority.Writer, store.authority.State = "v2", "writable"
	store.authority.Generation++
	store.authority.RowVersion++
	return store.authority, nil
}

func TestActivateV2RequiresZeroInventoryAndExplicitFence(t *testing.T) {
	store := &cutoverStoreFake{}
	application, _ := NewApplication(cutoverBinderFake{store})
	scope := cutoverScope{tenant: uuid.MustParse("10000000-0000-4000-8000-000000000001"),
		actor: uuid.MustParse("10000000-0000-4000-8000-000000000002")}
	authority, err := application.CreateAuthority(t.Context(), scope, OriginV1Adopted)
	if err != nil || authority.Writer != "v1" {
		t.Fatalf("bootstrap=%+v err=%v", authority, err)
	}
	authority, err = application.PauseV1(t.Context(), scope, authority)
	if err != nil || authority.State != "quiescing" {
		t.Fatalf("quiesce=%+v err=%v", authority, err)
	}
	inventory := Inventory{RunID: uuid.MustParse("20000000-0000-4000-8000-000000000001"),
		SourceQueueDepth: 1, SourceEvidenceDigest: sha256.Sum256([]byte("evidence")),
		ObservedAt: time.Now().UTC()}
	evidence, _ := application.RecordEvidence(t.Context(), scope, inventory)
	if _, err = application.DisableV1(t.Context(), scope, authority, evidence); err == nil {
		t.Fatal("non-zero source queue passed fence")
	}
	inventory.SourceQueueDepth = 0
	evidence, _ = application.RecordEvidence(t.Context(), scope, inventory)
	authority, err = application.DisableV1(t.Context(), scope, authority, evidence)
	if err != nil || authority.State != "fenced" {
		t.Fatalf("fence=%+v err=%v", authority, err)
	}
	authority, err = application.ActivateV2(t.Context(), scope, authority)
	if err != nil || authority.Writer != "v2" || authority.Generation != 2 {
		t.Fatalf("activate=%+v err=%v", authority, err)
	}
}
