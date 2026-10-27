package iam

import (
	"context"
	"errors"
	"testing"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func TestLayoutPublishCatalogIntentIsExact(t *testing.T) {
	tests := []struct {
		name      string
		mode      platformtx.ScopeMode
		operation string
		allowed   bool
	}{
		{"exact", platformtx.Command, publicresource.LayoutPublishCatalogOperation, true},
		{"snapshot", platformtx.SnapshotQuery, "system_resource_layout.publish", false},
		{"generic query", platformtx.SnapshotQuery, "restquery.execute", false},
		{"other layout command", platformtx.Command, "system_resource_layout.update", false},
		{"empty", platformtx.Command, "", false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if actual := identitySystemResourceLayoutPublishIntent(
				test.mode, test.operation,
			); actual != test.allowed {
				t.Fatalf("intent=%v want=%v", actual, test.allowed)
			}
		})
	}
}

func TestProductionCatalogPortRejectsUnownedLayoutPublishScope(t *testing.T) {
	tenant, _ := publicresource.ParseID("018f47f4-7b1b-7cc0-8000-000000000001")
	actorID, _ := publicresource.ParseID("018f47f4-7b1b-7cc0-8000-000000000002")
	actor, err := publicresource.NewActorEvidence(tenant, actorID, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	resourcePort := (&identityHTTPApplication{}).SystemResources()
	port, ok := resourcePort.(publicresource.LayoutCatalogPort)
	if !ok {
		t.Fatal("production IAM port does not expose the layout catalog contract")
	}
	_, err = port.CatalogForLayoutPublish(
		context.Background(), platformtx.TenantScope{}, actor, publicresource.Users,
	)
	if !errors.Is(err, publicresource.ErrContract) {
		t.Fatalf("unowned publish scope error=%v", err)
	}
}
