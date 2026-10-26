package csrf

import (
	"context"
	"testing"
)

func TestCSRFCredentialRoundTrip(t *testing.T) {
	fixture := newCSRFFixture(t)
	wire := fixture.issue(t)
	if len(wire) > MaximumWireSize {
		t.Fatalf("wire length = %d", len(wire))
	}
	if err := fixture.verifier(t).Verify(
		context.Background(), wire, append([]byte(nil), wire...),
		fixture.selector, fixture.origin, fixture.now,
	); err != nil {
		t.Fatal(err)
	}
}

func TestCredentialRevealIsOneTimeAndWipes(t *testing.T) {
	fixture := newCSRFFixture(t)
	signer, err := NewSigner(fixture.policy, fixture.keyID, fixture.key, zeroReader{})
	if err != nil {
		t.Fatal(err)
	}
	defer signer.Close()
	if _, err := signer.Sign(fixture.selector, fixture.origin, fixture.now); err != ErrInvalidCredential {
		t.Fatalf("all-zero randomness result = %v", err)
	}
	credential := &Credential{wire: []byte("secret"), active: true}
	if err := credential.RevealOnce(func([]byte) error { return nil }); err != nil {
		t.Fatal(err)
	}
	if credential.wire != nil || credential.active {
		t.Fatal("revealed credential remained active")
	}
	if err := credential.RevealOnce(func([]byte) error { return nil }); err != ErrInvalidCredential {
		t.Fatalf("second reveal result = %v", err)
	}
}

type zeroReader struct{}

func (zeroReader) Read(value []byte) (int, error) {
	for index := range value {
		value[index] = 0
	}
	return len(value), nil
}
