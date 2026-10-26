package postgres

import (
	"context"
	"testing"

	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
)

func TestListDecodesRevisionArray(t *testing.T) {
	revision := fixtureRevisionJSON(t, layoutapi.RevisionDraft)
	document := append(append([]byte{'['}, revision...), ']')
	store := fixtureLayoutStore(fixtureRow{document: document})
	result, err := store.List(context.Background(), layoutapi.ResourceUser)
	if err != nil || len(result) != 1 || result[0].ID != adapterRevisionID.String() {
		t.Fatalf("list result = %+v: %v", result, err)
	}
}
