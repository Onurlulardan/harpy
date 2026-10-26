//go:build integration

package integrationtest

import (
	"context"
	"encoding/json"
	"testing"

	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	"github.com/google/uuid"
)

type recoveryPlanRow struct {
	id       uuid.UUID
	schema   profileapp.Schema
	revision int64
	document profileapp.Document
}

func seedRecoveryUserRows(
	t *testing.T, ctx context.Context, db publicationDatabase,
) []recoveryPlanRow {
	t.Helper()
	source := publicationCurrentSchema(t)
	values := []struct {
		id, email string
		revision  int64
	}{
		{publicationActorID, "a@b.co", 3},
		{recoveryUserTwoID, "c@d.co", 1},
		{recoveryUserThreeID, "e@f.co", 1},
	}
	documents := make([]profileapp.Document, len(values))
	for index, value := range values {
		documents[index] = recoveryDocument(t, source, map[string]string{
			publicationSourceField: value.email,
		})
	}
	digestOne := documents[0].SHA256()
	digestTwo := documents[1].SHA256()
	digestThree := documents[2].SHA256()
	const statement = `UPDATE iam_users SET profile_schema_version_id=$2,
profile_values=$3::jsonb,profile_sha256=$4,revision=revision+1,
updated_at=pg_catalog.now() WHERE id=$1;
INSERT INTO iam_users(
 id,identity_kind,status,profile_resource_kind,profile_schema_version_id,
 profile_values,profile_sha256,token_version,authorization_version,revision,
 created_at,updated_at,tombstoned_at) VALUES
($5,'human','active','user',$2,$6::jsonb,$7,1,0,1,
 pg_catalog.now(),pg_catalog.now(),NULL),
($8,'human','active','user',$2,$9::jsonb,$10,1,0,1,
 pg_catalog.now(),pg_catalog.now(),NULL);`
	execPublicationTenant(t, ctx, db, statement,
		values[0].id, source.ID(), string(documents[0].CanonicalJSON()),
		digestOne[:], values[1].id, string(documents[1].CanonicalJSON()),
		digestTwo[:], values[2].id, string(documents[2].CanonicalJSON()),
		digestThree[:])
	rows := make([]recoveryPlanRow, len(values))
	for index, value := range values {
		rows[index] = recoveryPlanRow{
			id: uuid.MustParse(value.id), schema: source,
			revision: value.revision, document: documents[index],
		}
	}
	return rows
}

func seedRecoveryOrganizationRow(
	t *testing.T, ctx context.Context, db publicationDatabase,
) []recoveryPlanRow {
	t.Helper()
	source := publicationEmptyCurrentSchema(t)
	document := recoveryDocument(t, source, map[string]string{
		publicationEmptyField: "team",
	})
	const statement = `INSERT INTO iam_organizations(
 id,name,slug,status,parent_id,profile_resource_kind,profile_schema_version_id,
 profile_values,profile_sha256,revision,created_at,updated_at,tombstoned_at)
VALUES($1,'Recovery Team','recovery-team','active',NULL,'organization',$2,
 $3::jsonb,$4,1,pg_catalog.now(),pg_catalog.now(),NULL);`
	digest := document.SHA256()
	execPublicationTenant(t, ctx, db, statement,
		recoveryOrganizationV2, source.ID(), string(document.CanonicalJSON()), digest[:])
	empty := publicationEmptySchema(t)
	historical := recoveryDocument(t, empty, map[string]string{})
	return []recoveryPlanRow{
		{id: uuid.MustParse(recoveryOrganizationV2), schema: source,
			revision: 1, document: document},
		{id: uuid.MustParse(publicationEmptyOrganization), schema: empty,
			revision: 1, document: historical},
	}
}

func recoveryDocument(
	t *testing.T, schema profileapp.Schema, values map[string]string,
) profileapp.Document {
	t.Helper()
	wire, err := json.Marshal(values)
	if err != nil {
		t.Fatal(err)
	}
	document, err := profileapp.DecodeDocument(
		schema, wire, nil, recoveryTextNormalizer{},
	)
	if err != nil {
		t.Fatalf("decode recovery fixture document %s: %v", wire, err)
	}
	return document
}

func readRecoveryUserRows(
	t *testing.T, ctx context.Context, db publicationDatabase,
	schemas ...profileapp.Schema,
) []recoveryPlanRow {
	t.Helper()
	return readRecoveryRows(t, ctx, db, `SELECT id,profile_schema_version_id,
revision,profile_values::text,profile_sha256 FROM t_profile_publication.iam_users
ORDER BY id`, schemas)
}

func readRecoveryOrganizationRows(
	t *testing.T, ctx context.Context, db publicationDatabase,
	schemas ...profileapp.Schema,
) []recoveryPlanRow {
	t.Helper()
	return readRecoveryRows(t, ctx, db, `SELECT id,profile_schema_version_id,
revision,profile_values::text,profile_sha256
FROM t_profile_publication.iam_organizations ORDER BY id`, schemas)
}

func readRecoveryRows(
	t *testing.T, ctx context.Context, db publicationDatabase,
	query string, schemas []profileapp.Schema,
) []recoveryPlanRow {
	t.Helper()
	byID := make(map[uuid.UUID]profileapp.Schema, len(schemas))
	for _, schema := range schemas {
		byID[schema.ID()] = schema
	}
	rows, err := db.admin.Query(ctx, query)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	result := make([]recoveryPlanRow, 0)
	for rows.Next() {
		var row recoveryPlanRow
		var schemaID uuid.UUID
		var wire, digest []byte
		if err = rows.Scan(&row.id, &schemaID, &row.revision, &wire, &digest); err != nil {
			t.Fatal(err)
		}
		row.schema = byID[schemaID]
		row.document, err = profileapp.DecodeDocument(
			row.schema, wire, digest, recoveryTextNormalizer{},
		)
		if err != nil {
			t.Fatalf("decode stored recovery row %s/%s: %v", row.id, schemaID, err)
		}
		result = append(result, row)
	}
	if rows.Err() != nil {
		t.Fatal(rows.Err())
	}
	return result
}
