//go:build integration

package maintenance

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"testing"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/google/uuid"
)

const (
	maintenanceOptionalSchemaID = "42000000-0000-4000-8000-000000000002"
	maintenanceOptionalFieldID  = "44000000-0000-4000-8000-000000000001"
)

const maintenanceOptionalFieldsSQL = `INSERT INTO iam_profile_fields(
 id,resource_kind,field_key,value_kind,cardinality,required,classification,
 normalization_profile,state,created_at,updated_at,retired_at,revision)
SELECT (field.value->>'id')::uuid,'user',field.value->>'key',
 field.value->>'logical_type',field.value->>'cardinality',
 (field.value->>'required')::boolean,field.value->>'classification',
 field.value->>'normalization_profile','active','2026-07-16T02:00:00Z',
 '2026-07-16T02:00:00Z',NULL,1
FROM pg_catalog.jsonb_array_elements($1::jsonb->'fields') AS field(value)
WHERE NOT EXISTS (SELECT 1 FROM iam_profile_fields AS stored
 WHERE stored.id=(field.value->>'id')::uuid)`

const maintenanceOptionalVersionSQL = `INSERT INTO iam_profile_schema_versions(
 id,resource_kind,version,state,definition_sha256,activated_at,retired_at,
 created_at,updated_at,revision)
VALUES($1,'user',2,'draft',$2,NULL,NULL,'2026-07-16T02:00:00Z',
 '2026-07-16T02:00:00Z',1)`

const maintenanceOptionalRevisionsSQL = `INSERT INTO iam_profile_field_revisions(
 schema_version_id,resource_kind,field_id,validation_schema,
 validation_sha256,created_at,field_key,value_kind,cardinality,required,
 classification,normalization_profile,field_state,semantic_sha256)
SELECT $1,'user',(field.value->>'id')::uuid,field.value->'validation',
 pg_catalog.sha256(pg_catalog.convert_to(
  iam_profile_validation_canonical_v1(field.value->'validation'),'UTF8')),
 '2026-07-16T02:00:00Z',field.value->>'key',field.value->>'logical_type',
 field.value->>'cardinality',(field.value->>'required')::boolean,
 field.value->>'classification',field.value->>'normalization_profile',
 field.value->>'state',iam_profile_semantic_sha256_v1(
  (field.value->>'id')::uuid,'user',field.value->>'key',
  field.value->>'logical_type',field.value->>'cardinality',
  (field.value->>'required')::boolean,field.value->>'classification',
  field.value->>'normalization_profile')
FROM pg_catalog.jsonb_array_elements($2::jsonb->'fields') AS field(value)`

const maintenanceOptionalRetireSQL = `UPDATE iam_profile_schema_versions SET state='retired',
 retired_at='2026-07-16T02:00:00Z',updated_at='2026-07-16T02:00:00Z',revision=2
WHERE id=$1 AND state='active'`

const maintenanceOptionalActivateSQL = `UPDATE iam_profile_schema_versions SET state='active',
 activated_at='2026-07-16T02:00:00Z',updated_at='2026-07-16T02:00:00Z',revision=2
WHERE id=$1 AND state='draft'`

func seedMaintenancePG18OptionalProfileVersion(
	t *testing.T, ctx context.Context, database *sql.DB,
) {
	t.Helper()
	for index, tenant := range maintenancePG18Tenants {
		schema := newMaintenanceOptionalSchema(t, index)
		definition := schema.DefinitionSHA256()
		execMaintenancePG18Tenant(t, ctx, database, tenant,
			maintenanceOptionalFieldsSQL, schema.DefinitionJSON())
		execMaintenancePG18Tenant(t, ctx, database, tenant,
			maintenanceOptionalVersionSQL, schema.ID(), definition[:])
		execMaintenancePG18Tenant(t, ctx, database, tenant,
			maintenanceOptionalRevisionsSQL, schema.ID(), schema.DefinitionJSON())
		execMaintenancePG18Tenant(t, ctx, database, tenant,
			maintenanceOptionalRetireSQL, maintenanceUserSchemaID)
		execMaintenancePG18Tenant(t, ctx, database, tenant,
			maintenanceOptionalActivateSQL, schema.ID())
		document, digest := maintenanceMixedSourceDocument(index, 1)
		execMaintenancePG18Tenant(t, ctx, database, tenant, `UPDATE iam_users
SET profile_schema_version_id=$2,profile_values=$3::jsonb,profile_sha256=$4,
 revision=revision+1,updated_at='2026-07-16T02:01:00Z' WHERE id=$1`,
			maintenanceUserTwo, schema.ID(), document, digest[:])
	}
}

func newMaintenanceOptionalSchema(t *testing.T, tenantIndex int) profileapp.Schema {
	t.Helper()
	old := maintenanceField(t, maintenanceProfileField, "display_name",
		profileapp.FieldActive, 8)
	optional := maintenanceField(t, maintenanceOptionalFieldID, "department_code",
		profileapp.FieldActive, 64)
	result, err := profileapp.NewPublishedSchema(profileapp.SchemaInput{
		ID: uuid.MustParse(maintenanceOptionalSchemaID), Resource: profileapp.ResourceUser,
		Version: 2, State: profileapp.SchemaActive, Fields: []profileapp.Field{old, optional},
	})
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func maintenanceField(
	t *testing.T, id, key string, state profileapp.FieldState, maximum int,
) profileapp.Field {
	t.Helper()
	field, err := profileapp.NewField(profileapp.FieldInput{
		ID: uuid.MustParse(id), Resource: profileapp.ResourceUser, Key: key,
		Type: profileapp.LogicalString, Cardinality: profileapp.CardinalitySingle,
		Classification: profileapp.ClassificationInternal,
		Normalization:  profileapp.NormalizationNoneV1, State: state,
		Validation: profileapp.ValidationInput{MaximumBytes: maximum},
	})
	if err != nil {
		t.Fatal(err)
	}
	return field
}

func maintenanceMixedSourceDocument(tenantIndex, userIndex int) (string, [32]byte) {
	name := maintenanceMixedProfileName(tenantIndex, userIndex)
	if userIndex != 1 {
		return maintenanceProfileDocument(maintenanceProfileField, name)
	}
	document := fmt.Sprintf(`{"%s":%q,"%s":%q}`,
		maintenanceProfileField, name, maintenanceOptionalFieldID,
		fmt.Sprintf("D-%d", tenantIndex+1))
	return document, sha256.Sum256([]byte(document))
}

func maintenanceMixedTargetDocument(tenantIndex, userIndex int) (string, [32]byte) {
	name := maintenanceMixedProfileName(tenantIndex, userIndex)
	if userIndex != 1 {
		document := fmt.Sprintf(`{"%s":%q,"%s":%q}`,
			maintenanceProfileField, name, maintenanceTargetFieldID, name)
		return document, sha256.Sum256([]byte(document))
	}
	document := fmt.Sprintf(`{"%s":%q,"%s":%q,"%s":%q}`,
		maintenanceProfileField, name, maintenanceOptionalFieldID,
		fmt.Sprintf("D-%d", tenantIndex+1), maintenanceTargetFieldID, name)
	return document, sha256.Sum256([]byte(document))
}

func maintenanceMixedProfileName(tenantIndex, userIndex int) string {
	if userIndex == 1 {
		return fmt.Sprintf("T%dU2", tenantIndex+1)
	}
	if userIndex == 2 {
		return fmt.Sprintf("Tenant %d User 3 With A Long Compatibility Name", tenantIndex+1)
	}
	return fmt.Sprintf("Tenant %d User %d", tenantIndex+1, userIndex+1)
}
