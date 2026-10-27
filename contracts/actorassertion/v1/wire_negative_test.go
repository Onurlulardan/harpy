package actorassertionv1

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

func TestVerifierRejectsAlgorithmConfusionUnknownAndDuplicateFields(t *testing.T) {
	fixture := newAssertionFixture(t)
	credential, _ := fixture.issue(t)
	wire, _ := credential.Wire()
	parts := strings.Split(wire, ".")
	payload, _ := base64.RawURLEncoding.DecodeString(parts[1])
	var object map[string]any
	if err := json.Unmarshal(payload, &object); err != nil {
		t.Fatal(err)
	}
	algorithm := cloneObject(object)
	algorithm["algorithm"] = "EdDSA"
	none := cloneObject(object)
	none["algorithm"] = "none"
	unknown := cloneObject(object)
	unknown["unexpected"] = true
	duplicate := append(bytes.TrimSuffix(payload, []byte("}")),
		[]byte(`,"iss":"urn:harpy:core-api"}`)...)
	cases := []string{
		resignObject(t, fixture.private, algorithm),
		resignObject(t, fixture.private, none),
		resignObject(t, fixture.private, unknown),
		resignPayload(fixture.private, duplicate),
	}
	verifier, _ := NewVerifier(map[string]ed25519.PublicKey{"edge-2026-07": fixture.public})
	for index, candidate := range cases {
		if _, err := verifier.Verify(candidate, fixture.expectation, fixture.now); err == nil {
			t.Fatalf("ambiguous/confused wire %d was accepted", index)
		}
	}
}

func TestVerifierRejectsWrongUnknownAndMalformedSignatures(t *testing.T) {
	fixture := newAssertionFixture(t)
	credential, _ := fixture.issue(t)
	wire, _ := credential.Wire()
	wrongPrivate := ed25519.NewKeyFromSeed(bytes.Repeat([]byte{8}, ed25519.SeedSize))
	wrongPublic := append(ed25519.PublicKey(nil), wrongPrivate[ed25519.SeedSize:]...)
	wrongVerifier, _ := NewVerifier(map[string]ed25519.PublicKey{"edge-2026-07": wrongPublic})
	if _, err := wrongVerifier.Verify(wire, fixture.expectation, fixture.now); err != ErrSignature {
		t.Fatalf("wrong signature error = %v", err)
	}
	unknownVerifier, _ := NewVerifier(map[string]ed25519.PublicKey{"other-key": fixture.public})
	if _, err := unknownVerifier.Verify(wire, fixture.expectation, fixture.now); err != ErrSignature {
		t.Fatalf("unknown key error = %v", err)
	}
	parts := strings.Split(wire, ".")
	parts[2] = base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{1}, ed25519.SignatureSize))
	if _, err := wrongVerifier.Verify(strings.Join(parts, "."), fixture.expectation, fixture.now); err == nil {
		t.Fatal("malformed signature was accepted")
	}
}

func cloneObject(source map[string]any) map[string]any {
	result := make(map[string]any, len(source))
	for key, value := range source {
		result[key] = value
	}
	return result
}

func resignObject(t *testing.T, key ed25519.PrivateKey, object map[string]any) string {
	t.Helper()
	document, err := json.Marshal(object)
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, MaximumWireSize)
	if err != nil {
		t.Fatal(err)
	}
	return resignPayload(key, canonical)
}

func resignPayload(key ed25519.PrivateKey, payload []byte) string {
	segment := base64.RawURLEncoding.EncodeToString(payload)
	input := wirePrefix + "." + segment
	signature := ed25519.Sign(key, []byte(input))
	return input + "." + base64.RawURLEncoding.EncodeToString(signature)
}
