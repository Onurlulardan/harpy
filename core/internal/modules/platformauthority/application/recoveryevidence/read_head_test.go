package recoveryevidence

import (
	"context"
	"testing"

	evidenceports "github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports/recoveryevidence"
)

type fixtureStore struct{ calls int }

func (store *fixtureStore) ReadRecoveryHead(
	context.Context,
	string,
) (evidenceports.Head, error) {
	store.calls++
	return evidenceports.Head{EnvironmentID: "environment"}, nil
}

func TestReadHeadDelegatesBoundRequest(t *testing.T) {
	store := &fixtureStore{}
	service, err := New(store)
	if err != nil {
		t.Fatal(err)
	}
	head, err := service.ReadRecoveryHead(t.Context(), "environment")
	if err != nil || head.EnvironmentID != "environment" || store.calls != 1 {
		t.Fatalf("head=%v calls=%d err=%v", head, store.calls, err)
	}
}
