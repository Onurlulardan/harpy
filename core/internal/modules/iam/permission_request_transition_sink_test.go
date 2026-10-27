package iam

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/permissionrequest"
	"github.com/google/uuid"
)

func TestPermissionRequestTransitionPayloadIsCanonical(t *testing.T) {
	record := permissionrequest.TransitionRecord{
		OperationID: uuid.New(), RequestSHA256: [32]byte{1},
		GroupID: uuid.New(), ActorID: uuid.New(),
		Command: permissionrequest.TransitionReview,
		From:    permissionrequest.StatePending, To: permissionrequest.StateInReview,
		ReviewTaskID: uuid.New(), Revision: 2,
	}
	payload, digest, err := permissionRequestTransitionPayload(record, 7)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"authorization_clock":7,"command":"iam.permission_request.review","from_state":"PENDING","revision":2,"to_state":"IN_REVIEW"}`
	if string(payload) != want || len(digest) != 64 {
		t.Fatalf("payload = %s / %s", payload, digest)
	}
}
