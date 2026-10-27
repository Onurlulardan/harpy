package compiler

import (
	"reflect"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestRootRecordAndExactCountIgnoreIncludeStateGolden(t *testing.T) {
	without := newRootParityPlanFixture(t, api.CountModeExact, false)
	with := newRootParityPlanFixture(t, api.CountModeExact, true)
	compiler := includeCompiler(t)
	withoutBatch, err := compiler.CompileRecordBatch(without.plan)
	if err != nil {
		t.Fatal(err)
	}
	withBatch, err := compiler.CompileRecordBatch(with.plan)
	if err != nil {
		t.Fatal(err)
	}
	withoutRecordSQL, _ := withoutBatch.record.SQL()
	withRecordSQL, _ := withBatch.record.SQL()
	withoutRecordArgs, _ := withoutBatch.record.Arguments()
	withRecordArgs, _ := withBatch.record.Arguments()
	withoutCountSQL, _ := withoutBatch.exactCount.SQL()
	withCountSQL, _ := withBatch.exactCount.SQL()
	withoutCountArgs, _ := withoutBatch.exactCount.Arguments()
	withCountArgs, _ := withBatch.exactCount.Arguments()
	wantRecordSQL := `SELECT "fld_root_id", "fld_root_label" FROM ` +
		`"frm_include_root" WHERE ("fld_root_label" COLLATE ` +
		`pg_catalog.pg_unicode_fast = $1) ORDER BY "fld_root_label" COLLATE ` +
		`pg_catalog.pg_unicode_fast ASC NULLS LAST, "fld_root_id" ASC NULLS LAST LIMIT $2`
	wantCountSQL := `SELECT COUNT(*) FROM "frm_include_root" WHERE ` +
		`("fld_root_label" COLLATE pg_catalog.pg_unicode_fast = $1)`
	if withoutRecordSQL != wantRecordSQL || withRecordSQL != wantRecordSQL ||
		withoutCountSQL != wantCountSQL || withCountSQL != wantCountSQL ||
		!reflect.DeepEqual(withoutRecordArgs, []any{"root-filter", 3}) ||
		!reflect.DeepEqual(withRecordArgs, withoutRecordArgs) ||
		!reflect.DeepEqual(withoutCountArgs, []any{"root-filter"}) ||
		!reflect.DeepEqual(withCountArgs, withoutCountArgs) ||
		withoutBatch.ValidatePlan(without.plan) != nil ||
		withBatch.ValidatePlan(with.plan) != nil {
		t.Fatal("include state changed root record/count SQL semantics")
	}
	direct, err := compiler.CompileExactCount(with.plan)
	directSQL, sqlErr := direct.SQL()
	directArgs, argsErr := direct.Arguments()
	if err != nil || sqlErr != nil || argsErr != nil || directSQL != wantCountSQL ||
		!reflect.DeepEqual(directArgs, withoutCountArgs) ||
		direct.ValidatePlan(with.plan) != nil {
		t.Fatal("direct root exact-count did not ignore include state")
	}
}
