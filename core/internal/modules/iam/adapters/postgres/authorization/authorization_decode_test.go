package postgres

import (
	"errors"
	"strings"
	"testing"
)

func TestDecodeAuthorizationSnapshotRejectsUnknownOrIncompleteEvidence(t *testing.T) {
	validActor := []byte(`{"id":"11000000-0000-4000-8000-000000000001","status":"active","token_version":7,"authorization_version":9}`)
	validResource := []byte(`{"id":"12000000-0000-4000-8000-000000000001","state":"active","protected":true}`)
	validAction := []byte(`{"id":"13000000-0000-4000-8000-000000000001","state":"active","protected":false}`)
	if _, err := decodeAuthorizationSnapshot(
		append(validActor[:len(validActor)-1], []byte(`,"unknown":true}`)...),
		validResource, validAction, 1, []byte(`[]`), []byte(`[]`),
	); !errors.Is(err, errAuthorizationRead) {
		t.Fatalf("unknown evidence error = %v", err)
	}
	if _, err := decodeAuthorizationSnapshot(
		validActor, validResource, validAction, -1, []byte(`[]`), []byte(`[]`),
	); !errors.Is(err, errAuthorizationRead) {
		t.Fatalf("missing clock error = %v", err)
	}
}

func TestDecodeAuthorizationSnapshotAcceptsProtectedTenantAllBinding(t *testing.T) {
	snapshot, err := decodeAuthorizationSnapshot(
		[]byte(`{"id":"11000000-0000-4000-8000-000000000001","status":"active","token_version":2,"authorization_version":1}`),
		[]byte(`{"id":"12000000-0000-4000-8000-000000000001","state":"active","protected":false}`),
		[]byte(`{"id":"13000000-0000-4000-8000-000000000001","state":"active","protected":false}`),
		1,
		[]byte(`[{
            "path":"role",
            "binding_id":"15000000-0000-4000-8000-000000000001",
            "binding_state":"active",
            "permission_id":"14000000-0000-4000-8000-000000000001",
            "permission_match_kind":"tenant_all",
            "permission_state":"active",
            "permission_protected":true,
            "permission_resource":null,
            "permission_action":null,
            "resource_state":"",
            "action_state":"",
            "assignment_id":"17000000-0000-4000-8000-000000000001",
            "assignment_state":"active",
            "role_id":"16000000-0000-4000-8000-000000000001",
            "role_status":"active",
            "role_scope_kind":"tenant",
            "role_protected":true,
            "membership_id":null,
            "membership_state":"",
            "organization_id":null,
            "organization_status":"",
            "policy_role_id":"16000000-0000-4000-8000-000000000001",
            "policy_permission_id":"14000000-0000-4000-8000-000000000001",
            "policy_binding_id":"15000000-0000-4000-8000-000000000001",
            "policy_state":"active"
        }]`),
		[]byte(`[]`),
	)
	if err != nil || snapshot.Validate() != nil {
		t.Fatalf("tenant_all snapshot err=%v validate=%v", err, snapshot.Validate())
	}
}

func TestDecodeAuthorizationSnapshotWrapsUnknownFieldCause(t *testing.T) {
	_, err := decodeAuthorizationSnapshot(
		[]byte(`{"id":"11000000-0000-4000-8000-000000000001","status":"active","token_version":7,"authorization_version":9,"unknown":true}`),
		[]byte(`{"id":"12000000-0000-4000-8000-000000000001","state":"active","protected":true}`),
		[]byte(`{"id":"13000000-0000-4000-8000-000000000001","state":"active","protected":false}`),
		1, []byte(`[]`), []byte(`[]`),
	)
	if !errors.Is(err, errAuthorizationRead) || !strings.Contains(err.Error(), "unknown") {
		t.Fatalf("unknown field cause missing: %v", err)
	}
}
