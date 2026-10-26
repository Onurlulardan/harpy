package publication

import (
	"context"
	"errors"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"testing"
)

func TestCreateRequiresIdempotencyKey(t *testing.T) {
	_, err := (&Application{}).Create(context.Background(), platformtx.TenantScope{}, CreateRequest{ResourceKind: ResourceQuery, ResourceVersionID: publicationFixtureVersion, Audience: "public", CachePolicy: "none"})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("create error = %v", err)
	}
}
