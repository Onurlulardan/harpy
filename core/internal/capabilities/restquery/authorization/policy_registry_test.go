package authorization

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type registryPolicyPort struct{}

func (*registryPolicyPort) EvaluateQuery(
	context.Context,
	transaction.TenantScope,
	PolicyRequest,
) (PolicyDecision, error) {
	return PolicyDecision{}, nil
}

func TestPolicyRegistryIsFixedCanonicalAndRedacted(t *testing.T) {
	registry, err := NewPolicyRegistry([]PolicyRegistration{
		{Owner: "workflow", Port: &registryPolicyPort{}},
		{Owner: "forms", Port: &registryPolicyPort{}},
	})
	if err != nil {
		t.Fatal(err)
	}
	owners, ownersErr := registry.Owners()
	if ownersErr != nil || len(owners) != 2 || owners[0] != "forms" ||
		owners[1] != "workflow" {
		t.Fatal("policy registry owners are not canonical")
	}
	owners[0] = "tampered"
	if _, found := registry.lookup("forms"); !found || registry.Validate() != nil {
		t.Fatal("returned owners mutated policy registry")
	}
	if fmt.Sprintf("%v", registry) != authorizationRedaction ||
		registry.LogValue().String() != authorizationRedaction {
		t.Fatal("policy registry was not redacted")
	}
	if _, err := json.Marshal(registry); err == nil {
		t.Fatal("policy registry allowed direct JSON marshal")
	}
}

func TestPolicyRegistryRejectsDuplicateInvalidAndTypedNilPorts(t *testing.T) {
	var nilPort *registryPolicyPort
	cases := [][]PolicyRegistration{
		nil,
		{{Owner: "Forms", Port: &registryPolicyPort{}}},
		{{Owner: "forms", Port: nilPort}},
		{
			{Owner: "forms", Port: &registryPolicyPort{}},
			{Owner: "forms", Port: &registryPolicyPort{}},
		},
	}
	for index, registrations := range cases {
		if _, err := NewPolicyRegistry(registrations); err == nil {
			t.Fatalf("invalid registry %d was accepted", index)
		}
	}
}
