package compatibility

import (
	"context"
	"testing"
)

func TestResolveRequiresExactLegacyTuple(t *testing.T) {
	mapping := mappingFixture()
	application, _ := applicationFixture(t, mapping)
	result, err := application.Resolve(context.Background(), mapping.Key)
	if err != nil || result.Key != mapping.Key || result.PublicID != mapping.PublicID {
		t.Fatalf("result=%+v error=%v", result, err)
	}
	substituted := mapping.Key
	substituted.Slug = "other"
	if _, err = application.Resolve(context.Background(), substituted); err != ErrNotFound {
		t.Fatalf("substitution error=%v", err)
	}
}
