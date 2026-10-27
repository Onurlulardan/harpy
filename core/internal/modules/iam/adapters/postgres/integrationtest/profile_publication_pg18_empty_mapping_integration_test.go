//go:build integration

package integrationtest

import (
	"context"
	"crypto/sha256"
	"testing"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/google/uuid"
)

const (
	publicationEmptyOrganization = "51000000-0000-4000-8000-000000000101"
	publicationEmptyV1           = "41000000-0000-4000-8000-000000000101"
	publicationEmptyV2           = "41000000-0000-4000-8000-000000000102"
	publicationEmptyV3           = "41000000-0000-4000-8000-000000000103"
	publicationEmptyField        = "43000000-0000-4000-8000-000000000101"
	publicationOrganizationPerm  = "93000000-0000-4000-8000-000000000101"
)

func seedPublicationEmptyMappingSource(
	t *testing.T, ctx context.Context, db publicationDatabase,
) {
	t.Helper()
	empty := publicationEmptySchema(t)
	current := publicationEmptyCurrentSchema(t)
	field := current.Fields()[0]
	emptyDefinition, currentDefinition := empty.DefinitionSHA256(), current.DefinitionSHA256()
	validation, validationSHA := field.ValidationJSON(), field.ValidationSHA256()
	documentSHA := sha256.Sum256([]byte(`{}`))
	const statement = `INSERT INTO iam_profile_schema_versions(
 id,resource_kind,version,state,definition_sha256,activated_at,retired_at,
 created_at,updated_at,revision) VALUES
($1,'organization',1,'retired',$2,'2026-07-16T00:00:00Z','2026-07-16T00:01:00Z',
 '2026-07-16T00:00:00Z','2026-07-16T00:01:00Z',2),
($3,'organization',2,'active',$4,'2026-07-16T00:01:00Z',NULL,
 '2026-07-16T00:01:00Z','2026-07-16T00:01:00Z',1);
INSERT INTO iam_profile_fields(
 id,resource_kind,field_key,value_kind,cardinality,required,classification,
 normalization_profile,state,created_at,updated_at,retired_at,revision)
VALUES($5,'organization','nickname','string','single',false,'internal','none_v1',
 'active','2026-07-16T00:01:00Z','2026-07-16T00:01:00Z',NULL,1);
INSERT INTO iam_profile_field_revisions(
 schema_version_id,resource_kind,field_id,validation_schema,
 validation_sha256,created_at) VALUES($3,'organization',$5,$6::jsonb,$7,
 '2026-07-16T00:01:00Z');
INSERT INTO iam_organizations(
 id,name,slug,status,parent_id,profile_resource_kind,profile_schema_version_id,
 profile_values,profile_sha256,revision,created_at,updated_at,tombstoned_at)
VALUES($8,'Empty Profile','empty-profile','active',NULL,'organization',$1,
 '{}'::jsonb,$9,1,'2026-07-16T00:01:00Z','2026-07-16T00:01:00Z',NULL);`
	execPublicationTenant(t, ctx, db, statement,
		empty.ID(), emptyDefinition[:], current.ID(), currentDefinition[:], field.ID(),
		string(validation), validationSHA[:], publicationEmptyOrganization, documentSHA[:])
}

func seedPublicationEmptyMappingAuthority(
	t *testing.T, ctx context.Context, db publicationDatabase,
) {
	t.Helper()
	const statement = `INSERT INTO iam_resources(
 id,resource_key,owner_domain,resource_kind,display_name,description,is_protected,
 state,revision,created_at,updated_at,retired_at) VALUES
('91000000-0000-4000-8000-000000000101','iam.organizations','iam',
 'system_resource','Organizations',NULL,true,'active',1,pg_catalog.now(),pg_catalog.now(),NULL);
INSERT INTO iam_permissions(
 id,match_kind,resource_id,action_id,is_protected,state,revision,
 created_at,updated_at,retired_at) VALUES($1,'exact',
 '91000000-0000-4000-8000-000000000101',
 '92000000-0000-4000-8000-000000000001',false,'active',1,
 pg_catalog.now(),pg_catalog.now(),NULL);
INSERT INTO iam_user_permission_bindings(
 id,user_id,permission_id,permission_match_kind,state,revision,
 created_at,updated_at,activated_at,revoked_at) VALUES
('a0040000-0000-4000-8000-000000000101',$2,$1,'exact','active',1,
 pg_catalog.now(),pg_catalog.now(),pg_catalog.now(),NULL);`
	execPublicationTenant(t, ctx, db, statement,
		publicationOrganizationPerm, publicationActorID)
}

func assertPublicationEmptyMappingLifecycle(
	t *testing.T, ctx context.Context, db publicationDatabase, runtime publicationRuntime,
) {
	t.Helper()
	planned, err := runtime.application.PublishSchema(
		ctx, publicationClaims(t, db.tenant), publicationEmptyTargetInput(),
	)
	if err != nil || planned.BackfillPlanID == nil || planned.PublicationState != "planned" {
		t.Fatalf("empty mapping publication = %#v/%v", planned, err)
	}
	executePublicationProfileWorker(t, ctx, db, *planned.BackfillPlanID)
	var planState, targetState string
	var rowSchema uuid.UUID
	var document string
	err = db.admin.QueryRow(ctx, `SELECT plan.state,target.state,
organization.profile_schema_version_id,organization.profile_values::text
FROM t_profile_publication.iam_profile_backfill_plans AS plan
JOIN t_profile_publication.iam_profile_schema_versions AS target
 ON target.id=plan.target_schema_id
JOIN t_profile_publication.iam_organizations AS organization ON organization.id=$2
WHERE plan.id=$1`, *planned.BackfillPlanID, publicationEmptyOrganization).
		Scan(&planState, &targetState, &rowSchema, &document)
	if err != nil || planState != "activated" || targetState != "active" ||
		rowSchema != uuid.MustParse(publicationEmptyV3) || document != "{}" {
		t.Fatalf("empty mapping activation = %s/%s/%s/%s/%v",
			planState, targetState, rowSchema, document, err)
	}
}

func publicationEmptySchema(t *testing.T) profileapp.Schema {
	t.Helper()
	schema, err := profileapp.NewPublishedSchema(profileapp.SchemaInput{
		ID: uuid.MustParse(publicationEmptyV1), Resource: profileapp.ResourceOrganization,
		Version: 1, State: profileapp.SchemaRetired, Fields: []profileapp.Field{},
	})
	if err != nil {
		t.Fatal(err)
	}
	return schema
}

func publicationEmptyCurrentSchema(t *testing.T) profileapp.Schema {
	t.Helper()
	field, err := profileapp.NewField(profileapp.FieldInput{
		ID: uuid.MustParse(publicationEmptyField), Resource: profileapp.ResourceOrganization,
		Key: "nickname", Type: profileapp.LogicalString,
		Cardinality: profileapp.CardinalitySingle, Required: false,
		Classification: profileapp.ClassificationInternal,
		Normalization:  profileapp.NormalizationNoneV1, State: profileapp.FieldActive,
		Validation: profileapp.ValidationInput{MaximumBytes: 8},
	})
	if err != nil {
		t.Fatal(err)
	}
	schema, err := profileapp.NewPublishedSchema(profileapp.SchemaInput{
		ID: uuid.MustParse(publicationEmptyV2), Resource: profileapp.ResourceOrganization,
		Version: 2, State: profileapp.SchemaActive, Fields: []profileapp.Field{field},
	})
	if err != nil {
		t.Fatal(err)
	}
	return schema
}

func publicationEmptyTargetInput() iammodule.ProfileSchemaPublish {
	return iammodule.ProfileSchemaPublish{
		IdempotencyKey: "profile-empty-mapping-v3",
		SchemaID:       uuid.MustParse(publicationEmptyV3),
		Resource:       profileapp.ResourceOrganization, Version: 3,
		ExpectedActiveSchemaID: uuid.MustParse(publicationEmptyV2),
		Fields: []iammodule.ProfileFieldInput{{
			ID: uuid.MustParse(publicationEmptyField), Key: "nickname",
			Type: profileapp.LogicalString, Cardinality: profileapp.CardinalitySingle,
			Classification: profileapp.ClassificationInternal,
			Normalization:  profileapp.NormalizationNoneV1, State: profileapp.FieldActive,
			Validation: iammodule.ProfileValidationInput{MaximumBytes: 16},
		}},
	}
}
