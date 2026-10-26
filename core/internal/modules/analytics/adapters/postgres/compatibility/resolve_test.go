package compatibility

import (
	"context"
	"testing"
)

func TestResolveBindsWholeLegacyLookupKey(t *testing.T) {
	mapping := legacyMapping()
	store, database := newStore(t, mappingRow(mapping))
	result, err := store.Resolve(context.Background(), mapping.Key)
	if err != nil || result.Key != mapping.Key || len(database.args) != 4 ||
		database.args[0] != mapping.Key.TenantIdentity || database.args[3] != mapping.Key.Slug {
		t.Fatalf("result=%+v args=%v error=%v", result, database.args, err)
	}
}
