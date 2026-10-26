package iam

import (
	"testing"

	"github.com/google/uuid"
)

func TestIdentityStateContractBindsStateWithoutChangingReplayOperation(t *testing.T) {
	fixture := newAuthHTTPTestFixture(t)
	input := IdentityStateMutation{
		IdentityMutation: IdentityMutation{
			ID:               uuid.MustParse("77000000-0000-4000-8000-000000000001"),
			ExpectedRevision: 2, IdempotencyKey: "stable-key",
		},
		State: "INACTIVE",
	}
	operation, digest, err := identityStateContract(
		fixture.claims, "iam.identity.user.disable", input,
	)
	if err != nil || operation == uuid.Nil || digest == [32]byte{} {
		t.Fatalf("operation=%s digest=%x error=%v", operation, digest, err)
	}
	changed := input
	changed.State = "ACTIVE"
	changedOperation, changedDigest, changedErr := identityStateContract(
		fixture.claims, "iam.identity.user.disable", changed,
	)
	if changedErr != nil || changedOperation != operation || changedDigest == digest {
		t.Fatal("state is not bound only to the replay request digest")
	}
}
