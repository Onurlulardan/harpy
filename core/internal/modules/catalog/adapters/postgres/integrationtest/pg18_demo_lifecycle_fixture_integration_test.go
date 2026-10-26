//go:build integration

package integrationtest

import (
	"context"
	"database/sql"
	"testing"
	"time"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
	"github.com/google/uuid"
)

const catalogDemoLifecycleInstallationID = "019c3400-0000-7000-8000-000000000001"
const catalogDemoLifecyclePackageID = "e3400000-0000-0000-0000-000000000001"
const catalogDemoLifecycleComponentID = "019c3400-0000-7000-8000-000000000002"
const catalogDemoLifecycleReleaseID = "019c3400-0000-7000-8000-000000000003"
const catalogDemoLifecycleProvisionID = "019c3400-0000-7000-8000-000000000004"
const catalogDemoLifecycleHandoffID = "019c3400-0000-7000-8000-000000000005"
const catalogDemoLifecycleSourceRunID = "019c3400-0000-7000-8000-000000000006"
const catalogDemoLifecycleOperationID = "019c3400-0000-7000-8000-000000000007"
const catalogDemoLifecycleReceiptID = "019c3400-0000-7000-8000-000000000008"
const catalogDemoLifecycleRecordID = "364aafc0-711c-52bd-aa33-1921d2fc2a39"
const catalogDemoLifecycleContent = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
const catalogDemoLifecyclePlan = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"

var catalogDemoLifecycleAt = time.Date(2026, 8, 1, 12, 0, 0, 123000, time.UTC)

func seedCatalogDemoLifecycleFixture(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	tx, err := database.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()
	statements := []string{
		`INSERT INTO catalog_a.catalog_installations VALUES
($1,$2,'1.0.0','ACTIVE',1,1,$3)`,
		`INSERT INTO catalog_a.catalog_component_bindings VALUES
($1,$2,'catalog',7,'catalog.demo-data.v1',$3,'ACTIVE')`,
		`INSERT INTO catalog_a.catalog_catalog_components VALUES
($1,'DEMO_DATA','catalog.demo-data.v1',$2,7,'v2','ACTIVE',$3)`,
		`INSERT INTO catalog_a.catalog_demo_payloads VALUES
($1,$2,'project','1.0.0',$3,'project','test.demo.v1',decode(repeat('11',32),'hex'),1,1,$4,$5)`,
		`INSERT INTO catalog_a.catalog_demo_provisions VALUES
($1,$2,'project','1.0.0',$3,$4,7,'catalog.demo-data.v1',$5,
 'demo-lifecycle-provision',decode(repeat('22',32),'hex'),true,NULL,$6,$6,$6)`,
		`INSERT INTO catalog_a.catalog_demo_handoffs VALUES
($1,$2,$3,$4,'1.0.0',$5,$6,$7,decode(repeat('22',32),'hex'),1,1,
 decode(repeat('33',32),'hex'),$8)`,
		`INSERT INTO catalog_a.catalog_demo_component_heads VALUES
($1,$2,'1.0.0',$3,$4,$5,$6,$7,$8)`,
	}
	arguments := [][]any{
		{catalogDemoLifecycleInstallationID, catalogDemoLifecyclePackageID, catalogDemoLifecycleAt},
		{catalogDemoLifecycleInstallationID, catalogDemoLifecycleComponentID, catalogDemoLifecycleContent},
		{catalogDemoLifecycleComponentID, catalogDemoLifecycleContent, catalogDemoLifecycleAt},
		{catalogDemoLifecycleComponentID, catalogDemoLifecyclePackageID, catalogDemoLifecycleReleaseID,
			[]byte(`{"demo":true}`), catalogDemoLifecycleAt},
		{catalogDemoLifecycleProvisionID, catalogDemoLifecyclePackageID, catalogDemoLifecycleReleaseID,
			catalogDemoLifecycleComponentID, catalogDemoLifecycleContent, catalogDemoLifecycleAt},
		{catalogDemoLifecycleHandoffID, catalogDemoLifecycleProvisionID, catalogDemoLifecycleComponentID,
			catalogDemoLifecyclePackageID, catalogDemoLifecycleReleaseID, catalogDemoLifecycleSourceRunID,
			catalogDemoLifecycleContent, catalogDemoLifecycleAt},
		{catalogDemoLifecycleComponentID, catalogDemoLifecyclePackageID, catalogDemoLifecycleReleaseID,
			catalogDemoLifecycleProvisionID, catalogDemoLifecycleHandoffID,
			catalogDemoLifecycleSourceRunID, catalogDemoLifecycleContent, catalogDemoLifecycleAt},
	}
	for index := range statements {
		if _, err = tx.ExecContext(ctx, statements[index], arguments[index]...); err != nil {
			t.Fatalf("seed Catalog demo lifecycle row %d: %v", index, err)
		}
	}
	if err = tx.Commit(); err != nil {
		t.Fatal(err)
	}
}

func catalogDemoRemoveMutation(t *testing.T) catalogapp.StartMutation {
	t.Helper()
	installationID, _ := packagecontract.ParseStableID(catalogDemoLifecycleInstallationID)
	packageID, _ := packagecontract.ParsePackageID(catalogDemoLifecyclePackageID)
	componentID, _ := packagecontract.ParseComponentID(catalogDemoLifecycleComponentID)
	version, _ := packagecontract.ParseVersion("1.0.0")
	plan, err := catalogdomain.NewPlan(packageID, version, nil, []catalogdomain.ComponentPin{{
		ComponentID: componentID, Owner: "catalog", OwnerGeneration: 7,
		Contract: "catalog.demo-data.v1", ContentSHA256: catalogDemoLifecycleContent,
	}})
	if err != nil {
		t.Fatal(err)
	}
	operationID, _ := packagecontract.ParseStableID(catalogDemoLifecycleOperationID)
	operation, err := catalogdomain.NewOperation(
		operationID, catalogdomain.OperationDemoRemove, plan, catalogDemoLifecycleAt.Add(time.Second),
	)
	if err != nil {
		t.Fatal(err)
	}
	return catalogapp.StartMutation{
		Installation: catalogdomain.Installation{ID: installationID, PackageID: packageID,
			ActiveVersion: version, State: catalogdomain.InstallationActive,
			ReferenceEpoch: 1, RowVersion: 1},
		Operation: operation, ActorID: uuid.MustParse(catalogActorID),
		IdempotencyKey: "catalog-demo-remove-proof",
		RequestSHA256:  "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc",
		OccurredAt:     operation.StartedAt,
	}
}

func catalogDemoStoredPlan() catalogapp.DemoOperationPlan {
	return catalogapp.DemoOperationPlan{
		OperationID:    uuid.MustParse(catalogDemoLifecycleOperationID),
		InstallationID: uuid.MustParse(catalogDemoLifecycleInstallationID),
		Action:         catalogapp.DemoActionRemove, ExpectedPlanSHA256: catalogDemoLifecyclePlan,
		Components: []catalogapp.DemoRemovalComponent{{
			ComponentID: uuid.MustParse(catalogDemoLifecycleComponentID),
			SourceRunID: uuid.MustParse(catalogDemoLifecycleSourceRunID),
			RecordIDs:   []uuid.UUID{uuid.MustParse(catalogDemoLifecycleRecordID)},
		}},
		CreatedAt: catalogDemoLifecycleAt.Add(time.Second),
	}
}

func catalogDemoRemoveResult() catalogapp.DemoResultMutation {
	return catalogapp.DemoResultMutation{
		OperationID:    uuid.MustParse(catalogDemoLifecycleOperationID),
		ReceiptID:      uuid.MustParse(catalogDemoLifecycleReceiptID),
		InstallationID: uuid.MustParse(catalogDemoLifecycleInstallationID),
		ComponentID:    uuid.MustParse(catalogDemoLifecycleComponentID),
		Action:         catalogapp.DemoActionRemove, ContentSHA256: catalogDemoLifecycleContent,
		IdempotencyKey: "catalog-demo-remove-result", SourceRunID: uuid.MustParse(catalogDemoLifecycleSourceRunID),
		RecordIDs:        []uuid.UUID{uuid.MustParse(catalogDemoLifecycleRecordID)},
		ExpectedRevision: 1, CompletedAt: catalogDemoLifecycleAt.Add(2 * time.Second),
	}
}
