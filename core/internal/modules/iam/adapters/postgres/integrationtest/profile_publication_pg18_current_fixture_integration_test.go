//go:build integration

package integrationtest

import (
	"context"
	"testing"

	profilepostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/identity/profile"
	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/systemresource"
	identitydomain "github.com/Onurlulardan/harpy/core/internal/modules/iam/domain/identity"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func seedPublicationCurrentSource(
	t *testing.T, ctx context.Context, db publicationDatabase,
) {
	t.Helper()
	current := publicationCurrentSchema(t)
	fields := current.Fields()
	definition := current.DefinitionSHA256()
	oldValidation, optionalValidation := fields[0].ValidationJSON(), fields[1].ValidationJSON()
	oldSHA, optionalSHA := fields[0].ValidationSHA256(), fields[1].ValidationSHA256()
	const statement = `UPDATE iam_profile_schema_versions SET state='retired',
retired_at='2026-07-16T00:01:00Z',updated_at='2026-07-16T00:01:00Z',revision=2
WHERE id=$1;
INSERT INTO iam_profile_schema_versions(
 id,resource_kind,version,state,definition_sha256,activated_at,retired_at,
 created_at,updated_at,revision) VALUES($2,'user',2,'active',$3,
 '2026-07-16T00:01:00Z',NULL,'2026-07-16T00:01:00Z','2026-07-16T00:01:00Z',1);
INSERT INTO iam_profile_fields(
 id,resource_kind,field_key,value_kind,cardinality,required,classification,
 normalization_profile,state,created_at,updated_at,retired_at,revision)
VALUES($4,'user','nickname','string','single',false,'internal','none_v1',
 'active','2026-07-16T00:01:00Z','2026-07-16T00:01:00Z',NULL,1);
INSERT INTO iam_profile_field_revisions(
 schema_version_id,resource_kind,field_id,validation_schema,
 validation_sha256,created_at) VALUES
($2,'user',$5,$6::jsonb,$7,'2026-07-16T00:01:00Z'),
($2,'user',$4,$8::jsonb,$9,'2026-07-16T00:01:00Z');`
	execPublicationTenant(
		t, ctx, db, statement, publicationSourceID, current.ID(), definition[:],
		fields[1].ID(), fields[0].ID(), string(oldValidation), oldSHA[:],
		string(optionalValidation), optionalSHA[:],
	)
}

func assertPublicationProfileSourceRead(
	t *testing.T, ctx context.Context, db publicationDatabase, runtime publicationRuntime,
) {
	t.Helper()
	writer, err := profilepostgres.NewSelfWriter()
	if err != nil {
		t.Fatal(err)
	}
	profiles, err := profilepostgres.NewStore()
	if err != nil {
		t.Fatal(err)
	}
	tenantID := uuid.MustParse(db.tenant.id)
	generations, err := runtime.pool.ResolveSnapshotDomainGenerations(
		ctx, tenantID, []string{"iam", "system_resource_layout"},
	)
	if err != nil {
		t.Fatal(err)
	}
	request, err := queryapp.SnapshotScopeRequest(
		uuid.MustParse(publicationActorID), tenantID,
		map[string]int64{"iam": generations["iam"]},
	)
	if err != nil {
		t.Fatal(err)
	}
	err = runtime.coordinator.InTenant(
		ctx, request, func(scoped context.Context, scope platformtx.TenantScope) error {
			_, _, readErr := writer.SourceForSelf(
				scoped, scope, uuid.MustParse(publicationActorID),
			)
			if readErr != nil {
				return readErr
			}
			return assertPublicationHistoricalAlias(
				scoped, scope, profiles,
			)
		},
	)
	if err != nil {
		t.Fatalf("read historical profile source: %v", err)
	}
}

func assertPublicationHistoricalAlias(
	ctx context.Context,
	scope platformtx.TenantScope,
	profiles interface {
		FieldIDsForSchema(
			context.Context, platformtx.TenantScope,
			identitydomain.ProfileResource, identitydomain.IdentityID, []string,
		) (map[string]identitydomain.IdentityID, error)
	},
) error {
	aliases, err := profiles.FieldIDsForSchema(
		ctx, scope, identitydomain.ProfileUser,
		identitydomain.IdentityID(uuid.MustParse(publicationSourceID)),
		[]string{"display_name"},
	)
	if err != nil {
		return err
	}
	if aliases["display_name"] !=
		identitydomain.IdentityID(uuid.MustParse(publicationSourceField)) {
		return identitydomain.ErrStorage
	}
	if _, err = profiles.FieldIDsForSchema(
		ctx, scope, identitydomain.ProfileUser,
		identitydomain.IdentityID(
			uuid.MustParse("41000000-0000-4000-8000-000000000099"),
		),
		[]string{"display_name"},
	); err == nil {
		return identitydomain.ErrStorage
	}
	return nil
}

func publicationCurrentSchema(t *testing.T) profileapp.Schema {
	t.Helper()
	fields := []profileapp.Field{
		publicationCurrentField(t, publicationSourceField, "display_name"),
		publicationCurrentField(t, publicationOptionalID, "nickname"),
	}
	schema, err := profileapp.NewPublishedSchema(profileapp.SchemaInput{
		ID: uuid.MustParse(publicationOptionalV2ID), Resource: profileapp.ResourceUser,
		Version: 2, State: profileapp.SchemaActive, Fields: fields,
	})
	if err != nil {
		t.Fatal(err)
	}
	return schema
}

func publicationCurrentField(t *testing.T, id, key string) profileapp.Field {
	t.Helper()
	classification := profileapp.ClassificationInternal
	if key == "display_name" {
		classification = profileapp.ClassificationPublic
	}
	field, err := profileapp.NewField(profileapp.FieldInput{
		ID: uuid.MustParse(id), Resource: profileapp.ResourceUser, Key: key,
		Type: profileapp.LogicalString, Cardinality: profileapp.CardinalitySingle,
		Classification: classification,
		Normalization:  profileapp.NormalizationNoneV1, State: profileapp.FieldActive,
		Validation: profileapp.ValidationInput{MaximumBytes: 8},
	})
	if err != nil {
		t.Fatal(err)
	}
	return field
}
