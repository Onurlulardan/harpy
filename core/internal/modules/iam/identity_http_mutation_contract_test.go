package iam

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestIdentityMutationContractBindsRevisionAndIdempotency(t *testing.T) {
	fixture := newAuthHTTPTestFixture(t)
	input := IdentityMutation{
		ID:               uuid.MustParse("75000000-0000-4000-8000-000000000001"),
		ExpectedRevision: 4, IdempotencyKey: "stable-key",
	}
	operation, digest, err := identityMutationContract(
		fixture.claims, "iam.identity.user.delete", input,
	)
	if err != nil || operation == uuid.Nil || digest == [32]byte{} {
		t.Fatalf("operation=%s digest=%x error=%v", operation, digest, err)
	}
	replayOperation, replayDigest, replayErr := identityMutationContract(
		fixture.claims, "iam.identity.user.delete", input,
	)
	if replayErr != nil || replayOperation != operation || replayDigest != digest {
		t.Fatal("identical replay changed mutation identity")
	}
	changed := input
	changed.ExpectedRevision++
	changedOperation, changedDigest, _ := identityMutationContract(
		fixture.claims, "iam.identity.user.delete", changed,
	)
	if changedOperation != operation || changedDigest == digest {
		t.Fatal("revision changed operation identity or failed to change request digest")
	}
}

func TestIdentityMutationContractRejectsIncompleteInput(t *testing.T) {
	fixture := newAuthHTTPTestFixture(t)
	valid := IdentityMutation{
		ID:               uuid.MustParse("76000000-0000-4000-8000-000000000001"),
		ExpectedRevision: 1, IdempotencyKey: "stable-key",
	}
	for name, mutate := range map[string]func(*IdentityMutation){
		"id":       func(input *IdentityMutation) { input.ID = uuid.Nil },
		"revision": func(input *IdentityMutation) { input.ExpectedRevision = 0 },
		"key":      func(input *IdentityMutation) { input.IdempotencyKey = "" },
	} {
		t.Run(name, func(t *testing.T) {
			input := valid
			mutate(&input)
			_, _, err := identityMutationContract(
				fixture.claims, "iam.identity.user.delete", input,
			)
			if !errors.Is(err, ErrIdentityHTTPInvalid) {
				t.Fatalf("error=%v", err)
			}
		})
	}
}
