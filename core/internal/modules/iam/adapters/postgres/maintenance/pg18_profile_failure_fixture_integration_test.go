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

const (
	maintenanceInvalidTargetFieldID  = "46000000-0000-4000-8000-000000000002"
	maintenanceInvalidTargetSchemaID = "45000000-0000-4000-8000-000000000003"
	maintenanceInvalidPlanID         = "47000000-0000-4000-8000-000000000003"
)

func createMaintenancePG18InvalidProfilePlan(
	t *testing.T, ctx context.Context, database *sql.DB,
	source maintenancePlanFixture,
) maintenancePlanFixture {
	t.Helper()
	assertMaintenancePG18InvalidSourceInventory(t, ctx, database, source)
	target := newMaintenanceInvalidTargetSchema(t, source.targetSchema)
	jobs := [2]uuid.UUID{
		uuid.MustParse("019c6a7e-7c00-7000-8000-000000000501"),
		uuid.MustParse("019c6a7e-7c00-7000-8000-000000000502"),
	}
	planID := uuid.MustParse(maintenanceInvalidPlanID)
	document := maintenanceInvalidPlanDocument(t, source.tenant, planID,
		source.targetSchema, target, jobs)
	assertMaintenancePG18InvalidPlanWire(t, ctx, database, source, document)
	digest := sha256.Sum256(document)
	request := sha256.Sum256([]byte("profile-invalid-plan-" + source.tenant.id))
	createMaintenancePG18ProfilePlanWithOperation(
		t, ctx, database, source.tenant, planID, document, digest, request,
		uuid.MustParse("48000000-0000-4000-8000-000000000003"), false,
	)
	return maintenancePlanFixture{
		tenant: source.tenant, planID: planID, targetSchema: target, jobs: jobs,
	}
}

func newMaintenanceInvalidTargetSchema(
	t *testing.T, source profileapp.Schema,
) profileapp.Schema {
	t.Helper()
	retiredOld, _ := source.Field(uuid.MustParse(maintenanceProfileField))
	optional, _ := source.Field(uuid.MustParse(maintenanceOptionalFieldID))
	preferred, _ := source.Field(uuid.MustParse(maintenanceTargetFieldID))
	retiredPreferred, err := profileapp.RetiredField(preferred)
	if err != nil {
		t.Fatal(err)
	}
	short := maintenanceField(t, maintenanceInvalidTargetFieldID,
		"short_name", profileapp.FieldActive, 20)
	target, err := profileapp.NewPublishedSchema(profileapp.SchemaInput{
		ID:       uuid.MustParse(maintenanceInvalidTargetSchemaID),
		Resource: profileapp.ResourceUser, Version: 4,
		State:  profileapp.SchemaActive,
		Fields: []profileapp.Field{retiredOld, optional, retiredPreferred, short},
	})
	if err != nil {
		t.Fatal(err)
	}
	return target
}

func maintenanceInvalidPlanDocument(
	t *testing.T, tenant maintenancePG18Tenant, planID uuid.UUID,
	source, target profileapp.Schema, jobs [2]uuid.UUID,
) []byte {
	t.Helper()
	sources, targets := make([][32]byte, 4), make([][32]byte, 4)
	revisions := []string{"4", "5", "4", "1"}
	for index := range maintenanceProfileUsers {
		_, sources[index] = maintenanceMixedTargetDocument(0, index)
		_, targets[index] = maintenanceInvalidTargetDocument(0, index)
	}
	_, sources[3] = maintenanceLiveCurrentDocument(0)
	_, targets[3] = maintenanceInvalidLiveTargetDocument(0)
	pageOneSource := maintenancePageDigest([][]byte{
		[]byte(maintenanceUserOne), []byte(source.ID().String()),
		[]byte(revisions[0]), sources[0][:],
	})
	pageTwoSource := maintenancePageDigest(
		[][]byte{[]byte(maintenanceUserTwo), []byte(source.ID().String()),
			[]byte(revisions[1]), sources[1][:]},
		[][]byte{[]byte(maintenanceUserThree), []byte(source.ID().String()),
			[]byte(revisions[2]), sources[2][:]},
		[][]byte{[]byte(maintenanceLiveUser), []byte(source.ID().String()),
			[]byte(revisions[3]), sources[3][:]},
	)
	pageOneTarget := maintenancePageDigest(
		[][]byte{[]byte(maintenanceUserOne), targets[0][:]},
	)
	pageTwoTarget := maintenancePageDigest(
		[][]byte{[]byte(maintenanceUserTwo), targets[1][:]},
		[][]byte{[]byte(maintenanceUserThree), targets[2][:]},
		[][]byte{[]byte(maintenanceLiveUser), targets[3][:]},
	)
	sourceAggregate := sha256.Sum256(append(pageOneSource[:], pageTwoSource[:]...))
	initial := sha256.Sum256(nil)
	rollingOne := sha256.Sum256(append(initial[:], pageOneTarget[:]...))
	rollingTwo := sha256.Sum256(append(rollingOne[:], pageTwoTarget[:]...))
	short, _ := target.Field(uuid.MustParse(maintenanceInvalidTargetFieldID))
	optional, _ := target.Field(uuid.MustParse(maintenanceOptionalFieldID))
	shortSemantic := short.SemanticSHA256()
	optionalSemantic := optional.SemanticSHA256()
	definition := target.DefinitionSHA256()
	runAt := time.Date(2026, 7, 16, 15, 0, 0, 0, time.UTC)
	pages := []map[string]any{
		maintenancePlanPage(tenant, planID, jobs[0], 1, nil,
			maintenanceUserOne, 1, pageOneSource, pageOneTarget, runAt),
		maintenancePlanPage(tenant, planID, jobs[1], 2,
			&maintenanceProfileUsers[0], maintenanceLiveUser, 3,
			pageTwoSource, pageTwoTarget, runAt.Add(time.Minute)),
	}
	wire := map[string]any{
		"id": planID.String(), "resource_kind": "user",
		"source_schema_id": source.ID().String(),
		"target_schema_id": target.ID().String(), "target_version": 4,
		"target_definition":        json.RawMessage(target.DefinitionJSON()),
		"target_definition_sha256": hex.EncodeToString(definition[:]),
		"converter_version":        "typed_map_v1", "source_count": 4,
		"expected_source_sha256": hex.EncodeToString(sourceAggregate[:]),
		"expected_target_sha256": hex.EncodeToString(rollingTwo[:]),
		"page_size":              3, "initial_rolling_sha256": hex.EncodeToString(initial[:]),
		"mappings": []map[string]any{
			{"ordinal": 1, "source_field_id": maintenanceTargetFieldID,
				"target_field_id": maintenanceInvalidTargetFieldID,
				"converter_kind":  "copy_normalize_v1",
				"semantic_sha256": hex.EncodeToString(shortSemantic[:])},
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

func maintenanceInvalidTargetDocument(tenantIndex, userIndex int) (string, [32]byte) {
	name := maintenanceMixedProfileName(tenantIndex, userIndex)
	document := fmt.Sprintf(`{"%s":%q,"%s":%q,"%s":%q}`,
		maintenanceProfileField, name, maintenanceTargetFieldID, name,
		maintenanceInvalidTargetFieldID, name)
	if userIndex == 1 {
		document = fmt.Sprintf(`{"%s":%q,"%s":%q,"%s":%q,"%s":%q}`,
			maintenanceProfileField, name, maintenanceOptionalFieldID,
			fmt.Sprintf("D-%d", tenantIndex+1), maintenanceTargetFieldID, name,
			maintenanceInvalidTargetFieldID, name)
	}
	return document, sha256.Sum256([]byte(document))
}

func maintenanceLiveCurrentDocument(tenantIndex int) (string, [32]byte) {
	name := fmt.Sprintf("Live Tenant %d User", tenantIndex+1)
	document := fmt.Sprintf(`{"%s":%q,"%s":%q}`,
		maintenanceProfileField, name, maintenanceTargetFieldID, name)
	return document, sha256.Sum256([]byte(document))
}

func maintenanceInvalidLiveTargetDocument(tenantIndex int) (string, [32]byte) {
	name := fmt.Sprintf("Live Tenant %d User", tenantIndex+1)
	document := fmt.Sprintf(`{"%s":%q,"%s":%q,"%s":%q}`,
		maintenanceProfileField, name, maintenanceTargetFieldID, name,
		maintenanceInvalidTargetFieldID, name)
	return document, sha256.Sum256([]byte(document))
}
