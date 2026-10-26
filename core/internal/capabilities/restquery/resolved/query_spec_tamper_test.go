package resolved

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestQuerySpecTamperAndAccessorsFailClosed(t *testing.T) {
	original := querySpecFixture(t)
	if err := original.Validate(); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name   string
		tamper func(*QuerySpec)
	}{
		{"identity size", func(spec *QuerySpec) { spec.identitySize++ }},
		{"state size", func(spec *QuerySpec) { spec.stateSize++ }},
		{"identity hash", func(spec *QuerySpec) { spec.identityHash[0] ^= 1 }},
		{"state seal", func(spec *QuerySpec) { spec.stateSeal[0] ^= 1 }},
		{"catalog seal", func(spec *QuerySpec) { spec.catalogSeal[0] ^= 1 }},
		{"catalog set", func(spec *QuerySpec) { spec.catalogs = api.CatalogSet{} }},
		{"mode state", func(spec *QuerySpec) { spec.aggregate = &AggregateQuery{} }},
		{"sort provenance", func(spec *QuerySpec) {
			spec.record.sorts[0].implicit = false
		}},
		{"cursor state", func(spec *QuerySpec) {
			spec.record.page.after = "different"
			spec.record.page.hasAfter = true
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			spec := original
			record := *original.record
			record.sorts = append([]Sort{}, original.record.sorts...)
			spec.record = &record
			test.tamper(&spec)
			if spec.Validate() == nil {
				t.Fatal("tampered QuerySpec validated")
			}
			if _, err := spec.IdentityHash(); err == nil {
				t.Fatal("tampered QuerySpec accessor did not fail closed")
			}
		})
	}
}

func TestQuerySpecConstructorRejectsNonCanonicalAuthorizationPath(t *testing.T) {
	original := querySpecFixture(t)
	record := *original.record
	record.projections = append([]Projection{}, record.projections...)
	badPath, _ := NewRequestPath("/secret")
	record.projections[0].path = badPath
	if record.Validate() != nil {
		t.Fatal("component path should be checked at QuerySpec structural binding")
	}
	if _, err := NewRecordQuerySpec(
		original.catalogs, original.reference, original.storage,
		original.stableID, original.rootPath, record,
	); err == nil {
		t.Fatal("QuerySpec accepted a non-canonical authorization path")
	}
}

func TestAggregateMeasureRejectsNonCanonicalAlias(t *testing.T) {
	path, _ := NewRequestPath("/aggregate/measures/0")
	key, _ := NewOutputKey("Not_A_Measure")
	if _, err := NewAggregateMeasure(
		key, api.AggregateFunctionCount, nil, path,
	); err == nil {
		t.Fatal("aggregate measure accepted a non-canonical output alias")
	}
}
