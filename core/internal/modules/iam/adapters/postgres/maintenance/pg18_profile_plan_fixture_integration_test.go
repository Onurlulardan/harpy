//go:build integration

package maintenance

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/google/uuid"
)

type maintenancePlanFixture struct {
	tenant       maintenancePG18Tenant
	planID       uuid.UUID
	targetSchema profileapp.Schema
	jobs         [2]uuid.UUID
}

func createMaintenancePG18ProfilePlans(
	t *testing.T, ctx context.Context, database *sql.DB,
) []maintenancePlanFixture {
	t.Helper()
	result := make([]maintenancePlanFixture, 0, 2)
	for index, tenant := range maintenancePG18Tenants {
		planID := []string{maintenancePlanA, maintenancePlanB}[index]
		targetID := []string{maintenanceTargetSchemaA, maintenanceTargetSchemaB}[index]
		jobs := [2]uuid.UUID{
			uuid.MustParse(fmt.Sprintf("019c6a7e-7c00-7000-8000-0000000003%d1", index)),
			uuid.MustParse(fmt.Sprintf("019c6a7e-7c00-7000-8000-0000000003%d2", index)),
		}
		target := newMaintenanceTargetSchema(t, targetID, index)
		document := maintenancePlanDocument(
			t, tenant, uuid.MustParse(planID), target, jobs, index,
		)
		digest := sha256.Sum256(document)
		request := sha256.Sum256([]byte("profile-plan-" + tenant.id))
		createMaintenancePG18ProfilePlan(t, ctx, database, tenant,
			uuid.MustParse(planID), document, digest, request, index)
		result = append(result, maintenancePlanFixture{
			tenant: tenant, planID: uuid.MustParse(planID),
			targetSchema: target, jobs: jobs,
		})
	}
	return result
}

func maintenancePlanDocument(
	t *testing.T, tenant maintenancePG18Tenant, planID uuid.UUID,
	target profileapp.Schema, jobs [2]uuid.UUID, tenantIndex int,
) []byte {
	t.Helper()
	sources, targets := make([][32]byte, 3), make([][32]byte, 3)
	for index := range maintenanceProfileUsers {
		_, sources[index] = maintenanceMixedSourceDocument(tenantIndex, index)
		_, targets[index] = maintenanceMixedTargetDocument(tenantIndex, index)
	}
	pageOneSource := maintenancePageDigest([][]byte{
		[]byte(maintenanceUserOne), []byte(maintenanceUserSchemaID),
		[]byte("3"), sources[0][:],
	})
	pageTwoSource := maintenancePageDigest(
		[][]byte{[]byte(maintenanceUserTwo), []byte(maintenanceOptionalSchemaID),
			[]byte("4"), sources[1][:]},
		[][]byte{[]byte(maintenanceUserThree), []byte(maintenanceUserSchemaID),
			[]byte("3"), sources[2][:]},
	)
	pageOneTarget := maintenancePageDigest([][]byte{
		[]byte(maintenanceUserOne), targets[0][:],
	})
	pageTwoTarget := maintenancePageDigest(
		[][]byte{[]byte(maintenanceUserTwo), targets[1][:]},
		[][]byte{[]byte(maintenanceUserThree), targets[2][:]},
	)
	sourceAggregate := sha256.Sum256(append(pageOneSource[:], pageTwoSource[:]...))
	initial := sha256.Sum256(nil)
	rollingOne := sha256.Sum256(append(initial[:], pageOneTarget[:]...))
	rollingTwo := sha256.Sum256(append(rollingOne[:], pageTwoTarget[:]...))
	field, _ := target.Field(uuid.MustParse(maintenanceTargetFieldID))
	semantic := field.SemanticSHA256()
	optional, _ := target.Field(uuid.MustParse(maintenanceOptionalFieldID))
	optionalSemantic := optional.SemanticSHA256()
	definition := target.DefinitionSHA256()
	runAt := time.Date(2026, 7, 16, 13, 0, 0, 0, time.UTC)
	pages := []map[string]any{
		maintenancePlanPage(tenant, planID, jobs[0], 1, nil,
			maintenanceUserOne, 1, pageOneSource, pageOneTarget, runAt),
		maintenancePlanPage(tenant, planID, jobs[1], 2,
			&maintenanceProfileUsers[0], maintenanceUserThree, 2,
			pageTwoSource, pageTwoTarget, runAt.Add(time.Minute)),
	}
	wire := map[string]any{
		"id": planID.String(), "resource_kind": "user",
		"source_schema_id": maintenanceOptionalSchemaID,
		"target_schema_id": target.ID().String(), "target_version": 3,
		"target_definition":        json.RawMessage(target.DefinitionJSON()),
		"target_definition_sha256": hex.EncodeToString(definition[:]),
		"converter_version":        "typed_map_v1", "source_count": 3,
		"expected_source_sha256": hex.EncodeToString(sourceAggregate[:]),
		"expected_target_sha256": hex.EncodeToString(rollingTwo[:]),
		"page_size":              2, "initial_rolling_sha256": hex.EncodeToString(initial[:]),
		"mappings": []map[string]any{
			{"ordinal": 1, "source_field_id": maintenanceProfileField,
				"target_field_id": maintenanceTargetFieldID,
				"converter_kind":  "copy_normalize_v1",
				"semantic_sha256": hex.EncodeToString(semantic[:])},
			{"ordinal": 2, "source_field_id": maintenanceOptionalFieldID,
				"target_field_id": maintenanceOptionalFieldID,
				"converter_kind":  "copy_v1",
				"semantic_sha256": hex.EncodeToString(optionalSemantic[:])},
		},
		"pages": pages,
	}
	document, err := json.Marshal(wire)
	if err != nil {
		t.Fatal(err)
	}
	return document
}

func maintenancePlanPage(
	tenant maintenancePG18Tenant, planID, jobID uuid.UUID, number int,
	after *string, through string, count int,
	source, target [32]byte, runAt time.Time,
) map[string]any {
	reference := fmt.Sprintf("urn:harpy:iam:profile-backfill:%s:%s:%d:v1",
		tenant.id, planID, number)
	referenceSHA := sha256.Sum256([]byte(reference))
	return map[string]any{
		"page_number": number, "job_id": jobID.String(),
		"scheduled_run_at":  runAt.Format(time.RFC3339Nano),
		"payload_reference": reference,
		"payload_sha256":    hex.EncodeToString(referenceSHA[:]),
		"after_resource_id": after, "through_resource_id": through,
		"expected_count": count, "source_sha256": hex.EncodeToString(source[:]),
		"expected_target_sha256": hex.EncodeToString(target[:]),
	}
}
