//go:build integration

package integrationtest

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"testing"
	"time"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	profilebackfillpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/profilebackfill"
	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func assertRecoveryCursorEvidenceCAS(
	t *testing.T, ctx context.Context, scenario recoveryPG18Scenario,
	parentID uuid.UUID,
) {
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
	evidence := recoveryCASEvidence(parentID)
	beforeCounts := readRecoveryWriteCounts(t, ctx, scenario.db, parentID)
	beforeEvidence := readRecoveryEvidence(t, ctx, scenario.db, parentID)
	stage := "scope"
	err = scenario.runtime.coordinator.InTenant(
		ctx, recoveryCommandScope(t, ctx, scenario, "iam.profile.recover.cas"),
		func(scoped context.Context, scope platformtx.TenantScope) error {
			contract, readErr := reader.InspectParent(scoped, scope, parentID, evidence)
			if readErr != nil {
				return readErr
			}
			stage = "cursor"
			inventory, readErr := recoveryInventoryBeforeDrift(
				scoped, scope, scenario.db, reader, contract, evidence,
			)
			if readErr != nil {
				return readErr
			}
			stage = "plan"
			plan, planErr := profilebackfillapp.PlanRecovery(
				uuid.MustParse("55000000-0000-4000-8000-000000000002"),
				contract.Parent, inventory, profilebackfillapp.MaximumPageSize,
				time.Date(2026, 7, 16, 17, 30, 0, 0, time.UTC),
			)
			if planErr != nil {
				return planErr
			}
			stage = "evidence-cas"
			_, planErr = service.Start(scoped, scope,
				profilebackfillapp.RecoveryCreateCommand{Plan: plan, Evidence: evidence})
			return planErr
		},
	)
	if !errors.Is(err, profilebackfillapp.ErrConflict) {
		t.Fatalf("recovery cursor evidence CAS stage %s = %v", stage, err)
	}
	if after := readRecoveryWriteCounts(t, ctx, scenario.db, parentID); after != beforeCounts {
		t.Fatalf("recovery cursor evidence CAS wrote state = %#v -> %#v",
			beforeCounts, after)
	}
	assertRecoveryEvidenceUnchanged(t, ctx, scenario.db, parentID, beforeEvidence)
}

func recoveryCASEvidence(parentID uuid.UUID) profilebackfillapp.Evidence {
	return profilebackfillapp.Evidence{
		OperationID:          uuid.MustParse("56000000-0000-4000-8000-000000000002"),
		RequestSHA256:        sha256.Sum256([]byte("recovery-cas:" + parentID.String())),
		ActorID:              uuid.MustParse(publicationActorID),
		ManagePermissionID:   uuid.MustParse(publicationManagePerm),
		ExpectedTokenVersion: 1, ExpectedAuthorizationVersion: 1,
	}
}

func recoveryInventoryBeforeDrift(
	ctx context.Context, scope platformtx.TenantScope, db publicationDatabase,
	reader *profilebackfillpostgres.RecoveryReader,
	contract profilebackfillpostgres.RecoveryContract,
	evidence profilebackfillapp.Evidence,
) (profilebackfillapp.RecoveryInventory, error) {
	cursor, err := reader.OpenInventory(ctx, scope, contract.Parent.ID, evidence)
	if err != nil {
		return profilebackfillapp.RecoveryInventory{}, err
	}
	defer func() { _ = cursor.Close(ctx) }()
	if _, err = cursor.Fetch(context.Background()); !errors.Is(err, profilebackfillapp.ErrInvalid) {
		return profilebackfillapp.RecoveryInventory{}, fmt.Errorf(
			"unbound cursor fetch = %w", err)
	}
	rows, err := cursor.Fetch(ctx)
	if err != nil || len(rows) != 3 {
		return profilebackfillapp.RecoveryInventory{}, fmt.Errorf(
			"first cursor batch rows=%d: %w", len(rows), err)
	}
	classifier, err := iammodule.NewProfileRecoveryInventoryClassifier(
		iammodule.ProfileRecoveryInventoryInput{
			Parent: contract.Parent, Source: contract.Source, Target: contract.Target,
			Normalizer: recoveryTextNormalizer{}, Clock: func() time.Time {
				return time.Date(2026, 7, 16, 17, 15, 0, 0, time.UTC)
			},
		},
	)
	if err != nil {
		return profilebackfillapp.RecoveryInventory{}, err
	}
	resources := make([]profilebackfillapp.SourceResource, len(rows))
	var drift profilebackfillapp.SourceResource
	for index := range rows {
		resources[index] = rows[index].Resource
		if rows[index].LineageKind == "source" && drift.ID == uuid.Nil {
			drift = rows[index].Resource
		}
	}
	if drift.ID == uuid.Nil {
		return profilebackfillapp.RecoveryInventory{}, fmt.Errorf("source row is absent")
	}
	if driftErr := commitRecoveryNonProfileUpdate(ctx, db, drift); driftErr != nil {
		return profilebackfillapp.RecoveryInventory{}, fmt.Errorf(
			"non-profile drift: %w", driftErr)
	}
	end, err := cursor.Fetch(ctx)
	if err != nil || len(end) != 0 {
		return profilebackfillapp.RecoveryInventory{}, fmt.Errorf(
			"terminal cursor batch rows=%d: %w", len(end), err)
	}
	if err = cursor.Close(ctx); err != nil {
		return profilebackfillapp.RecoveryInventory{}, fmt.Errorf("close cursor: %w", err)
	}
	if err = cursor.Close(ctx); err != nil {
		return profilebackfillapp.RecoveryInventory{}, fmt.Errorf(
			"idempotent close cursor: %w", err)
	}
	return classifyRecoveryCASInventory(ctx, contract, classifier, resources)
}
