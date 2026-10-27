//go:build integration

package integrationtest

import (
	"context"
	"encoding/json"
	"testing"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	catalogpackagev1 "github.com/Onurlulardan/harpy/contracts/catalogpackage/v1"
	iampayload "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/catalogpayload"
	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/google/uuid"
)

const (
	iamCatalogRoleSchemaID = "018f0000-0004-7000-8000-000000000002"
	iamCatalogRoleFieldID  = "018f0000-0004-7000-8000-000000000005"
)

func seedIAMCatalogRoleSchema(
	t *testing.T, ctx context.Context, db publicationDatabase,
) {
	t.Helper()
	field, err := profileapp.NewField(profileapp.FieldInput{
		ID: uuid.MustParse(iamCatalogRoleFieldID), Resource: profileapp.ResourceRole,
		Key: "display_name", Type: profileapp.LogicalString,
		Cardinality: profileapp.CardinalitySingle, Required: true,
		Classification: profileapp.ClassificationPublic,
		Normalization:  profileapp.NormalizationNoneV1,
		State:          profileapp.FieldActive,
		Validation:     profileapp.ValidationInput{MaximumBytes: 255},
	})
	if err != nil {
		t.Fatal(err)
	}
	schema, err := profileapp.NewPublishedSchema(profileapp.SchemaInput{
		ID: uuid.MustParse(iamCatalogRoleSchemaID), Resource: profileapp.ResourceRole,
		Version: 1, State: profileapp.SchemaActive, Fields: []profileapp.Field{field},
	})
	if err != nil {
		t.Fatal(err)
	}
	definition := schema.DefinitionSHA256()
	validation, validationSHA := field.ValidationJSON(), field.ValidationSHA256()
	const statement = `INSERT INTO iam_profile_schema_versions(
 id,resource_kind,version,state,definition_sha256,activated_at,retired_at,
 created_at,updated_at,revision) VALUES(
 $1,'role',1,'active',$2,'2026-07-30T10:00:00Z',NULL,
 '2026-07-30T10:00:00Z','2026-07-30T10:00:00Z',1);
INSERT INTO iam_profile_fields(
 id,resource_kind,field_key,value_kind,cardinality,required,classification,
 normalization_profile,state,created_at,updated_at,retired_at,revision)
VALUES($3,'role','display_name','string','single',true,'public','none_v1',
 'active','2026-07-30T10:00:00Z','2026-07-30T10:00:00Z',NULL,1);
INSERT INTO iam_profile_field_revisions(
 schema_version_id,resource_kind,field_id,validation_schema,
 validation_sha256,created_at) VALUES(
 $1,'role',$3,$4::jsonb,$5,'2026-07-30T10:00:00Z')`
	execPublicationTenant(
		t, ctx, db, statement, schema.ID(), definition[:], field.ID(),
		string(validation), validationSHA[:],
	)
}

func iamCatalogDocument(t *testing.T) []byte {
	t.Helper()
	payload := iampayload.Payload{
		ContractVersion: catalogpackagev1.ContractIAM,
		ComponentID:     iamCatalogComponentID, ModuleSlug: "project",
		SourceContract: "harpy-v1.enterprise-seed.project",
		Resources: []iampayload.Resource{
			{Slug: "module.project", Kind: "module", Name: "Project", Description: "Module"},
			{Slug: "form.project-task", Kind: "form", Name: "Task", Description: "Form"},
		},
		Grants: []iampayload.Grant{
			{ResourceSlug: "module.project", RoleName: "Project Administrator",
				RoleDescription: "Project administrator", Actions: []string{"read"}},
			{ResourceSlug: "form.project-task", RoleName: "Project Administrator",
				RoleDescription: "Project administrator", Actions: []string{"read", "write"}},
		},
	}
	document, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, 4<<20)
	if err != nil {
		t.Fatal(err)
	}
	return canonical
}
