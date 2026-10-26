//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"

	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
)

var concurrentLayoutIDs = []string{
	"10000000-0000-7000-8000-000000000001",
	"10000000-0000-7000-8000-000000000002",
	"10000000-0000-7000-8000-000000000003",
	"10000000-0000-7000-8000-000000000004",
	"10000000-0000-7000-8000-000000000005",
	"10000000-0000-7000-8000-000000000006",
	"10000000-0000-7000-8000-000000000007",
	"10000000-0000-7000-8000-000000000008",
}

type concurrentLayoutResult struct {
	version int64
	err     error
}

func assertLayoutVersionConcurrency(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	results := make(chan concurrentLayoutResult, len(concurrentLayoutIDs))
	for _, id := range concurrentLayoutIDs {
		go func(revisionID string) {
			document, err := runLayoutRuntimeText(ctx, database, "layout_a", createSQL,
				revisionID, string(layoutapi.ResourceOrganization), adapterDefinition,
				layoutDigest(adapterDefinition), layoutNow(20))
			if err != nil {
				results <- concurrentLayoutResult{err: err}
				return
			}
			revision, err := decodeRevision([]byte(document))
			results <- concurrentLayoutResult{version: revision.Version, err: err}
		}(id)
	}
	versions := make(map[int64]struct{}, len(concurrentLayoutIDs))
	for range concurrentLayoutIDs {
		result := <-results
		if result.err != nil {
			t.Fatal(result.err)
		}
		versions[result.version] = struct{}{}
	}
	if len(versions) != len(concurrentLayoutIDs) {
		t.Fatalf("concurrent layout versions = %v", versions)
	}
	for version := int64(2); version <= int64(len(concurrentLayoutIDs)+1); version++ {
		if _, exists := versions[version]; !exists {
			t.Fatalf("concurrent layout version %d is absent", version)
		}
	}
}
