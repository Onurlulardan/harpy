package compatibility

import (
	"context"
	"testing"
)

func TestAuthorizeLegacyResolvesExactMappingAndPersistsHit(t *testing.T) {
	mapping := mappingFixture()
	application, _ := applicationFixture(t, mapping)
	result, err := application.AuthorizeLegacy(context.Background(), mapping.Key)
	if err != nil || result.Key != mapping.Key || result.HitCount != mapping.HitCount+1 ||
		result.LastHitAt == nil || !result.LastHitAt.Equal(compatNow) {
		t.Fatalf("result=%+v err=%v", result, err)
	}
}
