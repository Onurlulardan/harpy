package actorbridgeissue

import (
	"context"
	"testing"

	actorassertionv1 "github.com/Onurlulardan/harpy/contracts/actorassertion/v1"
)

type unusedStore struct{ called bool }

func (store *unusedStore) Resolve(
	context.Context,
	TenantScope,
	actorassertionv1.Domain,
) (Evidence, error) {
	store.called = true
	return NewEvidence(1, 1, 1)
}

func TestResolveRejectsUnownedScopeBeforeStore(t *testing.T) {
	store := &unusedStore{}
	resolver, err := NewResolver(store)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := resolver.Resolve(
		context.Background(), TenantScope{}, actorassertionv1.DomainDocuments,
	); err == nil {
		t.Fatal("actor bridge issue resolved outside an owned scope")
	}
	if store.called {
		t.Fatal("actor bridge issue store ran outside an owned scope")
	}
}
