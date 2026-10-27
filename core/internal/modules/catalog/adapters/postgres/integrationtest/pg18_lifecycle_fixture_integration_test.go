//go:build integration

package integrationtest

import (
	"testing"
	"time"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/packagecontract"
	"github.com/google/uuid"
)

const catalogInstallationID = "018f2000-0000-7000-8000-000000000003"
const catalogInstallOperationID = "018f2000-0000-7000-8000-000000000004"
const catalogRetryOperationID = "018f2000-0000-7000-8000-000000000005"

func catalogInstallMutation(t *testing.T) catalogapp.StartMutation {
	t.Helper()
	packageID, err := packagecontract.ParsePackageID(catalogPackageID)
	if err != nil {
		t.Fatal(err)
	}
	componentID, err := packagecontract.ParseComponentID(catalogComponentID)
	if err != nil {
		t.Fatal(err)
	}
	version, err := packagecontract.ParseVersion("1.0.0")
	if err != nil {
		t.Fatal(err)
	}
	plan, err := catalogdomain.NewPlan(packageID, version, nil, []catalogdomain.ComponentPin{{
		ComponentID: componentID, Owner: "catalog", OwnerGeneration: 7,
		Contract: "catalog-demo.v1", ContentSHA256: catalogManifestDigest,
	}})
	if err != nil {
		t.Fatal(err)
	}
	installationID, err := packagecontract.ParseStableID(catalogInstallationID)
	if err != nil {
		t.Fatal(err)
	}
	installation, err := catalogdomain.NewInstalling(installationID, packageID)
	if err != nil {
		t.Fatal(err)
	}
	operationID, err := packagecontract.ParseStableID(catalogInstallOperationID)
	if err != nil {
		t.Fatal(err)
	}
	at := time.Date(2026, 7, 18, 18, 45, 0, 123000, time.UTC)
	operation, err := catalogdomain.NewOperation(
		operationID, catalogdomain.OperationInstall, plan, at,
	)
	if err != nil {
		t.Fatal(err)
	}
	return catalogapp.StartMutation{Installation: installation, Operation: operation,
		ActorID: uuid.MustParse(catalogActorID), IdempotencyKey: "catalog-install-proof",
		RequestSHA256: "2222222222222222222222222222222222222222222222222222222222222222",
		OccurredAt:    at, MutateInstallation: true}
}

func catalogRetryMutation(t *testing.T, install catalogapp.StartMutation) catalogapp.StartMutation {
	t.Helper()
	operationID, err := packagecontract.ParseStableID(catalogRetryOperationID)
	if err != nil {
		t.Fatal(err)
	}
	at := install.OccurredAt.Add(time.Second)
	operation, err := catalogdomain.NewOperation(
		operationID, catalogdomain.OperationRetry, install.Operation.Plan, at,
	)
	if err != nil {
		t.Fatal(err)
	}
	previous := install.Operation.ID
	operation.RetryOf = &previous
	previousUUID := uuid.MustParse(catalogInstallOperationID)
	return catalogapp.StartMutation{Installation: install.Installation, Operation: operation,
		PreviousOperation: &previousUUID, ActorID: install.ActorID,
		IdempotencyKey: "catalog-retry-proof",
		RequestSHA256:  "3333333333333333333333333333333333333333333333333333333333333333",
		OccurredAt:     at, MutateInstallation: false}
}
