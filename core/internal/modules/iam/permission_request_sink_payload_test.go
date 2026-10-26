package iam

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissionrequest"
	"github.com/google/uuid"
)

func TestPermissionRequestEventPayloadIsSecretFreeAndCanonical(t *testing.T) {
	record := permissionrequest.DecisionRecord{
		OperationID: uuid.New(), RequestSHA256: [32]byte{1},
		GroupID: uuid.New(), ReviewerID: uuid.New(), ReviewTaskID: uuid.New(),
		Decision: permissionrequest.DecisionApprove,
		Result: permissionrequest.DecisionResult{
			Code: "iam.permission_request.approved", Revision: 3,
			AuthorizationClock: 9,
		},
	}
	payload, digest, err := permissionRequestEventPayload(record)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"authorization_clock":9,"decision":"APPROVE","result":"iam.permission_request.approved","revision":3}`
	if string(payload) != want || len(digest) != 64 {
		t.Fatalf("payload = %s / %s", payload, digest)
	}
}
