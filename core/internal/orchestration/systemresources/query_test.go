package systemresources

import (
	"context"
	"testing"
)

func TestQueryExecutesIAMAndLayoutInOneSnapshotScope(t *testing.T) {
	fixture := newTestApplication(t)
	result, err := fixture.application.Query(
		context.Background(), fixture.fixture.admission, testQueryInput(t),
	)
	if err != nil || result.Validate() != nil {
		t.Fatalf("query error=%v", err)
	}
	fixture.assertOneCommittedScope(t)
}
