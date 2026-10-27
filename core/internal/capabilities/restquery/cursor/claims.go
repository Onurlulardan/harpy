package cursor

import (
	"fmt"
	"log/slog"
	"time"
)

type Claims struct {
	issuedAt  time.Time
	expiresAt time.Time
	last      []LastValue
	valid     bool
}

func newClaims(issuedAt, expiresAt time.Time, last []LastValue) (Claims, error) {
	claims := Claims{
		issuedAt: issuedAt, expiresAt: expiresAt,
		last: append([]LastValue(nil), last...), valid: true,
	}
	if err := claims.Validate(); err != nil {
		return Claims{}, err
	}
	return claims, nil
}

func (claims Claims) Validate() error {
	if !claims.valid || claims.issuedAt.IsZero() ||
		claims.expiresAt.Sub(claims.issuedAt) != cursorLifetime ||
		len(claims.last) < 1 || len(claims.last) > maximumSorts {
		return errInvalidClaims
	}
	for _, value := range claims.last {
		if value.Validate() != nil {
			return errInvalidClaims
		}
	}
	return nil
}

func (claims Claims) IssuedAt() (time.Time, error) {
	if err := claims.Validate(); err != nil {
		return time.Time{}, err
	}
	return claims.issuedAt, nil
}

func (claims Claims) ExpiresAt() (time.Time, error) {
	if err := claims.Validate(); err != nil {
		return time.Time{}, err
	}
	return claims.expiresAt, nil
}

func (claims Claims) LastValues() ([]LastValue, error) {
	if err := claims.Validate(); err != nil {
		return nil, err
	}
	return append([]LastValue(nil), claims.last...), nil
}

func (Claims) String() string                        { return redacted }
func (Claims) GoString() string                      { return redacted }
func (Claims) LogValue() slog.Value                  { return redactedLogValue() }
func (claims Claims) Format(state fmt.State, _ rune) { writeRedaction(state) }
func (Claims) MarshalJSON() ([]byte, error)          { return rejectJSONMarshal() }
