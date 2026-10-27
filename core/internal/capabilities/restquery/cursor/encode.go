package cursor

import (
	"encoding/base64"
	"strings"
	"time"

	"github.com/Onurlulardan/harpy/internal/attestation"
)

func (codec Codec) Encode(context Context, lastValues []LastValue) (string, error) {
	if codec.Validate() != nil || context.Validate() != nil ||
		validateLastValues(context.sorts, lastValues) != nil {
		return "", errEncodeFailed
	}
	issuedAt := codec.state.clock().UTC().Truncate(time.Second)
	if context.scopeKind == "global" {
		issuedAt = context.snapshotAt.UTC().Truncate(time.Second)
	}
	if issuedAt.IsZero() || issuedAt.Unix() <= 0 {
		return "", errEncodeFailed
	}
	now := codec.state.clock().UTC()
	if issuedAt.After(now.Add(allowedClockSkew)) ||
		now.After(issuedAt.Add(cursorLifetime).Add(allowedClockSkew)) {
		return "", errEncodeFailed
	}
	claims, err := codec.buildWireClaims(context, lastValues, issuedAt)
	if err != nil {
		return "", errEncodeFailed
	}
	plaintext, err := attestation.MarshalCanonical(claims)
	if err != nil || len(plaintext) < 1 || len(plaintext) > maximumPlaintextBytes {
		return "", errEncodeFailed
	}
	defer clear(plaintext)
	nonce, err := codec.nextNonce()
	if err != nil {
		return "", errEncodeFailed
	}
	aad := []byte(cursorVersion + "." + codec.state.current.kid)
	ciphertext := codec.state.current.aead.Seal(nil, nonce[:], plaintext, aad)
	token := strings.Join([]string{
		cursorVersion, codec.state.current.kid,
		base64.RawURLEncoding.EncodeToString(nonce[:]),
		base64.RawURLEncoding.EncodeToString(ciphertext),
	}, ".")
	clear(ciphertext)
	if len(token) < 1 || len(token) > maximumCursorBytes {
		return "", errEncodeFailed
	}
	return token, nil
}
