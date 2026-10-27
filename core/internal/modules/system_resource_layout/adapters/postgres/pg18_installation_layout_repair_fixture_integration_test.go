//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
)

type layoutRepairCatalogField struct {
	ID           string `json:"id"`
	Mode         string `json:"mode"`
	TypeID       string `json:"type_id"`
	ValueProfile string `json:"value_profile"`
}

func layoutRepairCatalog(kind string) string {
	types := map[string][]string{
		"user": {"uuid", "text", "text", "enum", "enum", "uuid_list",
			"uuid_list", "uuid", "integer", "timestamp", "timestamp"},
		"role": {"uuid", "enum", "uuid", "relation", "text", "text", "enum",
			"boolean", "boolean", "uuid_list", "uuid", "integer",
			"timestamp", "timestamp"},
		"organization": {"uuid", "text", "text", "enum", "relation", "uuid_list",
			"uuid_list", "uuid", "integer", "timestamp", "timestamp"},
	}[kind]
	prefix := map[string]string{
		"user":         "a1000000-0000-4000-8000-",
		"role":         "b1000000-0000-4000-8000-",
		"organization": "c1000000-0000-4000-8000-",
	}[kind]
	editable := map[string]bool{
		"user:002": true, "user:006": true, "user:007": true,
		"role:002": true, "role:003": true, "role:004": true,
		"role:005": true, "role:006": true, "role:009": true,
		"organization:002": true, "organization:003": true,
		"organization:005": true,
	}
	profile := map[string]string{
		"uuid": "uuid_rfc4122_lower_v1", "text": "text_nfc_v1",
		"enum": "enum_code_v1", "uuid_list": "uuid_list_v1",
		"integer":   "integer_base10_v1",
		"timestamp": "timestamp_utc_microsecond_v1",
		"relation":  "relation_uuid_v1", "boolean": "boolean_v1",
	}
	fields := make([]layoutRepairCatalogField, 0, len(types)+1)
	for index, typeID := range types {
		suffix := fmt.Sprintf("%03x", index+1)
		mode := "read_only"
		if editable[kind+":"+suffix] {
			mode = "editable"
		}
		fields = append(fields, layoutRepairCatalogField{
			ID: prefix + fmt.Sprintf("%012x", index+1), Mode: mode,
			TypeID: typeID, ValueProfile: profile[typeID],
		})
	}
	dynamic := map[string]string{
		"user":         "71000000-0000-7000-8000-000000000001",
		"organization": "71000000-0000-7000-8000-000000000002",
		"role":         "71000000-0000-7000-8000-000000000003",
	}[kind]
	fields = append(fields, layoutRepairCatalogField{
		ID: dynamic, Mode: "editable", TypeID: "text",
		ValueProfile: "text_nfc_v1",
	})
	document, _ := json.Marshal(struct {
		Contract string                     `json:"contract"`
		Fields   []layoutRepairCatalogField `json:"fields"`
	}{"forms-presentation-field-catalog.v1", fields})
	return string(document)
}

func seedInstallationLayoutRepairPG18(
	t *testing.T, ctx context.Context, database *sql.DB, schema string,
) {
	t.Helper()
	for _, revision := range []struct{ id, kind string }{
		{layoutContractUserRevision, "user"},
		{layoutContractOrganizationRevision, "organization"},
		{layoutContractRoleRevision, "role"},
	} {
		createLayoutRevision(t, ctx, database, schema, revision.id,
			layoutapi.ResourceKind(revision.kind), adapterDefinition)
		connection, transaction := beginLayoutOwnerTx(t, ctx, database, schema)
		_, err := transaction.ExecContext(ctx, `
UPDATE system_resource_layout_revisions
SET state='published',published_at=$2,updated_at=$2,row_revision=2
WHERE id=$1::uuid AND state='draft' AND row_revision=1`,
			revision.id, layoutContractTime)
		if err == nil {
			catalog := layoutRepairCatalog(revision.kind)
			_, err = transaction.ExecContext(ctx, `
INSERT INTO system_resource_layout_publications (
 resource_kind,revision_id,generation,created_at,updated_at,
 source_definition_sha256,compiler_version,catalog_contract,catalog_json,
 catalog_sha256,presentation_contract,presentation_json,presentation_sha256)
SELECT resource_kind,id,1,$2,$2,definition_sha256,
 'system-resource-presentation-compiler.v2',
 'forms-presentation-field-catalog.v1',$3,
 pg_catalog.sha256(pg_catalog.convert_to($3,'UTF8')),
 'forms-presentation.v1',$4,
 pg_catalog.sha256(pg_catalog.convert_to($4,'UTF8'))
FROM system_resource_layout_revisions WHERE id=$1::uuid`,
				revision.id, layoutContractTime, catalog, layoutPresentation)
		}
		if err != nil {
			_ = transaction.Rollback()
			connection.Close()
			t.Fatal(err)
		}
		if err := transaction.Commit(); err != nil {
			connection.Close()
			t.Fatal(err)
		}
		connection.Close()
	}
	connection, transaction := beginLayoutOwnerTx(t, ctx, database, schema)
	defer connection.Close()
	_, err := transaction.ExecContext(ctx, `
INSERT INTO system_resource_layout_installation_baselines (
 operation_id,tenant_id,user_revision_id,organization_revision_id,
 role_revision_id,baseline_sha256,initialized_at)
SELECT $1::uuid,$2::uuid,$3::uuid,$4::uuid,$5::uuid,pg_catalog.sha256(
 pg_catalog.uuid_send($1::uuid)||pg_catalog.uuid_send($2::uuid)||
 pg_catalog.uuid_send($3::uuid)||u.catalog_sha256||u.presentation_sha256||
 pg_catalog.uuid_send($4::uuid)||o.catalog_sha256||o.presentation_sha256||
 pg_catalog.uuid_send($5::uuid)||r.catalog_sha256||r.presentation_sha256),$6
FROM system_resource_layout_publications u,
     system_resource_layout_publications o,
     system_resource_layout_publications r
WHERE u.resource_kind='user' AND o.resource_kind='organization'
  AND r.resource_kind='role'`,
		layoutContractOperationID, layoutContractTenantID,
		layoutContractUserRevision, layoutContractOrganizationRevision,
		layoutContractRoleRevision, layoutContractTime)
	if err != nil {
		_ = transaction.Rollback()
		t.Fatal(err)
	}
	if err := transaction.Commit(); err != nil {
		t.Fatal(err)
	}
}

func beginLayoutOwnerTx(
	t *testing.T, ctx context.Context, database *sql.DB, schema string,
) (*sql.Conn, *sql.Tx) {
	t.Helper()
	connection, err := database.Conn(ctx)
	if err != nil {
		t.Fatal(err)
	}
	transaction, err := connection.BeginTx(ctx, nil)
	if err != nil {
		connection.Close()
		t.Fatal(err)
	}
	var bound string
	if err := transaction.QueryRowContext(ctx,
		"SELECT pg_catalog.set_config('search_path',$1,true)",
		schema+", pg_catalog").Scan(&bound); err != nil {
		_ = transaction.Rollback()
		connection.Close()
		t.Fatal(err)
	}
	return connection, transaction
}

func seedCustomLayoutRepairPG18(
	t *testing.T, ctx context.Context, database *sql.DB, schema string,
) {
	t.Helper()
	createLayoutRevision(t, ctx, database, schema, layoutSecondRevision,
		layoutapi.ResourceUser, adapterDefinition)
	publishLayoutRevision(t, ctx, database, schema, layoutSecondRevision, 1,
		adapterDefinition, layoutContractTime.Add(time.Minute))
}
