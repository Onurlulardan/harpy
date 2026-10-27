//go:build integration

package integrationtest

import (
	"context"
	"crypto/sha256"
	"testing"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	publicationActorID      = "51000000-0000-4000-8000-000000000001"
	publicationSourceID     = "41000000-0000-4000-8000-000000000001"
	publicationSourceField  = "43000000-0000-4000-8000-000000000001"
	publicationManagePerm   = "93000000-0000-4000-8000-000000000001"
	publicationOptionalID   = "43000000-0000-4000-8000-000000000002"
	publicationReplacement  = "43000000-0000-4000-8000-000000000003"
	publicationOptionalV2ID = "41000000-0000-4000-8000-000000000002"
	publicationTargetV3ID   = "41000000-0000-4000-8000-000000000003"
)

func seedPublicationV1Snapshot(
	t *testing.T, ctx context.Context, db publicationDatabase,
) {
	t.Helper()
	if db.tenant.origin != "v1_adopted" {
		return
	}
	const sql = `SET ROLE harpy_owner;
SET search_path TO t_profile_publication,pg_catalog;
CREATE TABLE organizations_form_table(
 id uuid PRIMARY KEY,tenant_id uuid NOT NULL,name varchar(255) NOT NULL,
 slug varchar(255) NOT NULL,status varchar(255),parent_id varchar(255));
CREATE TABLE roles_form_table(
 id uuid PRIMARY KEY,tenant_id uuid NOT NULL,name varchar(255) NOT NULL,
 organization_id varchar(255),is_default boolean);
CREATE TABLE users_form_table(
 id uuid PRIMARY KEY,tenant_id uuid NOT NULL,email varchar(255) NOT NULL,
 password varchar(255),role_ids text[],organization_id varchar(255),status varchar(255),
 token_version integer NOT NULL,account_type varchar(30) NOT NULL,
 portal_subject_type varchar(100),portal_subject_id uuid,
 portal_profile_form_slug varchar(255),portal_login_identifier varchar(100),
 portal_identifier_field varchar(255),portal_setup_required boolean NOT NULL,
 portal_setup_completed_at timestamp,workflow_status varchar(50));
CREATE TABLE refresh_tokens(
 id uuid PRIMARY KEY,user_id uuid NOT NULL,token_hash varchar(255) NOT NULL,
 expires_at timestamp NOT NULL);
RESET search_path;
RESET ROLE;`
	if _, err := db.admin.Exec(ctx, sql); err != nil {
		t.Fatalf("seed v1 adopted publication snapshot: %v", err)
	}
}

func publicationSourceSchema(t *testing.T) profileapp.Schema {
	t.Helper()
	field, err := profileapp.NewField(profileapp.FieldInput{
		ID: uuid.MustParse(publicationSourceField), Resource: profileapp.ResourceUser,
		Key: "display_name", Type: profileapp.LogicalString,
		Cardinality: profileapp.CardinalitySingle, Required: false,
		Classification: profileapp.ClassificationPublic,
		Normalization:  profileapp.NormalizationNoneV1, State: profileapp.FieldActive,
		Validation: profileapp.ValidationInput{MaximumBytes: 64},
	})
	if err != nil {
		t.Fatal(err)
	}
	schema, err := profileapp.NewPublishedSchema(profileapp.SchemaInput{
		ID: uuid.MustParse(publicationSourceID), Resource: profileapp.ResourceUser,
		Version: 1, State: profileapp.SchemaActive, Fields: []profileapp.Field{field},
	})
	if err != nil {
		t.Fatal(err)
	}
	return schema
}

func seedPublicationSource(
	t *testing.T, ctx context.Context, db publicationDatabase, schema profileapp.Schema,
) {
	t.Helper()
	field := schema.Fields()[0]
	definition := schema.DefinitionSHA256()
	validation := field.ValidationJSON()
	validationSHA := field.ValidationSHA256()
	document := []byte(`{"43000000-0000-4000-8000-000000000001":" User@Exämple.COM "}`)
	documentSHA := sha256.Sum256(document)
	const sql = `
INSERT INTO iam_profile_schema_versions(
 id,resource_kind,version,state,definition_sha256,activated_at,retired_at,
 created_at,updated_at,revision) VALUES($1,'user',1,'active',$2,
 '2026-07-16T00:00:00Z',NULL,'2026-07-16T00:00:00Z','2026-07-16T00:00:00Z',1);
INSERT INTO iam_profile_fields(
 id,resource_kind,field_key,value_kind,cardinality,required,classification,
 normalization_profile,state,created_at,updated_at,retired_at,revision)
VALUES($3,'user','display_name','string','single',false,'public','none_v1',
 'active','2026-07-16T00:00:00Z','2026-07-16T00:00:00Z',NULL,1);
INSERT INTO iam_profile_field_revisions(
 schema_version_id,resource_kind,field_id,validation_schema,
 validation_sha256,created_at) VALUES($1,'user',$3,$4::jsonb,$5,
 '2026-07-16T00:00:00Z');
INSERT INTO iam_users(
 id,identity_kind,status,profile_resource_kind,profile_schema_version_id,
 profile_values,profile_sha256,token_version,authorization_version,revision,
 created_at,updated_at,tombstoned_at) VALUES($6,'human','active','user',$1,
 $7::jsonb,$8,1,0,1,'2026-07-16T00:00:00Z','2026-07-16T00:00:00Z',NULL);`
	execPublicationTenant(t, ctx, db, sql, schema.ID(), definition[:], field.ID(),
		string(validation), validationSHA[:], publicationActorID, string(document), documentSHA[:])
}

func seedPublicationAuthority(t *testing.T, ctx context.Context, db publicationDatabase) {
	t.Helper()
	const sql = `
INSERT INTO iam_resources(
 id,resource_key,owner_domain,resource_kind,display_name,description,is_protected,
 state,revision,created_at,updated_at,retired_at) VALUES
('91000000-0000-4000-8000-000000000001','iam.users','iam','system_resource',
 'Users',NULL,true,'active',1,pg_catalog.now(),pg_catalog.now(),NULL);
INSERT INTO iam_actions(
 id,action_key,display_name,description,is_protected,state,revision,
 created_at,updated_at,retired_at) VALUES
('92000000-0000-4000-8000-000000000001','manage','Manage',NULL,false,'active',1,
 pg_catalog.now(),pg_catalog.now(),NULL);
INSERT INTO iam_permissions(
 id,match_kind,resource_id,action_id,is_protected,state,revision,
 created_at,updated_at,retired_at) VALUES($1,'exact',
 '91000000-0000-4000-8000-000000000001',
 '92000000-0000-4000-8000-000000000001',false,'active',1,
 pg_catalog.now(),pg_catalog.now(),NULL);
INSERT INTO iam_authorization_clock VALUES(true,0,1,pg_catalog.now(),pg_catalog.now());
UPDATE iam_users SET authorization_version=1,revision=revision+1,
 updated_at=pg_catalog.now() WHERE id=$2;
INSERT INTO iam_user_permission_bindings(
 id,user_id,permission_id,permission_match_kind,state,revision,
 created_at,updated_at,activated_at,revoked_at) VALUES
('a0040000-0000-4000-8000-000000000001',$2,$1,'exact','active',1,
 pg_catalog.now(),pg_catalog.now(),pg_catalog.now(),NULL);`
	execPublicationTenant(t, ctx, db, sql, publicationManagePerm, publicationActorID)
}

func execPublicationTenant(
	t *testing.T, ctx context.Context, db publicationDatabase,
	statement string, arguments ...any,
) {
	t.Helper()
	tx, err := db.admin.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tx.Exec(ctx, "SET LOCAL ROLE harpy_owner")
	if err == nil {
		_, err = tx.Exec(ctx, "SELECT pg_catalog.set_config('search_path',$1,true)",
			db.tenant.schema+",platform,pg_catalog")
	}
	if err == nil {
		arguments = append([]any{pgx.QueryExecModeSimpleProtocol}, arguments...)
		_, err = tx.Exec(ctx, statement, arguments...)
	}
	if err != nil {
		_ = tx.Rollback(ctx)
		t.Fatalf("execute publication fixture: %v", err)
	}
	if err = tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
}
