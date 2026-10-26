package refreshtoken

import (
	"bytes"
	"errors"
	"testing"
)

func TestIssuedTokenRoundTripAndOneTimeReveal(t *testing.T) {
	random := make([]byte, SelectorBytes+SecretBytes)
	for index := range random {
		random[index] = byte(index + 1)
	}
	issued, err := (Generator{randomness: bytes.NewReader(random)}).Generate()
	if err != nil {
		t.Fatal(err)
	}
	selector, err := issued.Selector()
	if err != nil || selector[0] != 1 || selector[15] != 16 {
		t.Fatalf("selector = %x, %v", selector, err)
	}
	key := bytes.Repeat([]byte{9}, 32)
	expected, err := issued.SecretHMAC(key)
	if err != nil {
		t.Fatal(err)
	}
	var wire []byte
	err = issued.RevealOnce(func(value []byte) error {
		wire = append([]byte(nil), value...)
		return nil
	})
	if err != nil || len(wire) != wireBytes || wire[22] != '.' {
		t.Fatalf("wire = %q, %v", wire, err)
	}
	if _, err := issued.Selector(); !errors.Is(err, ErrTokenDestroyed) {
		t.Fatalf("revealed token remained active: %v", err)
	}
	presented, err := Parse(wire)
	if err != nil {
		t.Fatal(err)
	}
	defer presented.Destroy()
	if parsed, _ := presented.Selector(); parsed != selector {
		t.Fatalf("parsed selector = %x", parsed)
	}
	matched, err := presented.VerifyHMAC(key, expected)
	if err != nil || !matched {
		t.Fatalf("valid HMAC: matched=%v err=%v", matched, err)
	}
	wrong := expected
	wrong[0] ^= 1
	matched, err = presented.VerifyHMAC(key, wrong)
	if err != nil || matched {
		t.Fatalf("wrong HMAC: matched=%v err=%v", matched, err)
	}
}

func TestRefreshTokenParserIsCanonicalAndFailClosed(t *testing.T) {
	valid := []byte("AQIDBAUGBwgJCgsMDQ4PEA.ERITFBUWFxgZGhscHR4fICEiIyQlJicoKSorLC0uLzA")
	if _, err := Parse(valid); err != nil {
		t.Fatalf("canonical token rejected: %v", err)
	}
	tests := [][]byte{
		nil,
		[]byte("short.secret"),
		append(append([]byte(nil), valid...), '='),
		append([]byte("AQIDBAUGBwgJCgsMDQ4PEA="), valid[22:]...),
		append([]byte("AQIDBAUGBwgJCgsMDQ4PEA-"), valid[23:]...),
		[]byte("AAAAAAAAAAAAAAAAAAAAAA.AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"),
	}
	for index, wire := range tests {
		if token, err := Parse(wire); !errors.Is(err, ErrTokenInvalid) || token != nil {
			t.Fatalf("invalid token %d result = %v, %v", index, token, err)
		}
	}
}

func TestRefreshTokenGenerationAndLifetimeFailures(t *testing.T) {
	if _, err := (Generator{}).Generate(); !errors.Is(err, ErrRandomUnavailable) {
		t.Fatalf("nil randomness error = %v", err)
	}
	if _, err := (Generator{randomness: bytes.NewReader([]byte{1})}).Generate(); !errors.Is(err, ErrRandomUnavailable) {
		t.Fatalf("short randomness error = %v", err)
	}
	issued, err := (Generator{randomness: bytes.NewReader(bytes.Repeat([]byte{1}, 48))}).Generate()
	if err != nil {
		t.Fatal(err)
	}
	consumeErr := errors.New("cookie write failed")
	if err := issued.RevealOnce(func([]byte) error { return consumeErr }); !errors.Is(err, consumeErr) {
		t.Fatalf("consumer error = %v", err)
	}
	if _, err := issued.SecretHMAC(bytes.Repeat([]byte{1}, 32)); !errors.Is(err, ErrTokenDestroyed) {
		t.Fatalf("consumer failure retained token: %v", err)
	}
	if _, err := secretHMAC([]byte("short"), []byte("secret")); !errors.Is(err, ErrKeyInvalid) {
		t.Fatalf("short key error = %v", err)
	}
}
