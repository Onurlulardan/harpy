//go:build integration

package postgres

import (
	"context"
	"testing"

	asyncadoptionv1 "github.com/Onurlulardan/harpy/contracts/asyncadoption/v1"
	"github.com/jackc/pgx/v5/pgxpool"
)

func beginAndSealInventory(
	t *testing.T, ctx context.Context, adopter *pgxpool.Pool, evidence adoptionEvidence,
) {
	t.Helper()
	i := evidence.inventory
	_, err := adopter.Exec(ctx, `SELECT platform.platform_async_begin_v1_adoption_v1(
        $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)`,
		i.RunID, i.ScopeKind, i.ScopeID, i.Domain, i.GroupID,
		i.SourceGeneration, i.TargetGeneration,
		i.SourceGroupGeneration, i.TargetGroupGeneration,
		i.InventorySHA256, i.ComponentManifestSHA256,
		i.ExpectedEnqueueFences, i.ExpectedSchedulerFences,
		i.ExpectedConsumerFences, i.ItemCount, i.CapturedAt,
	)
	if err != nil {
		t.Fatalf("begin V1 adoption: %v", err)
	}
	item := evidence.item
	_, err = adopter.Exec(ctx, `SELECT platform.platform_async_record_v1_inventory_item_v1(
        $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21)`,
		i.RunID, item.InventoryClass, item.SourceSurface, item.SourceName,
		item.SourceItemID, item.SourceContractVersion, item.SourceRevision,
		item.SourcePayloadSHA256, item.TransformVersion, item.IdempotencyIdentity,
		item.TargetKind, item.TargetPayloadRevision, item.TargetPayloadReference,
		item.TargetPayloadSHA256, item.TargetRunAt, item.TargetPriority,
		item.TargetMaximumAttempts, item.TargetEffectClass, item.TargetEffectKey,
		item.CapturedAt, item.ItemSHA256,
	)
	if err != nil {
		t.Fatalf("record V1 inventory: %v", err)
	}
	if _, err := adopter.Exec(ctx,
		`SELECT platform.platform_async_seal_v1_inventory_v1($1,$2)`,
		i.RunID, i.InventorySHA256,
	); err != nil {
		t.Fatalf("seal V1 inventory: %v", err)
	}
}

func adoptInventoryItem(
	t *testing.T, ctx context.Context, adopter *pgxpool.Pool, evidence adoptionEvidence,
) {
	t.Helper()
	item, disposition := evidence.item, evidence.disposition
	_, err := adopter.Exec(ctx, `SELECT platform.platform_async_adopt_v1_inventory_item_v1(
        $1,$2,$3,$4,$5,$6,$7,$8)`,
		evidence.inventory.RunID, item.InventoryClass, item.SourceName,
		item.SourceItemID, item.ItemSHA256, adoptionJobID,
		disposition.DisposedAt, disposition.ReceiptSHA256,
	)
	if err != nil {
		t.Fatalf("adopt V1 inventory item: %v", err)
	}
}

func recordFences(
	t *testing.T, ctx context.Context, cluster *pgxpool.Pool,
	receipts []asyncadoptionv1.RestartFenceReceipt,
) {
	t.Helper()
	for _, receipt := range receipts {
		_, err := cluster.Exec(ctx, `SELECT platform.platform_async_record_v1_restart_fence_v1(
            $1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
			receipt.RunID, receipt.Kind, receipt.ComponentName,
			receipt.ComponentContractVersion, receipt.BinarySHA256,
			receipt.ConfigSHA256, receipt.RestartProbeSHA256,
			receipt.FencedAt, receipt.RestartVerifiedAt, receipt.ReceiptSHA256,
		)
		if err != nil {
			t.Fatalf("record %s restart fence: %v", receipt.Kind, err)
		}
	}
}

func recordZeroAndAccept(
	t *testing.T, ctx context.Context, cluster *pgxpool.Pool, evidence adoptionEvidence,
) {
	t.Helper()
	zero := evidence.zero
	_, err := cluster.Exec(ctx, `SELECT platform.platform_async_record_v1_zero_inventory_v1(
        $1,$2,$3,$4,$5,$6,$7,$8)`, zero.RunID, zero.InventorySHA256,
		zero.QueueCount, zero.TimerCount, zero.FutureCount, zero.RunningCount,
		zero.ObservedAt, zero.ReceiptSHA256,
	)
	if err != nil {
		t.Fatalf("record zero V1 inventory: %v", err)
	}
	a := evidence.acceptance
	_, err = cluster.Exec(ctx, `SELECT platform.platform_async_accept_v1_adoption_v1(
        $1,$2,$3,$4,$5)`, a.RunID, a.ZeroInventorySHA256,
		a.DispositionSHA256, a.AcceptedAt, a.AcceptanceSHA256,
	)
	if err != nil {
		t.Fatalf("accept V1 adoption: %v", err)
	}
}
