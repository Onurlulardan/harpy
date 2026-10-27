package csrf

import (
	"context"
	"crypto/hmac"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type Verifier struct {
	policy   OriginPolicy
	resolver KeyResolver
}

func NewVerifier(policy OriginPolicy, resolver KeyResolver) (*Verifier, error) {
	if len(policy.allowed) == 0 || nilKeyResolver(resolver) {
		return nil, ErrInvalidPolicy
	}
	return &Verifier{policy: policy, resolver: resolver}, nil
}

func (verifier *Verifier) Verify(
	ctx context.Context,
	cookie, header []byte,
	selector [16]byte,
	origin string,
	now time.Time,
) error {
	if verifier == nil || nilKeyResolver(verifier.resolver) || ctx == nil ||
		ctx.Err() != nil || selector == [16]byte{} ||
		!verifier.policy.Allows(origin) || now.IsZero() || len(cookie) == 0 ||
		len(cookie) > MaximumWireSize || len(cookie) != len(header) || !hmac.Equal(cookie, header) {
		return ErrInvalidCredential
	}
	parts := strings.Split(string(cookie), ".")
	if len(parts) != 5 || !validKeyID(parts[0]) {
		return ErrInvalidCredential
	}
	nonce, ok := decode(parts[1], nonceBytes)
	if !ok || allZero(nonce) {
		wipe(nonce)
		return ErrInvalidCredential
	}
	wipe(nonce)
	issuedAt, ok := parseTimestamp(parts[2])
	if !ok {
		return ErrInvalidCredential
	}
	expiry, ok := parseTimestamp(parts[3])
	skew := int64(ClockSkew.Seconds())
	if !ok || expiry != issuedAt+int64(Lifetime.Seconds()) ||
		issuedAt > now.Unix()+skew || now.Unix() >= expiry+skew {
		return ErrInvalidCredential
	}
	presentedMAC, ok := decode(parts[4], keyBytes)
	if !ok {
		return ErrInvalidCredential
	}
	resolved, err := verifier.resolver.CSRFKey(ctx, parts[0])
	if err != nil || len(resolved) != keyBytes {
		wipe(presentedMAC)
		return ErrInvalidCredential
	}
	key := append([]byte(nil), resolved...)
	expected := authenticationCode(
		key, parts[0], parts[1], encode(selector[:]), origin, parts[2], parts[3],
	)
	wipe(key)
	matched := hmac.Equal(presentedMAC, expected[:])
	wipe(presentedMAC)
	wipe(expected[:])
	if !matched {
		return ErrInvalidCredential
	}
	return nil
}

func nilKeyResolver(resolver KeyResolver) bool {
	if resolver == nil {
		return true
	}
	value := reflect.ValueOf(resolver)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface,
		reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func parseTimestamp(value string) (int64, bool) {
	if value == "" || (len(value) > 1 && value[0] == '0') {
		return 0, false
	}
	for _, character := range value {
		if character < '0' || character > '9' {
			return 0, false
		}
	}
	expiry, err := strconv.ParseInt(value, 10, 64)
	return expiry, err == nil && expiry > 0
}
