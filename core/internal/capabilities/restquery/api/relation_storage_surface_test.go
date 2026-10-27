package api_test

import (
	"reflect"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestRelationStorageExportsOnlyOpaqueMappingSurface(t *testing.T) {
	typeOfStorage := reflect.TypeFor[api.RelationStorage]()
	for index := 0; index < typeOfStorage.NumField(); index++ {
		if typeOfStorage.Field(index).IsExported() {
			t.Fatal("relation storage exposes a mutable field")
		}
	}
	for _, forbidden := range []string{
		"SourceColumn", "TargetColumn", "SourceID", "TargetID", "RelationID",
	} {
		if _, found := typeOfStorage.MethodByName(forbidden); found {
			t.Fatalf("relation storage exposes compiler-owned member %s", forbidden)
		}
	}
	for _, required := range []string{"Validate", "Table", "FingerprintSHA256"} {
		if _, found := typeOfStorage.MethodByName(required); !found {
			t.Fatalf("relation storage is missing sealed method %s", required)
		}
	}
}
