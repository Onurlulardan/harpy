package authentication

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
)

const authenticationRedacted = "[REDACTED IAM AUTHENTICATION]"

type Grant struct {
	accessToken     string
	refreshToken    IssuedRefresh
	sessionID       uuid.UUID
	absoluteExpires time.Time
	valid           bool
}

func newGrant(
	accessToken string,
	refreshToken IssuedRefresh,
	sessionID uuid.UUID,
	absoluteExpires time.Time,
) (Grant, error) {
	if accessToken == "" || nilInterface(refreshToken) || sessionID == uuid.Nil ||
		absoluteExpires.IsZero() {
		return Grant{}, ErrAuthenticationContract
	}
	return Grant{
		accessToken: accessToken, refreshToken: refreshToken,
		sessionID: sessionID, absoluteExpires: absoluteExpires.UTC(), valid: true,
	}, nil
}

func (grant Grant) AccessToken() (string, error) {
	if !grant.valid || grant.accessToken == "" {
		return "", ErrAuthenticationContract
	}
	return grant.accessToken, nil
}

func (grant Grant) SessionID() uuid.UUID { return grant.sessionID }

func (grant Grant) AbsoluteExpiresAt() time.Time { return grant.absoluteExpires }

func (grant *Grant) RevealRefreshOnce(consume func([]byte) error) error {
	if grant == nil || !grant.valid || nilInterface(grant.refreshToken) {
		return ErrAuthenticationContract
	}
	err := grant.refreshToken.RevealOnce(consume)
	grant.refreshToken = nil
	return err
}

func (grant *Grant) Destroy() {
	if grant == nil {
		return
	}
	if !nilInterface(grant.refreshToken) {
		grant.refreshToken.Destroy()
	}
	grant.refreshToken, grant.accessToken, grant.valid = nil, "", false
}

func (Grant) String() string   { return authenticationRedacted }
func (Grant) GoString() string { return authenticationRedacted }
func (Grant) LogValue() slog.Value {
	return slog.StringValue(authenticationRedacted)
}
func (Grant) MarshalJSON() ([]byte, error) {
	return []byte(`"[REDACTED IAM AUTHENTICATION]"`), nil
}
func (Grant) Format(state fmt.State, _ rune) {
	_, _ = state.Write([]byte(authenticationRedacted))
}
