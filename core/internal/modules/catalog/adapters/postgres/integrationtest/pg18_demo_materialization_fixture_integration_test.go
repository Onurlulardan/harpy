//go:build integration

package integrationtest

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"testing"
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	contract "github.com/Onurlulardan/harpy/contracts/catalogcomponent/v1"
	catalogpackagev1 "github.com/Onurlulardan/harpy/contracts/catalogpackage/v1"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

const catalogDemoPackageID = "e3000000-0000-0000-0000-000000000009"
const catalogDemoReleaseID = "019c3300-0000-7000-8000-000000000001"
const catalogDemoComponentID = "018f0000-0004-7000-8000-000000000009"
const catalogDemoInstallationID = "019c3300-0000-7000-8000-000000000040"

var catalogDemoInstallationAt = time.Date(2026, 8, 1, 11, 0, 0, 0, time.UTC)

func catalogDemoProvisionRequest(
	t *testing.T,
) contract.ProvisionRequest {
	t.Helper()
	document, err := canonicaljsonv1.CanonicalizeObject([]byte(fmt.Sprintf(
		`{"component_id":"%s","contract_version":"%s","demo_forms":[`+
			`{"legacy_form_config_id":"e4000000-0000-0000-0000-000000000561",`+
			`"rows":[{"id":"364aafc0-711c-52bd-aa33-1921d2fc2a39"}],`+
			`"slug":"project-type"}],"excluded_form_slugs":[],"module_slug":`+
			`"project","source_contract":"harpy-v1.enterprise-seed.project"}`,
		catalogDemoComponentID, catalogpackagev1.ContractCatalog,
	)), catalogpackagev1.MaximumPayloadBytes)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(document)
	return contract.ProvisionRequest{
		PackageID: catalogDemoPackageID, PackageSlug: "project",
		PackageVersion: "1.0.0", ReleaseID: catalogDemoReleaseID,
		Component: contract.Component{
			ID: catalogDemoComponentID, Owner: "catalog", Generation: 7,
			Contract:      catalogpackagev1.ContractCatalog,
			ContentSHA256: hex.EncodeToString(digest[:]),
		},
		Document:       document,
		IdempotencyKey: "catalog-demo-project-pg18-proof",
	}
}

func seedCatalogDemoInstallationBinding(
	t *testing.T, ctx context.Context, database catalogPG18,
	request contract.ProvisionRequest,
) {
	t.Helper()
	transaction, err := database.owner.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer transaction.Rollback()
	if _, err = transaction.ExecContext(ctx, `INSERT INTO catalog_a.catalog_installations(
installation_id,package_id,active_version,state,reference_epoch,row_version,updated_at)
VALUES ($1,$2,NULL,'INSTALLING',1,1,$3)`, catalogDemoInstallationID,
		request.PackageID, catalogDemoInstallationAt); err != nil {
		t.Fatalf("seed Catalog demo installation: %v", err)
	}
	if _, err = transaction.ExecContext(ctx, `INSERT INTO catalog_a.catalog_component_bindings(
installation_id,component_id,owner_module,owner_generation,owner_contract,
content_sha256,state) VALUES ($1,$2,'catalog',$3,$4,$5,'STAGED')`,
		catalogDemoInstallationID, request.Component.ID,
		request.Component.Generation, request.Component.Contract,
		request.Component.ContentSHA256); err != nil {
		t.Fatalf("seed Catalog demo component binding: %v", err)
	}
	if err = transaction.Commit(); err != nil {
		t.Fatal(err)
	}
}

type catalogDemoLifecycleOwner interface {
	Stage(context.Context, platformtx.TenantScope,
		contract.Command) (contract.Receipt, error)
	Activate(context.Context, platformtx.TenantScope,
		contract.Command) (contract.Receipt, error)
}

func assertCatalogDemoStageDenied(
	t *testing.T, ctx context.Context, database catalogPG18,
	coordinator *platformtx.Coordinator, owner catalogDemoLifecycleOwner,
	component contract.Component,
) {
	t.Helper()
	err := runCatalogDemoCommand(t, ctx, coordinator, owner, component,
		"stage", "019c3300-0000-7000-8000-000000000010", "stage-denied")
	if err == nil {
		t.Fatal("Catalog demo stage without Forms handoff was accepted")
	}
	assertCatalogDemoLifecycleState(t, ctx, database, "AVAILABLE")
}

func assertCatalogDemoLifecycleAllowed(
	t *testing.T, ctx context.Context, database catalogPG18,
	coordinator *platformtx.Coordinator, owner catalogDemoLifecycleOwner,
	component contract.Component,
) {
	t.Helper()
	if err := runCatalogDemoCommand(t, ctx, coordinator, owner, component,
		"stage", "019c3300-0000-7000-8000-000000000011", "stage-ready"); err != nil {
		t.Fatalf("stage ready Catalog demo: %v", err)
	}
	if err := runCatalogDemoCommand(t, ctx, coordinator, owner, component,
		"activate", "019c3300-0000-7000-8000-000000000012", "activate-ready"); err != nil {
		t.Fatalf("activate ready Catalog demo: %v", err)
	}
	assertCatalogDemoLifecycleState(t, ctx, database, "ACTIVE")
}

func runCatalogDemoCommand(
	t *testing.T, ctx context.Context, coordinator *platformtx.Coordinator,
	owner catalogDemoLifecycleOwner, component contract.Component,
	action, operationID, idempotencyKey string,
) error {
	t.Helper()
	return coordinator.InTenant(ctx,
		catalogLifecycleRequest(t, catalogTenantA, "catalog.operation.resume"),
		func(scoped context.Context, scope platformtx.TenantScope) error {
			command := contract.Command{
				OperationID: operationID, Component: component,
				Action: action, IdempotencyKey: idempotencyKey,
			}
			if action == "stage" {
				_, err := owner.Stage(scoped, scope, command)
				return err
			}
			_, err := owner.Activate(scoped, scope, command)
			return err
		})
}

func assertCatalogDemoLifecycleState(
	t *testing.T, ctx context.Context, database catalogPG18, want string,
) {
	t.Helper()
	var state string
	err := database.owner.QueryRowContext(ctx, `SELECT lifecycle_state
FROM catalog_a.catalog_catalog_components WHERE component_id=$1`,
		catalogDemoComponentID).Scan(&state)
	if err != nil || state != want {
		t.Fatalf("Catalog demo lifecycle state=%q want=%q err=%v", state, want, err)
	}
}
