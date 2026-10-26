package postrestore

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

type issuedChallengeFixture struct {
	value []byte
}

func (value *issuedChallengeFixture) RevealOnce(consume func([]byte) error) error {
	return consume(value.value)
}
func (*issuedChallengeFixture) Destroy() {}

type challengeGeneratorFixture struct{}

func (challengeGeneratorFixture) Generate() (IssuedChallenge, error) {
	return &issuedChallengeFixture{value: []byte("challenge")}, nil
}

type assertionVerifierFixture struct{}

func (assertionVerifierFixture) Verify(
	[]byte, AssertionExpectation,
) (AssertionResult, error) {
	return AssertionResult{UserVerified: true, SignCount: 2}, nil
}

type passwordIssuerFixture struct{}

func (passwordIssuerFixture) Issue(
	context.Context, []byte, []string,
) (PasswordCredential, error) {
	return PasswordCredential{Salt: [16]byte{1}, Tag: [32]byte{1},
		MemoryKiB: 65536, Iterations: 3, Parallelism: 1}, nil
}

type postRestoreStoreFixture struct {
	binding       Binding
	context       ChallengeContext
	issuedCommand ChallengeCommand
	failures      int
}

func (store *postRestoreStoreFixture) Admission(context.Context) (Admission, error) {
	return Admission{}, nil
}

func (store *postRestoreStoreFixture) Binding(
	context.Context, uuid.UUID, uuid.UUID,
) (Binding, error) {
	return store.binding, nil
}

func (store *postRestoreStoreFixture) PersistChallenge(
	_ context.Context, command ChallengeCommand,
) (ChallengeRecord, error) {
	store.issuedCommand = command
	return ChallengeRecord{OperationID: command.OperationID,
		TenantID: store.binding.TenantID, OwnerUserID: store.binding.OwnerUserID,
		AnchorID: store.binding.AnchorID, Challenge: command.Challenge,
		RPID: command.RPID, CredentialID: []byte{1}, PublicKeyCOSE: []byte{2},
		ExpiresAt: command.ExpiresAt}, nil
}

func (store *postRestoreStoreFixture) ChallengeContext(
	context.Context, uuid.UUID, uuid.UUID, uuid.UUID, string, time.Time,
) (ChallengeContext, error) {
	return store.context, nil
}

func (store *postRestoreStoreFixture) FailChallenge(
	context.Context, uuid.UUID, uuid.UUID, uuid.UUID, time.Time,
) error {
	store.failures++
	return nil
}

func (*postRestoreStoreFixture) PersistReactivation(
	context.Context, CompleteCommand,
) error {
	return nil
}

func (*postRestoreStoreFixture) Reconcile(
	context.Context, ReconcileCommand,
) (ReconcileResult, error) {
	return ReconcileResult{}, nil
}

func newPostRestoreServiceFixture(
	t *testing.T, store *postRestoreStoreFixture, now time.Time, operationID uuid.UUID,
) *Service {
	t.Helper()
	security := Security{
		Challenges: challengeGeneratorFixture{},
		Assertions: assertionVerifierFixture{},
		Passwords:  passwordIssuerFixture{},
	}
	service, err := NewService(store, security,
		DirectoryKey{ID: "directory.v1", Material: [32]byte{1}},
		[]string{"https://console.example.com"}, func() time.Time { return now },
		func() (uuid.UUID, error) { return operationID, nil })
	if err != nil {
		t.Fatal(err)
	}
	return service
}
