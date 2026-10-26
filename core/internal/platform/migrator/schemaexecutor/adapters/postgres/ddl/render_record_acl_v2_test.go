package ddl

import (
	"reflect"
	"testing"
)

func TestRenderRecordACLStatementsV2PinsExactAuthority(t *testing.T) {
	statements := renderRecordACLStatementsV2("tenant_a", "frm_records")
	actual := make([]string, len(statements))
	for index, statement := range statements {
		actual[index] = statement.Text
	}
	wanted := []string{
		`REVOKE CREATE ON SCHEMA "tenant_a" FROM PUBLIC, harpy_runtime, ` +
			`harpy_v2_runtime, harpy_forms_record_reader, harpy_forms_record_writer`,
		`GRANT USAGE ON SCHEMA "tenant_a" TO harpy_forms_record_reader, ` +
			`harpy_forms_record_writer`,
		`REVOKE ALL PRIVILEGES ON TABLE "tenant_a"."frm_records" FROM PUBLIC, ` +
			`harpy_runtime, harpy_v2_runtime, harpy_forms_record_reader, ` +
			`harpy_forms_record_writer`,
		`GRANT SELECT ON TABLE "tenant_a"."frm_records" TO harpy_forms_record_reader`,
		`GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE "tenant_a"."frm_records" ` +
			`TO harpy_forms_record_writer`,
	}
	if !reflect.DeepEqual(actual, wanted) {
		t.Fatalf("record ACL statements = %#v", actual)
	}
}
