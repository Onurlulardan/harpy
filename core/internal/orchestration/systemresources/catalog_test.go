package systemresources

import (
	"context"
	"testing"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
)

func TestCatalogResolvesIAMAndLayoutInOneSnapshotScope(t *testing.T) {
	fixture := newTestApplication(t)
	result, err := fixture.application.Catalog(
		context.Background(), fixture.fixture.admission, publicresource.Users,
	)
	if err != nil || result.Validate() != nil {
		t.Fatalf("catalog error=%v", err)
	}
	fixture.assertOneCommittedScope(t)
}
