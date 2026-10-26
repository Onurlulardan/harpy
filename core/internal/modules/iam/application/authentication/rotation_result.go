package authentication

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

type Rotation struct {
	outcome         RefreshOutcome
	accessToken     string
	refreshToken    IssuedRefresh
	sessionID       uuid.UUID
	absoluteExpires time.Time
	valid           bool
}

func newTerminalRotation(outcome RefreshOutcome) (Rotation, error) {
	if outcome != RefreshReusedOutcome && outcome != RefreshExpiredOutcome &&
		outcome != RefreshRejectedOutcome {
		return Rotation{}, ErrAuthenticationContract
	}
	return Rotation{outcome: outcome, valid: true}, nil
}

func newRotatedRotation(
	access string,
	refresh IssuedRefresh,
	sessionID uuid.UUID,
	absoluteExpires time.Time,
) (Rotation, error) {
	if access == "" || nilInterface(refresh) || sessionID == uuid.Nil ||
		absoluteExpires.IsZero() {
		return Rotation{}, ErrAuthenticationContract
	}
	return Rotation{
		outcome: RefreshRotatedOutcome, accessToken: access,
		refreshToken: refresh, sessionID: sessionID,
		absoluteExpires: absoluteExpires.UTC(), valid: true,
	}, nil
}

func (rotation Rotation) Outcome() RefreshOutcome { return rotation.outcome }

func (rotation Rotation) AccessToken() (string, error) {
	if !rotation.valid || rotation.outcome != RefreshRotatedOutcome ||
		rotation.accessToken == "" {
		return "", ErrAuthenticationContract
	}
	return rotation.accessToken, nil
}

func (rotation Rotation) SessionID() uuid.UUID { return rotation.sessionID }

func (rotation Rotation) AbsoluteExpiresAt() time.Time { return rotation.absoluteExpires }

func (rotation *Rotation) RevealRefreshOnce(consume func([]byte) error) error {
	if rotation == nil || !rotation.valid || rotation.outcome != RefreshRotatedOutcome ||
		nilInterface(rotation.refreshToken) {
		return ErrAuthenticationContract
	}
	err := rotation.refreshToken.RevealOnce(consume)
	rotation.refreshToken = nil
	return err
}

func (rotation *Rotation) Destroy() {
	if rotation == nil {
		return
	}
	if !nilInterface(rotation.refreshToken) {
		rotation.refreshToken.Destroy()
	}
	rotation.refreshToken, rotation.accessToken, rotation.valid = nil, "", false
}

func (Rotation) String() string   { return authenticationRedacted }
func (Rotation) GoString() string { return authenticationRedacted }
func (Rotation) LogValue() slog.Value {
	return slog.StringValue(authenticationRedacted)
}
func (Rotation) MarshalJSON() ([]byte, error) {
	return []byte(`"[REDACTED IAM AUTHENTICATION]"`), nil
}
func (Rotation) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte(authenticationRedacted))
}
