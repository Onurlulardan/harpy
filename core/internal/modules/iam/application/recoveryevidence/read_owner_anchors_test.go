package recoveryevidence

import (
	"context"
	"testing"

	evidenceports "github.com/Onurlulardan/harpy/core/internal/modules/iam/ports/recoveryevidence"
)

type fixtureStore struct{ calls int }

func (store *fixtureStore) ReadOwnerAnchors(
	context.Context,
	string,
) ([]evidenceports.Anchor, error) {
	store.calls++
	return []evidenceports.Anchor{{EnvironmentID: "environment"}}, nil
}

func TestReadOwnerAnchorsDelegatesBoundRequest(t *testing.T) {
	store := &fixtureStore{}
	service, err := New(store)
	if err != nil {
		t.Fatal(err)
	}
	anchors, err := service.ReadOwnerAnchors(t.Context(), "environment")
	if err != nil || len(anchors) != 1 || store.calls != 1 {
		t.Fatalf("anchors=%v calls=%d err=%v", anchors, store.calls, err)
	}
}
