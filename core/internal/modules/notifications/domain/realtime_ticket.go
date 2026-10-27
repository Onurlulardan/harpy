package domain

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"net/url"
	"strings"
	"time"
)

const (
	RealtimeProtocol     = "harpy.notifications.v1"
	TicketProtocolPrefix = "fc-ticket."
	TicketLifetime       = time.Minute
)

var (
	ErrTicketInvalid  = errors.New("notification realtime ticket is invalid")
	ErrTicketConflict = errors.New("notification realtime ticket cannot be consumed")
)

type Ticket struct {
	ID, TenantID, UserID ID
	TokenHash            [32]byte
	AuthorizationVersion int64
	Origin               string
	CreatedAt, ExpiresAt time.Time
	ConsumedAt           *time.Time
}

type TicketConsume struct {
	TenantID, UserID     ID
	AuthorizationVersion int64
	Origin, Token        string
	At                   time.Time
}

func NewTicket(id, tenantID, userID ID, authorizationVersion int64,
	origin string, now time.Time, entropy io.Reader,
) (Ticket, string, error) {
	if !id.IsV7() || !tenantID.Valid() || !userID.Valid() ||
		authorizationVersion < 1 || !validHTTPSOrigin(origin) || now.IsZero() || entropy == nil {
		return Ticket{}, "", ErrTicketInvalid
	}
	raw := make([]byte, 32)
	if _, err := io.ReadFull(entropy, raw); err != nil {
		return Ticket{}, "", ErrTicketInvalid
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	return Ticket{ID: id, TenantID: tenantID, UserID: userID,
		TokenHash: sha256.Sum256([]byte(token)), AuthorizationVersion: authorizationVersion,
		Origin: origin, CreatedAt: now.UTC(), ExpiresAt: now.UTC().Add(TicketLifetime)}, token, nil
}

func (value Ticket) Consume(input TicketConsume) (Ticket, error) {
	if value.ConsumedAt != nil || input.At.IsZero() || !input.At.Before(value.ExpiresAt) ||
		input.At.Before(value.CreatedAt) || input.TenantID != value.TenantID ||
		input.UserID != value.UserID || input.AuthorizationVersion != value.AuthorizationVersion ||
		input.Origin != value.Origin || sha256.Sum256([]byte(input.Token)) != value.TokenHash {
		return Ticket{}, ErrTicketConflict
	}
	consumedAt := input.At.UTC()
	value.ConsumedAt = &consumedAt
	return value, nil
}

func TicketProtocol(token string) (string, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil || len(decoded) != 32 {
		return "", ErrTicketInvalid
	}
	return TicketProtocolPrefix + token, nil
}

func ParseOfferedProtocols(header string) (string, error) {
	var ticket string
	fixed := false
	for _, offered := range strings.Split(header, ",") {
		value := strings.TrimSpace(offered)
		switch {
		case value == RealtimeProtocol:
			fixed = true
		case strings.HasPrefix(value, TicketProtocolPrefix):
			if ticket != "" {
				return "", ErrTicketInvalid
			}
			ticket = strings.TrimPrefix(value, TicketProtocolPrefix)
		}
	}
	if !fixed || ticket == "" {
		return "", ErrTicketInvalid
	}
	if _, err := TicketProtocol(ticket); err != nil {
		return "", err
	}
	return ticket, nil
}

func validHTTPSOrigin(value string) bool {
	parsed, err := url.Parse(value)
	return err == nil && parsed.Scheme == "https" && parsed.Host != "" &&
		parsed.User == nil && parsed.RawQuery == "" && parsed.Fragment == "" && parsed.Path == ""
}
