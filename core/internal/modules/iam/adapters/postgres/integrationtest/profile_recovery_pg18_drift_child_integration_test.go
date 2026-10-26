//go:build integration

package integrationtest

import (
	"context"
	"crypto/sha256"
	"testing"
	"time"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	profilebackfillpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/profilebackfill"
	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func createRecoveryDriftChild(
	t *testing.T, ctx context.Context, scenario recoveryPG18Scenario,
	parentID, planID, operationID uuid.UUID,
) uuid.UUID {
	t.Helper()
	reader, err := profilebackfillpostgres.NewRecoveryReader()
	if err != nil {
		t.Fatal(err)
	}
	store, err := profilebackfillpostgres.NewRecoveryStore()
	if err != nil {
		t.Fatal(err)
	}
	catalogs, err := platformasync.LoadWriteCatalogs()
	if err != nil {
		t.Fatal(err)
	}
	service, err := profilebackfillapp.NewRecoveryService(
		store, recoveryPGEffects{catalogs: catalogs},
	)
	if err != nil {
		t.Fatal(err)
	}
	evidence := profilebackfillapp.Evidence{
		OperationID:          operationID,
		RequestSHA256:        sha256.Sum256([]byte("drift-recovery:" + parentID.String())),
		ActorID:              uuid.MustParse(publicationActorID),
		ManagePermissionID:   uuid.MustParse(publicationManagePerm),
		ExpectedTokenVersion: 1, ExpectedAuthorizationVersion: 1,
	}
	request := recoveryCommandScope(t, ctx, scenario, "iam.profile.recover.drift")
	var result profilebackfillapp.RecoveryResult
	stage := "scope"
	err = scenario.runtime.coordinator.InTenant(
		ctx, request, func(scoped context.Context, scope platformtx.TenantScope) error {
			stage = "parent"
			contract, readErr := reader.InspectParent(scoped, scope, parentID, evidence)
			if readErr != nil {
				return readErr
			}
			stage = "inventory"
			rows, readErr := drainRecoveryInventory(
				scoped, reader, scope, parentID, evidence,
			)
			if readErr != nil {
				return readErr
			}
			resources := make([]profilebackfillapp.SourceResource, 0, len(rows))
			for _, row := range rows {
				resources = append(resources, row.Resource)
			}
			stage = "classifier"
			classifier, classifyErr := iammodule.NewProfileRecoveryInventoryClassifier(
				iammodule.ProfileRecoveryInventoryInput{
					Parent: contract.Parent, Source: contract.Source,
					Target: contract.Target,
					Normalizer: recoveryTextNormalizer{
						drift: true, calls: make(map[string]int),
					},
					Clock: func() time.Time {
						return time.Date(2026, 7, 16, 18, 0, 0, 0, time.UTC)
					},
				},
			)
			if classifyErr != nil {
				return classifyErr
			}
			schemas := map[uuid.UUID]profileapp.Schema{
				contract.Source.ID(): contract.Source,
				contract.Target.ID(): contract.Target,
			}
			for index := range resources {
				stage = "append:" + resources[index].ID.String()
				classifyErr = classifier.AppendBatch(
					scoped, schemas, resources[index:index+1],
				)
				if classifyErr != nil {
					return classifyErr
				}
			}
			stage = "finish"
			inventory, classifyErr := classifier.Finish(scoped)
			if classifyErr != nil {
				return classifyErr
			}
			stage = "plan"
			plan, planErr := profilebackfillapp.PlanRecovery(
				planID, contract.Parent, inventory, profilebackfillapp.MaximumPageSize,
				time.Date(2026, 7, 16, 18, 0, 0, 0, time.UTC),
			)
			if planErr != nil {
				return planErr
			}
			stage = "start"
			result, planErr = service.Start(
				scoped, scope, profilebackfillapp.RecoveryCreateCommand{
					Plan: plan, Evidence: evidence,
				},
			)
			return planErr
		},
	)
	if err != nil || result.PlanID != planID || result.Replayed ||
		result.State != profilebackfillapp.RecoveryPlanned {
		t.Fatalf("create drift recovery child stage %s = %#v/%v", stage, result, err)
	}
	return result.PlanID
}

func recoveryCommandScope(
	t *testing.T, ctx context.Context, scenario recoveryPG18Scenario,
	operation string,
) platformtx.ScopeRequest {
	t.Helper()
	tenantID := uuid.MustParse(scenario.db.tenant.id)
	generations, err := scenario.runtime.pool.ResolveSnapshotDomainGenerations(
		ctx, tenantID, []string{"iam", "system_resource_layout"},
	)
	if err != nil {
		t.Fatal(err)
	}
	request, err := platformtx.AdmitCommandScope(
		uuid.MustParse(publicationActorID), tenantID, operation,
		"iam_system_resources", generations["iam"], generations,
	)
	if err != nil {
		t.Fatal(err)
	}
	return request
}
