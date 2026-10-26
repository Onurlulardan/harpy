//go:build integration

package integrationtest

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	contract "github.com/Onurlulardan/harpy/contracts/catalogcomponent/v1"
	catalogpackagev1 "github.com/Onurlulardan/harpy/contracts/catalogpackage/v1"
	catalogmodule "github.com/Onurlulardan/harpy/core/internal/modules/catalog"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/demomaterialization"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestCatalogDemoMaterializationWaitsForDurableFormsHandoffOnPostgreSQL184(
	t *testing.T,
) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			testCatalogDemoMaterializationWaitsForDurableFormsHandoff(t, origin)
		})
	}
}

func testCatalogDemoMaterializationWaitsForDurableFormsHandoff(
	t *testing.T, origin string,
) {
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
	defer cancel()
	database := startCatalogPG18(t, ctx, "demo_handoff_"+origin)
	prepareCatalogPG18(t, ctx, database, origin)
	coordinator := openCatalogCoordinator(t, ctx, database)
	owner, err := catalogmodule.NewCatalogComponentOwner()
	if err != nil {
		t.Fatal(err)
	}
	request := catalogDemoProvisionRequest(t)
	seedCatalogDemoInstallationBinding(t, ctx, database, request)

	pending, err := runCatalogDemoProvision(
		t, ctx, coordinator, owner, request,
	)
	if err != nil || pending.Ready || pending.RetryAt.IsZero() ||
		!pending.RetryAt.After(pending.CompletedAt) {
		t.Fatalf("initial Catalog demo receipt=%+v err=%v", pending, err)
	}
	replay, err := runCatalogDemoProvision(t, ctx, coordinator, owner, request)
	if err != nil || replay.ID != pending.ID ||
		!replay.CompletedAt.Equal(pending.CompletedAt) ||
		!replay.RetryAt.Equal(pending.RetryAt) {
		t.Fatalf("exact pending replay=%+v err=%v", replay, err)
	}
	assertCatalogDemoStageDenied(
		t, ctx, database, coordinator, owner, request.Component,
	)

	conflict := request.Clone()
	conflict.ReleaseID = "019c3300-0000-7000-8000-000000000002"
	if _, err = runCatalogDemoProvision(t, ctx, coordinator, owner, conflict); !errors.Is(err, demomaterialization.ErrConflict) {
		t.Fatalf("conflicting replay error=%v", err)
	}
	assertCatalogDemoProjectionCount(t, ctx, database, 1, "1.0.0", false)
	insertCatalogDemoHandoff(t, ctx, database, pending.ID,
		"019c3300-0000-7000-8000-000000000020",
		"019c3300-0000-7000-8000-000000000021")

	ready, err := runCatalogDemoProvision(t, ctx, coordinator, owner, request)
	if err != nil || !ready.Ready || !ready.RetryAt.IsZero() ||
		ready.ID != pending.ID {
		t.Fatalf("ready Catalog demo replay=%+v err=%v", ready, err)
	}
	const v1Operation = "019c3300-0000-7000-8000-000000000011"
	if err = runCatalogDemoCommand(t, ctx, coordinator, owner,
		request.Component, "stage", v1Operation, "stage-ready"); err != nil {
		t.Fatalf("stage v1 Catalog demo: %v", err)
	}
	assertCatalogDemoActivationEvidence(t, ctx, database, pending.ID)
	if err = runCatalogDemoCommand(t, ctx, coordinator, owner,
		request.Component, "activate", v1Operation, "activate-ready"); err != nil {
		t.Fatalf("activate v1 Catalog demo: %v", err)
	}
	assertCatalogDemoProjectionCount(t, ctx, database, 1, "1.0.0", true)
	assertCatalogDemoActiveHead(
		t, ctx, database, "1.0.0", request.Component.ContentSHA256,
	)

	const v2Operation = "019c3300-0000-7000-8000-000000000030"
	upgrade := catalogDemoUpgradeRequest(t, request, v2Operation)
	pendingV2, err := runCatalogDemoProvision(
		t, ctx, coordinator, owner, upgrade,
	)
	if err != nil || pendingV2.Ready || pendingV2.ID == pending.ID {
		t.Fatalf("pending v2 Catalog demo=%+v err=%v", pendingV2, err)
	}
	replayV2, err := runCatalogDemoProvision(t, ctx, coordinator, owner, upgrade)
	if err != nil || replayV2.ID != pendingV2.ID {
		t.Fatalf("pending v2 replay=%+v err=%v", replayV2, err)
	}
	assertCatalogDemoActiveHead(
		t, ctx, database, "1.0.0", request.Component.ContentSHA256,
	)
	conflictV2 := upgrade.Clone()
	conflictV2.ReleaseID = "019c3300-0000-7000-8000-000000000032"
	if _, err = runCatalogDemoProvision(t, ctx, coordinator, owner,
		conflictV2); !errors.Is(err, demomaterialization.ErrConflict) {
		t.Fatalf("conflicting v2 replay error=%v", err)
	}
	insertCatalogDemoHandoff(t, ctx, database, pendingV2.ID,
		"019c3300-0000-7000-8000-000000000033",
		"019c3300-0000-7000-8000-000000000034")
	readyV2, err := runCatalogDemoProvision(t, ctx, coordinator, owner, upgrade)
	if err != nil || !readyV2.Ready || readyV2.ID != pendingV2.ID {
		t.Fatalf("ready v2 Catalog demo=%+v err=%v", readyV2, err)
	}
	if err = runCatalogDemoCommand(t, ctx, coordinator, owner,
		upgrade.Component, "stage", v2Operation, "stage-v2"); err != nil {
		t.Fatalf("stage v2 Catalog demo: %v", err)
	}
	assertCatalogDemoActiveHead(
		t, ctx, database, "1.0.0", request.Component.ContentSHA256,
	)
	if err = runCatalogDemoCommand(t, ctx, coordinator, owner,
		upgrade.Component, "activate", v2Operation, "activate-v2"); err != nil {
		t.Fatalf("activate v2 Catalog demo: %v", err)
	}
	assertCatalogDemoActiveHead(
		t, ctx, database, "2.0.0", upgrade.Component.ContentSHA256,
	)
	assertCatalogDemoProjectionCount(t, ctx, database, 2, "2.0.0", true)
	assertCatalogDemoEvidenceVersions(t, ctx, database)
	assertCatalogDemoRuntimeCannotForgeHandoff(t, ctx, database)
}

func catalogDemoUpgradeRequest(
	t *testing.T, original contract.ProvisionRequest, operationID string,
) contract.ProvisionRequest {
	t.Helper()
	document := bytes.Replace(
		original.Document,
		[]byte(`{"id":"364aafc0-711c-52bd-aa33-1921d2fc2a39"}`),
		[]byte(`{"id":"364aafc0-711c-52bd-aa33-1921d2fc2a39","name":"v2"}`),
		1,
	)
	document, err := canonicaljsonv1.CanonicalizeObject(
		document, catalogpackagev1.MaximumPayloadBytes,
	)
	if err != nil || bytes.Equal(document, original.Document) {
		t.Fatalf("canonicalize v2 Catalog demo: %v", err)
	}
	digest := sha256.Sum256(document)
	request := original.Clone()
	request.PackageVersion = "2.0.0"
	request.ReleaseID = "019c3300-0000-7000-8000-000000000031"
	request.Document = document
	request.Component.ContentSHA256 = hex.EncodeToString(digest[:])
	request.IdempotencyKey = operationID + ":" + request.Component.ID + ":provision"
	return request
}

func runCatalogDemoProvision(
	t *testing.T,
	ctx context.Context,
	coordinator *platformtx.Coordinator,
	owner interface {
		Provision(context.Context, platformtx.TenantScope,
			contract.ProvisionRequest) (contract.ProvisionReceipt, error)
	},
	request contract.ProvisionRequest,
) (contract.ProvisionReceipt, error) {
	var receipt contract.ProvisionReceipt
	err := coordinator.InTenant(
		ctx, catalogLifecycleRequest(t, catalogTenantA,
			"catalog.operation.resume"),
		func(scoped context.Context, scope platformtx.TenantScope) error {
			var provisionErr error
			receipt, provisionErr = owner.Provision(scoped, scope, request)
			return provisionErr
		},
	)
	return receipt, err
}
