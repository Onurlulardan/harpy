package postgres

import (
	"reflect"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/schemacontract"
)

func TestOpenRequiresRegistryAuthorizedSchemaContract(t *testing.T) {
	t.Parallel()
	openType := reflect.TypeOf(Open)
	want := reflect.TypeOf(schemacontract.EmbeddedContract{})
	if openType.NumIn() != 4 || openType.In(3) != want {
		t.Fatalf("Open schema authority parameter = %v", openType)
	}
}
