package query

import (
	"context"
	"errors"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"testing"
)

func TestRevalidateRejectsMissingDependencies(t *testing.T) {
	_, err := (&Application{}).Revalidate(context.Background(), platformtx.TenantScope{},
		PublishRequest{QueryID: queryFixtureID, DraftVersionID: queryFixtureVersion,
			ExpectedRootRevision: 1, ExpectedDraftRevision: 1, IdempotencyKey: "revalidate-1"})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("revalidate error = %v", err)
	}
}
