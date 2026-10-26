package actorbridgeissue

import (
	"context"
	"testing"

	actorassertionv1 "github.com/Onurlulardan/harpy/contracts/actorassertion/v1"
	actorbridgeissueapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/actorbridgeissue"
)

func TestResolveRejectsUnownedScope(t *testing.T) {
	configured, err := NewStore()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := configured.Resolve(context.Background(),
		actorbridgeissueapp.TenantScope{}, actorassertionv1.DomainForms); err == nil {
		t.Fatal("actor bridge issue store accepted an unowned scope")
	}
}
