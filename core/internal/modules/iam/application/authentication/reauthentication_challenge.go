package authentication

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type ReauthenticationChallenge struct {
	ID                       uuid.UUID
	Method                   ReauthenticationMethod
	ExpiresAt                time.Time
	CredentialID             []byte
	RPID                     string
	UserVerificationRequired bool
	material                 IssuedChallenge
	valid                    bool
}

func (challenge *ReauthenticationChallenge) RevealOnce(
	consume func([]byte) error,
) error {
	if challenge == nil || !challenge.valid || nilInterface(challenge.material) {
		return ErrAuthenticationContract
	}
	err := challenge.material.RevealOnce(consume)
	challenge.material = nil
	return err
}

func (challenge *ReauthenticationChallenge) Destroy() {
	if challenge == nil {
		return
	}
	if !nilInterface(challenge.material) {
		challenge.material.Destroy()
	}
	challenge.material, challenge.valid = nil, false
}

func (ReauthenticationChallenge) String() string   { return authenticationRedacted }
func (ReauthenticationChallenge) GoString() string { return authenticationRedacted }
func (ReauthenticationChallenge) LogValue() slog.Value {
	return slog.StringValue(authenticationRedacted)
}
func (ReauthenticationChallenge) MarshalJSON() ([]byte, error) {
	return []byte(`"[REDACTED IAM AUTHENTICATION]"`), nil
}
func (ReauthenticationChallenge) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte(authenticationRedacted))
}
