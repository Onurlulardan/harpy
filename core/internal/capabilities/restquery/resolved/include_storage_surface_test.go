package resolved_test

import (
	"reflect"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func TestIncludeStorageAccessorsAreUnambiguous(t *testing.T) {
	typeOfInclude := reflect.TypeFor[resolved.Include]()
	if _, found := typeOfInclude.MethodByName("Storage"); found {
		t.Fatal("resolved Include retained ambiguous Storage accessor")
	}
	for _, required := range []string{"TargetStorage", "RelationStorage"} {
		if _, found := typeOfInclude.MethodByName(required); !found {
			t.Fatalf("resolved Include is missing %s", required)
		}
	}
}
