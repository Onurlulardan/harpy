package asyncadoptionv1

import (
	"encoding/json"
	"strings"
	"testing"
)

const (
	testRunID   = "11111111-1111-4111-8111-111111111111"
	testScopeID = "22222222-2222-4222-8222-222222222222"
	testDigest  = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	testTime    = "2026-07-15T12:00:00Z"
)

func TestInventoryFreezesV1SurfaceSemanticsAndCanonicalRoot(t *testing.T) {
	items := []Item{
		newTestItem(t, "platform.future_state", ClassFuture, SurfaceFutureState, "v1.platform-future-state.v1", "future-1"),
		newTestItem(t, "platform.running_state", ClassRunning, SurfaceRunningState, "v1.platform-running-state.v1", "running-1"),
		newTestItem(t, "platform.scheduler", ClassTimer, SurfaceScheduler, "v1.platform-scheduler.v1", "occurrence-1"),
		newTestItem(t, V1ScheduleQueue, ClassQueue, SurfaceQueue, V1QueueContract, "schedule-1"),
	}
	manifest, err := NewInventory(Inventory{
		RunID: testRunID, ScopeKind: "tenant", ScopeID: testScopeID,
		Domain: "platform", GroupID: "platform", SourceGeneration: 8,
		TargetGeneration: 9, SourceGroupGeneration: 11, TargetGroupGeneration: 12,
		ComponentManifestSHA256: testDigest,
		ExpectedEnqueueFences:   1, ExpectedSchedulerFences: 1, ExpectedConsumerFences: 3,
		CapturedAt: testTime, Items: items,
	})
	if err != nil {
		t.Fatal(err)
	}
	classes := map[InventoryClass]bool{}
	for _, item := range manifest.Items {
		classes[item.InventoryClass] = true
	}
	if manifest.ItemCount != 4 || len(classes) != 4 ||
		manifest.InventorySHA256 == testDigest {
		t.Fatalf("inventory was not normalized: %#v", manifest)
	}
	document, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeInventory(document)
	if err != nil || decoded.InventorySHA256 != manifest.InventorySHA256 {
		t.Fatalf("decode inventory: %#v %v", decoded, err)
	}
	decoded.Items[0].SourcePayloadSHA256 = strings.Repeat("b", 64)
	if decoded.Validate() == nil {
		t.Fatal("tampered item payload checksum was accepted")
	}
}

func newTestItem(
	t *testing.T,
	name string,
	class InventoryClass,
	surface SourceSurface,
	contract string,
	id string,
) Item {
	t.Helper()
	item, err := NewItem(Item{
		InventoryClass: class, SourceSurface: surface, SourceName: name,
		SourceItemID: id, SourceContractVersion: contract, SourceRevision: 1,
		SourcePayloadSHA256: testDigest, TransformVersion: "workflow-async.v1",
		IdempotencyIdentity: name + "/" + id, TargetKind: "workflow.resume",
		TargetPayloadRevision: 1, TargetPayloadReference: "v1-adoption/" + id,
		TargetPayloadSHA256: testDigest, TargetRunAt: testTime,
		TargetPriority: 1, TargetMaximumAttempts: 8,
		TargetEffectClass: EffectDatabaseTransactional, CapturedAt: testTime,
	})
	if err != nil {
		t.Fatal(err)
	}
	return item
}
