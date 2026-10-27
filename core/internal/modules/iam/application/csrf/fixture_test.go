package csrf

import (
	"bytes"
	"context"
	"encoding/base64"
	"testing"
	"time"
)

type csrfFixture struct {
	policy   OriginPolicy
	keyID    string
	key      []byte
	selector [16]byte
	origin   string
	now      time.Time
}

func newCSRFFixture(t *testing.T) csrfFixture {
	t.Helper()
	policy, err := NewOriginPolicy([]string{"http://localhost:5173"}, true)
	if err != nil {
		t.Fatal(err)
	}
	fixture := csrfFixture{
		policy: policy,
		keyID:  base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{3}, 12)),
		key:    bytes.Repeat([]byte{9}, keyBytes), origin: "http://localhost:5173",
		now: time.Unix(2_000_000_000, 0).UTC(),
	}
	copy(fixture.selector[:], []byte("0123456789abcdef"))
	return fixture
}

func (fixture csrfFixture) issue(t *testing.T) []byte {
	t.Helper()
	signer, err := NewSigner(
		fixture.policy, fixture.keyID, fixture.key,
		bytes.NewReader(bytes.Repeat([]byte{5}, nonceBytes)),
	)
	if err != nil {
		t.Fatal(err)
	}
	defer signer.Close()
	credential, err := signer.Sign(fixture.selector, fixture.origin, fixture.now)
	if err != nil {
		t.Fatal(err)
	}
	var wire []byte
	if err := credential.RevealOnce(func(value []byte) error {
		wire = append([]byte(nil), value...)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	return wire
}

func (fixture csrfFixture) verifier(t *testing.T) *Verifier {
	t.Helper()
	verifier, err := NewVerifier(fixture.policy, KeyResolverFunc(
		func(_ context.Context, keyID string) ([]byte, error) {
			if keyID != fixture.keyID {
				return nil, ErrInvalidKey
			}
			return append([]byte(nil), fixture.key...), nil
		},
	))
	if err != nil {
		t.Fatal(err)
	}
	return verifier
}
