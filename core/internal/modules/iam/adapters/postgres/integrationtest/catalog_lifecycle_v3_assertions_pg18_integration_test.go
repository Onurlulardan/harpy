//go:build integration

package integrationtest

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"
	"time"
)

type iamCatalogAuthorizationState struct {
	clock        int64
	userVersion  int64
	activeGrants int
	mappedGrants int
	lifecycle    string
}

func assertIAMCatalogLifecycleV3(
	t *testing.T, ctx context.Context, db publicationDatabase,
) {
	t.Helper()
	digest := sha256.Sum256(iamCatalogDocumentV2(t))
	contentSHA256 := hex.EncodeToString(digest[:])
	baseline := readIAMCatalogAuthorizationState(t, ctx, db)
	if baseline.lifecycle != "ACTIVE" || baseline.activeGrants != 5 ||
		baseline.mappedGrants != 5 {
		t.Fatalf("IAM Catalog lifecycle baseline differs: %+v", baseline)
	}
	disableReceipt := "018f0000-0004-7000-8000-000000000020"
	stored, ready, err := callIAMCatalogLifecycleAction(
		ctx, db, disableReceipt, iamCatalogDisableOperation, "disable",
		contentSHA256, 0, iamCatalogLifecycleTime(10),
	)
	if err != nil || !ready || stored != disableReceipt {
		t.Fatalf("disable IAM Catalog component = %s/%t/%v", stored, ready, err)
	}
	assertIAMCatalogAuthorizationState(
		t, ctx, db, "DISABLED", 0, baseline.clock+1, baseline.userVersion+1,
	)
	stored, ready, err = callIAMCatalogLifecycleAction(
		ctx, db, "018f0000-0004-7000-8000-00000000002e",
		iamCatalogDisableOperation, "disable", contentSHA256, 0,
		iamCatalogLifecycleTime(10),
	)
	if err != nil || !ready || stored != disableReceipt {
		t.Fatalf("replay IAM Catalog disable = %s/%t/%v", stored, ready, err)
	}
	assertIAMCatalogAuthorizationState(
		t, ctx, db, "DISABLED", 0, baseline.clock+1, baseline.userVersion+1,
	)
	applyIAMCatalogLifecycleStep(
		t, ctx, db, "018f0000-0004-7000-8000-000000000022",
		iamCatalogEnableOperation, "enable", contentSHA256, 0,
		iamCatalogLifecycleTime(11),
	)
	assertIAMCatalogAuthorizationState(
		t, ctx, db, "ACTIVE", 5, baseline.clock+2, baseline.userVersion+2,
	)
	applyIAMCatalogLifecycleStep(
		t, ctx, db, "018f0000-0004-7000-8000-000000000024",
		iamCatalogRetireOperation, "retire", contentSHA256, 0,
		iamCatalogLifecycleTime(12),
	)
	assertIAMCatalogAuthorizationState(
		t, ctx, db, "RETIRED", 0, baseline.clock+3, baseline.userVersion+3,
	)
	insertIAMCatalogExternalBinding(t, ctx, db, iamCatalogLifecycleTime(13))
	_, blockers, err := callIAMCatalogBlockersV2(
		ctx, db, "018f0000-0004-7000-8000-000000000028",
		iamCatalogPurgeOperation, 3, iamCatalogLifecycleTime(14),
	)
	if err != nil ||
		!bytes.Contains(blockers, []byte("IAM_CATALOG_EXTERNAL_PERMISSION_REFERENCE")) {
		t.Fatalf("IAM Catalog live blockers = %s/%v", blockers, err)
	}
	assertIAMCatalogBlockerReceiptCount(t, ctx, db, 0)
	revokeIAMCatalogExternalBinding(t, ctx, db, iamCatalogLifecycleTime(15))
	stored, blockers, err = callIAMCatalogBlockersV2(
		ctx, db, "018f0000-0004-7000-8000-00000000002a",
		iamCatalogPurgeOperation, 3, iamCatalogLifecycleTime(16),
	)
	if err != nil || string(blockers) != "[]" ||
		stored != "018f0000-0004-7000-8000-00000000002a" {
		t.Fatalf("IAM Catalog zero blockers = %s/%s/%v", stored, blockers, err)
	}
	assertIAMCatalogBlockerReceiptCount(t, ctx, db, 1)
	_, _, err = callIAMCatalogLifecycleAction(
		ctx, db, "018f0000-0004-7000-8000-00000000002b",
		iamCatalogPurgeOperation, "purge", contentSHA256, 4,
		iamCatalogLifecycleTime(16).Add(time.Microsecond),
	)
	if err == nil {
		t.Fatal("IAM Catalog purge accepted a stale reference epoch")
	}
	purgeReceipt := "018f0000-0004-7000-8000-00000000002c"
	applyIAMCatalogLifecycleStep(
		t, ctx, db, purgeReceipt, iamCatalogPurgeOperation, "purge",
		contentSHA256, 3, iamCatalogLifecycleTime(17),
	)
	assertIAMCatalogPurged(
		t, ctx, db, baseline.clock+4, baseline.userVersion+4,
	)
	stored, ready, err = callIAMCatalogLifecycleAction(
		ctx, db, "018f0000-0004-7000-8000-00000000002d",
		iamCatalogPurgeOperation, "purge", contentSHA256, 3,
		iamCatalogLifecycleTime(18),
	)
	if err != nil || !ready || stored != purgeReceipt {
		t.Fatalf("replay IAM Catalog purge = %s/%t/%v", stored, ready, err)
	}
	assertIAMCatalogPurged(
		t, ctx, db, baseline.clock+4, baseline.userVersion+4,
	)
}

func iamCatalogLifecycleTime(minute int) time.Time {
	return time.Date(2026, 7, 30, 10, minute, 0, 0, time.UTC)
}

func applyIAMCatalogLifecycleStep(
	t *testing.T, ctx context.Context, db publicationDatabase,
	receiptID, operationID, action, contentSHA256 string,
	referenceEpoch int64, completedAt time.Time,
) {
	t.Helper()
	stored, ready, err := callIAMCatalogLifecycleAction(
		ctx, db, receiptID, operationID, action, contentSHA256,
		referenceEpoch, completedAt,
	)
	if err != nil || !ready || stored != receiptID {
		t.Fatalf("%s IAM Catalog component = %s/%t/%v",
			action, stored, ready, err)
	}
}
