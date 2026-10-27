package postgres

import (
	"context"
	"testing"

	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
)

func TestGetRejectsDigestDrift(t *testing.T) {
	document := fixtureRevisionJSON(t, layoutapi.RevisionDraft)
	document[len(document)-3] ^= 1
	store := fixtureLayoutStore(fixtureRow{document: document})
	if _, err := store.Get(context.Background(), adapterRevisionID); err == nil {
		t.Fatal("corrupt revision was accepted")
	}
}
