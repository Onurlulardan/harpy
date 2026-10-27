package transaction

import (
	"testing"

	"github.com/google/uuid"
)

func TestAdmitCommandScopeSealsActorAndAuthority(t *testing.T) {
	request, err := AdmitCommandScope(
		uuid.New(), uuid.New(), "iam.identity.user.create",
		"iam_system_resources", 7,
		map[string]int64{"iam": 7, "system_resource_layout": 7},
	)
	if err != nil || !request.valid() || request.Mode != Command {
		t.Fatalf("command admission=%+v error=%v", request, err)
	}
}

func TestAdmitMultiGroupCommandScopeSealsExactParticipants(t *testing.T) {
	groups, domains, participants := validMultiGroupAuthorityInput()
	request, err := AdmitMultiGroupCommandScope(
		uuid.New(), uuid.New(), "workflow.forms.set-field",
		groups, domains, participants,
	)
	if err != nil || !request.valid() || request.Mode != Command ||
		!request.Operation.allowsOwner("workflow") ||
		!request.Operation.allowsOwner("forms") ||
		request.Operation.allowsOwner("billing") {
		t.Fatalf("multi-group admission=%+v error=%v", request, err)
	}
}
