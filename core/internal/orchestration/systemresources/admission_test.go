package systemresources

import (
	"errors"
	"testing"

	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
)

func TestAdmissionRequiresExactPairedGenerationAndTenant(t *testing.T) {
	tenant, _ := publicresource.ParseID("10000000-0000-4000-8000-000000000001")
	otherTenant, _ := publicresource.ParseID("20000000-0000-4000-8000-000000000002")
	actorID, _ := publicresource.ParseID("30000000-0000-4000-8000-000000000003")
	actor, _ := publicresource.NewActorEvidence(tenant, actorID, 1, 1)

	admission, err := NewAdmission(tenant, actor, 7, 7, 7)
	if err != nil || admission.Validate() != nil || AuthorityGroup != "iam_system_resources" {
		t.Fatalf("valid admission error=%v", err)
	}
	for name, input := range map[string]struct {
		tenant             publicresource.ID
		group, iam, layout int64
	}{
		"foreign_tenant": {otherTenant, 7, 7, 7},
		"iam_drift":      {tenant, 7, 6, 7},
		"layout_drift":   {tenant, 7, 7, 8},
		"zero":           {tenant, 0, 0, 0},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := NewAdmission(input.tenant, actor, input.group, input.iam, input.layout)
			if !errors.Is(err, ErrContract) {
				t.Fatalf("error=%v", err)
			}
		})
	}
}
