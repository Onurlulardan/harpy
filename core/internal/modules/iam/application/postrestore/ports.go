package postrestore

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Store interface {
	Admission(context.Context) (Admission, error)
	Binding(context.Context, uuid.UUID, uuid.UUID) (Binding, error)
	PersistChallenge(context.Context, ChallengeCommand) (ChallengeRecord, error)
	ChallengeContext(context.Context, uuid.UUID, uuid.UUID, uuid.UUID,
		string, time.Time) (ChallengeContext, error)
	FailChallenge(context.Context, uuid.UUID, uuid.UUID, uuid.UUID, time.Time) error
	PersistReactivation(context.Context, CompleteCommand) error
	Reconcile(context.Context, ReconcileCommand) (ReconcileResult, error)
}

type UUIDGenerator func() (uuid.UUID, error)
type Clock func() time.Time

type DirectoryKey struct {
	ID       string
	Material [32]byte
}

type IssuedChallenge interface {
	RevealOnce(func([]byte) error) error
	Destroy()
}

type ChallengeGenerator interface {
	Generate() (IssuedChallenge, error)
}

type AssertionExpectation struct {
	Challenge                []byte
	CredentialID             []byte
	PublicKeyCOSE            []byte
	RPID, Origin             string
	UserHandle               []byte
	UserVerificationRequired bool
}

type AssertionResult struct {
	SignCount                   uint32
	BackupEligible, BackupState bool
	UserVerified                bool
}

type AssertionVerifier interface {
	Verify([]byte, AssertionExpectation) (AssertionResult, error)
}

type PasswordIssuer interface {
	Issue(context.Context, []byte, []string) (PasswordCredential, error)
}

type Security struct {
	Challenges ChallengeGenerator
	Assertions AssertionVerifier
	Passwords  PasswordIssuer
}
