package authentication

import (
	"fmt"
	"log/slog"
	"time"
)

const initialChallengeRedacted = "[REDACTED IAM INITIAL PASSWORD]"

type InitialPasswordChallenge struct {
	material  *InitialPasswordMaterial
	expiresAt time.Time
}

func newInitialPasswordChallenge(
	material *InitialPasswordMaterial,
	expiresAt time.Time,
) (InitialPasswordChallenge, error) {
	if material == nil || !material.active || expiresAt.IsZero() {
		return InitialPasswordChallenge{}, ErrAuthenticationContract
	}
	return InitialPasswordChallenge{material: material, expiresAt: expiresAt}, nil
}

func (challenge InitialPasswordChallenge) ExpiresAt() time.Time { return challenge.expiresAt }

func (challenge *InitialPasswordChallenge) RevealOnce(
	consumer func(cookie, csrf []byte) error,
) error {
	if challenge == nil || challenge.material == nil {
		return ErrAuthenticationContract
	}
	err := challenge.material.RevealOnce(consumer)
	challenge.material = nil
	return err
}

func (challenge *InitialPasswordChallenge) Destroy() {
	if challenge == nil {
		return
	}
	challenge.material.Destroy()
	challenge.material = nil
	challenge.expiresAt = time.Time{}
}

func (InitialPasswordChallenge) String() string   { return initialChallengeRedacted }
func (InitialPasswordChallenge) GoString() string { return initialChallengeRedacted }
func (InitialPasswordChallenge) LogValue() slog.Value {
	return slog.StringValue(initialChallengeRedacted)
}
func (InitialPasswordChallenge) MarshalJSON() ([]byte, error) {
	return []byte(`"[REDACTED IAM INITIAL PASSWORD]"`), nil
}
func (InitialPasswordChallenge) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte(initialChallengeRedacted))
}
