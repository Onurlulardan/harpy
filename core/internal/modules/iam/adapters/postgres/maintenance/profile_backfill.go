package maintenance

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"fmt"
	"strconv"

	maintenanceapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/maintenance"
	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/google/uuid"
)

func (store *store) executeProfileBackfill(
	ctx context.Context,
	claim maintenanceapp.Claim,
	nextID uuid.UUID,
) (maintenanceapp.Outcome, error) {
	_ = nextID
	if ctx == nil || ctx.Err() != nil || store == nil || !store.runtime.usable() {
		return maintenanceapp.Outcome{}, ErrRuntimeInvalid
	}
	if receipt, found, err := store.readReceipt(ctx, claim); found || err != nil {
		return receipt, err
	}
	page, err := store.readProfilePage(ctx, claim)
	if err != nil {
		return maintenanceapp.Outcome{}, err
	}
	if err := store.verifyProfilePageEvidence(ctx, claim, page); err != nil {
		return maintenanceapp.Outcome{}, err
	}
	sourceSchema, err := store.loadProfileSchema(
		ctx, claim, page.planID, page.sourceSchemaID,
	)
	if err != nil {
		return maintenanceapp.Outcome{}, err
	}
	targetSchema, err := store.loadProfileSchema(
		ctx, claim, page.planID, page.targetSchemaID,
	)
	if err != nil {
		return maintenanceapp.Outcome{}, err
	}
	mappings, err := store.readProfileMappings(ctx, claim, page.planID)
	if err != nil {
		return maintenanceapp.Outcome{}, ErrRuntimeInvalid
	}
	migrationMappings, err := compileProfileMappings(targetSchema, mappings)
	if err != nil {
		return maintenanceapp.Outcome{}, err
	}
	return store.convertProfilePage(
		ctx, claim, page, sourceSchema, targetSchema, migrationMappings,
	)
}

func (store *store) convertProfilePage(
	ctx context.Context,
	claim maintenanceapp.Claim,
	page profilePage,
	sourceSchema, targetSchema profileapp.Schema,
	mappings []profileapp.MigrationMapping,
) (maintenanceapp.Outcome, error) {
	values, err := claimValues(claim)
	if err != nil {
		return maintenanceapp.Outcome{}, err
	}
	resources, err := store.readProfilePageResources(ctx, values, page)
	if err != nil {
		return maintenanceapp.Outcome{}, err
	}
	defer func() {
		for index := range resources {
			resources[index].wipe()
		}
	}()
	schemas := map[uuid.UUID]profileapp.Schema{sourceSchema.ID(): sourceSchema}
	selectedMappings := make(map[uuid.UUID][]profileapp.MigrationMapping)
	sourceDigest, resultDigest := sha256.New(), sha256.New()
	var scanned, converted int64
	for index := range resources {
		resource := &resources[index]
		writeDirectoryDigest(sourceDigest, []byte(resource.id.String()),
			[]byte(resource.actualSchemaID.String()),
			[]byte(strconv.FormatInt(resource.revision, 10)), resource.sha)
		if resource.preconverted {
			writeDirectoryDigest(resultDigest,
				[]byte(resource.id.String()), resource.plannedTarget)
			converted++
			scanned++
			continue
		}
		actualSchema, found := schemas[resource.actualSchemaID]
		if !found {
			actualSchema, err = store.loadProfileSchema(
				ctx, claim, page.planID, resource.actualSchemaID,
			)
			if err != nil {
				return maintenanceapp.Outcome{}, err
			}
			schemas[resource.actualSchemaID] = actualSchema
		}
		actualMappings, found := selectedMappings[resource.actualSchemaID]
		if !found {
			actualMappings, err = profileapp.SelectMigrationMappingsForSource(
				actualSchema, targetSchema, mappings,
			)
			if err != nil {
				return maintenanceapp.Outcome{}, ErrRuntimeInvalid
			}
			selectedMappings[resource.actualSchemaID] = actualMappings
		}
		document, conversionErr := convertProfileDocument(
			actualSchema, targetSchema, actualMappings, *resource,
		)
		if conversionErr != nil {
			if errors.Is(conversionErr, profileapp.ErrSchemaInvalid) {
				return maintenanceapp.Outcome{}, ErrRuntimeInvalid
			}
			evidence, recordErr := store.recordProfileFailure(
				ctx, values, page, *resource, "document_invalid",
			)
			if recordErr != nil {
				resource.wipe()
				return maintenanceapp.Outcome{}, recordErr
			}
			writeDirectoryDigest(resultDigest, []byte(resource.id.String()), evidence)
			clear(evidence)
		} else {
			if err := store.applyProfileResource(
				ctx, values, page, *resource, document,
			); err != nil {
				resource.wipe()
				return maintenanceapp.Outcome{}, err
			}
			digest := document.SHA256()
			writeDirectoryDigest(resultDigest, []byte(resource.id.String()), digest[:])
			converted++
		}
		scanned++
	}
	sourceSHA := sourceDigest.Sum(nil)
	defer clear(sourceSHA)
	if scanned != int64(page.expectedCount) || !hmac.Equal(sourceSHA, page.sourceSHA) {
		return maintenanceapp.Outcome{}, ErrRuntimeInvalid
	}
	resultSHA := resultDigest.Sum(nil)
	defer clear(resultSHA)
	arguments := finishProfileArguments(values, page, scanned, converted, resultSHA)
	outcome, err := scanOutcome(
		store.database.QueryRow(ctx, finishProfileBackfillSQL, arguments...), claim.Kind,
	)
	if err != nil {
		return maintenanceapp.Outcome{}, fmt.Errorf("finish profile backfill: %w", err)
	}
	return outcome, nil
}
