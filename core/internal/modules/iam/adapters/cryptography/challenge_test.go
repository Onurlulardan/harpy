package cryptography

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"
)

func TestChallengeRoundTripIsSingleRevealAndRedacted(t *testing.T) {
	issued, err := NewChallengeGenerator().Generate()
	if err != nil {
		t.Fatal(err)
	}
	defer issued.Destroy()
	var wire []byte
	if err := issued.RevealOnce(func(value []byte) error {
		wire = append([]byte(nil), value...)
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if len(wire) != 43 || issued.RevealOnce(func([]byte) error { return nil }) == nil {
		t.Fatal("challenge reveal contract drifted")
	}
	presented, err := NewPresentedChallenge(wire)
	if err != nil {
		t.Fatal(err)
	}
	defer presented.Destroy()
	key := bytes.Repeat([]byte{0x4a}, 32)
	want, err := issued.HMAC(key)
	if err != nil {
		t.Fatal(err)
	}
	got, err := presented.HMAC(key)
	if err != nil || got != want {
		t.Fatalf("challenge HMAC mismatch: %v", err)
	}
	assertMaterialRedacted(t, presented, wire)
	clear(wire)
}

func TestPresentedChallengeRejectsMalformedWire(t *testing.T) {
	for _, wire := range [][]byte{
		nil, bytes.Repeat([]byte{'a'}, 42), bytes.Repeat([]byte{'!'}, 43),
	} {
		if _, err := NewPresentedChallenge(wire); err == nil {
			t.Fatal("malformed challenge was accepted")
		}
	}
}

func assertMaterialRedacted(t *testing.T, material any, secret []byte) {
	t.Helper()
	encoded, err := json.Marshal(material)
	if err != nil {
		t.Fatal(err)
	}
	for _, rendered := range [][]byte{
		[]byte(fmt.Sprint(material)), []byte(fmt.Sprintf("%#v", material)), encoded,
	} {
		if bytes.Contains(rendered, secret) ||
			!bytes.Contains(rendered, []byte("REDACTED IAM AUTHENTICATION")) {
			t.Fatalf("authentication material was not redacted: %q", rendered)
		}
	}
	if len(secret) == 0 {
		t.Fatal(authentication.ErrAuthenticationContract)
	}
}
