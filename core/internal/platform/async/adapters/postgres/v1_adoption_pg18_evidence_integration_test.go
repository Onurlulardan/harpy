//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	asyncadoptionv1 "github.com/Onurlulardan/harpy/contracts/asyncadoption/v1"
)

const (
	adoptionRunID  = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	adoptionScope  = "11111111-1111-4111-8111-111111111111"
	adoptionJobID  = "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
	adoptionDomain = "workflow"
	adoptionGroup  = "operations"
)

type adoptionEvidence struct {
	item        asyncadoptionv1.Item
	inventory   asyncadoptionv1.Inventory
	fences      []asyncadoptionv1.RestartFenceReceipt
	disposition asyncadoptionv1.DispositionReceipt
	zero        asyncadoptionv1.ZeroInventoryReceipt
	acceptance  asyncadoptionv1.AcceptanceReceipt
}

func buildAdoptionEvidence(t *testing.T) adoptionEvidence {
	t.Helper()
	base := time.Now().UTC().Add(-10 * time.Minute).Truncate(time.Microsecond)
	stamp := func(offset time.Duration) string {
		return base.Add(offset).Format(time.RFC3339Nano)
	}
	item, err := asyncadoptionv1.NewItem(asyncadoptionv1.Item{
		InventoryClass:         asyncadoptionv1.ClassQueue,
		SourceSurface:          asyncadoptionv1.SurfaceQueue,
		SourceName:             asyncadoptionv1.V1ScheduleQueue,
		SourceItemID:           "delivery-0001",
		SourceContractVersion:  asyncadoptionv1.V1QueueContract,
		SourceRevision:         1,
		SourcePayloadSHA256:    strings.Repeat("1", 64),
		TransformVersion:       "v1.queue-to-job.v1",
		IdempotencyIdentity:    "v1/schedule.events/delivery-0001",
		TargetKind:             "workflow.resume",
		TargetPayloadRevision:  1,
		TargetPayloadReference: "payload://v1/schedule.events/delivery-0001",
		TargetPayloadSHA256:    strings.Repeat("2", 64),
		TargetRunAt:            stamp(30 * time.Minute),
		TargetPriority:         10,
		TargetMaximumAttempts:  5,
		TargetEffectClass:      asyncadoptionv1.EffectDatabaseTransactional,
		CapturedAt:             stamp(2 * time.Minute),
	})
	if err != nil {
		t.Fatal(err)
	}
	fences := make([]asyncadoptionv1.RestartFenceReceipt, 0, 3)
	for index, kind := range []asyncadoptionv1.FenceKind{
		asyncadoptionv1.FenceEnqueue,
		asyncadoptionv1.FenceScheduler,
		asyncadoptionv1.FenceConsumer,
	} {
		receipt, err := asyncadoptionv1.NewRestartFenceReceipt(
			asyncadoptionv1.RestartFenceReceipt{
				RunID: adoptionRunID, Kind: kind,
				ComponentName:            "v1." + string(kind),
				ComponentContractVersion: "v1.component.v1",
				BinarySHA256:             strings.Repeat(string(rune('3'+index)), 64),
				ConfigSHA256:             strings.Repeat(string(rune('6'+index)), 64),
				RestartProbeSHA256:       strings.Repeat(string(rune('9'-index)), 64),
				FencedAt:                 stamp(0), RestartVerifiedAt: stamp(time.Minute),
			},
		)
		if err != nil {
			t.Fatal(err)
		}
		fences = append(fences, receipt)
	}
	manifest, err := asyncadoptionv1.ComponentManifestSHA256(fences)
	if err != nil {
		t.Fatal(err)
	}
	inventory, err := asyncadoptionv1.NewInventory(asyncadoptionv1.Inventory{
		RunID: adoptionRunID, ScopeKind: "tenant", ScopeID: adoptionScope,
		Domain: adoptionDomain, GroupID: adoptionGroup,
		SourceGeneration: 1, TargetGeneration: 2,
		SourceGroupGeneration: 1, TargetGroupGeneration: 2,
		ComponentManifestSHA256: manifest,
		ExpectedEnqueueFences:   1, ExpectedSchedulerFences: 1,
		ExpectedConsumerFences: 1, CapturedAt: stamp(3 * time.Minute),
		Items: []asyncadoptionv1.Item{item},
	})
	if err != nil {
		t.Fatal(err)
	}
	jobID := adoptionJobID
	disposition, err := asyncadoptionv1.NewDispositionReceipt(
		asyncadoptionv1.DispositionReceipt{
			RunID: adoptionRunID, InventoryClass: item.InventoryClass,
			SourceName: item.SourceName, SourceItemID: item.SourceItemID,
			ItemSHA256: item.ItemSHA256, IdempotencyIdentity: item.IdempotencyIdentity,
			Disposition: asyncadoptionv1.DispositionAdopted,
			TargetJobID: &jobID, DisposedAt: stamp(4 * time.Minute),
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	zero, err := asyncadoptionv1.NewZeroInventoryReceipt(
		asyncadoptionv1.ZeroInventoryReceipt{
			RunID: adoptionRunID, InventorySHA256: inventory.InventorySHA256,
			ObservedAt: stamp(5 * time.Minute),
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	acceptance, err := asyncadoptionv1.NewAcceptanceReceipt(
		asyncadoptionv1.AcceptanceReceipt{
			RunID: adoptionRunID, InventorySHA256: inventory.InventorySHA256,
			ComponentManifestSHA256: manifest, ZeroInventorySHA256: zero.ReceiptSHA256,
			AcceptedAt:          stamp(6 * time.Minute),
			DispositionReceipts: []string{disposition.ReceiptSHA256},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	return adoptionEvidence{item, inventory, fences, disposition, zero, acceptance}
}

func refreshAcceptanceAtDatabaseClock(
	t *testing.T, ctx context.Context, database *sql.DB, evidence adoptionEvidence,
) adoptionEvidence {
	t.Helper()
	var acceptedAt time.Time
	if err := database.QueryRowContext(ctx, "SELECT pg_catalog.clock_timestamp()").
		Scan(&acceptedAt); err != nil {
		t.Fatal(err)
	}
	acceptance, err := asyncadoptionv1.NewAcceptanceReceipt(
		asyncadoptionv1.AcceptanceReceipt{
			RunID:                   evidence.inventory.RunID,
			InventorySHA256:         evidence.inventory.InventorySHA256,
			ComponentManifestSHA256: evidence.inventory.ComponentManifestSHA256,
			ZeroInventorySHA256:     evidence.zero.ReceiptSHA256,
			AcceptedAt:              acceptedAt.UTC().Format(time.RFC3339Nano),
			DispositionReceipts:     []string{evidence.disposition.ReceiptSHA256},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	evidence.acceptance = acceptance
	return evidence
}
