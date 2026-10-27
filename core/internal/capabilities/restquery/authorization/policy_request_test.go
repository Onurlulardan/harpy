package authorization

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
)

func TestPolicyRequestIsCanonicalImmutableAndRedacted(t *testing.T) {
	request := mustPolicyRequest(t)
	binding, bindingErr := request.BindingSHA256()
	tenantID, tenantErr := request.TenantID()
	actorID, actorErr := request.ActorID()
	queryIdentity, queryErr := request.QueryIdentitySHA256()
	owner, ownerErr := request.Owner()
	requirements, requirementsErr := request.Requirements()
	if bindingErr != nil || binding == [32]byte{} || tenantErr != nil || actorErr != nil ||
		queryErr != nil || queryIdentity == [32]byte{} || ownerErr != nil || owner != "forms" ||
		requirementsErr != nil || len(requirements) != 2 || tenantID == actorID {
		t.Fatal("policy request accessors are invalid")
	}
	leftID, _ := requirements[0].ID()
	rightID, _ := requirements[1].ID()
	if bytes.Compare(leftID[:], rightID[:]) >= 0 {
		t.Fatal("policy requirements are not canonically ordered")
	}
	requirements[0] = Requirement{}
	if request.Validate() != nil {
		t.Fatal("returned requirements mutated policy request")
	}
	if fmt.Sprintf("%v", request) != authorizationRedaction ||
		request.LogValue().String() != authorizationRedaction {
		t.Fatal("policy request was not redacted")
	}
	if _, err := json.Marshal(request); err == nil {
		t.Fatal("policy request allowed direct JSON marshal")
	}
}

func TestPolicyRequestRejectsIdentityTamper(t *testing.T) {
	request := mustPolicyRequest(t)
	tampered := request
	tampered.binding = [32]byte{1}
	if tampered.Validate() == nil {
		t.Fatal("tampered policy request binding was accepted")
	}
	tampered = request
	tampered.requirements[0], tampered.requirements[1] =
		tampered.requirements[1], tampered.requirements[0]
	if tampered.Validate() == nil {
		t.Fatal("non-canonical policy requirement order was accepted")
	}
}
