package accesstoken

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

func TestVerifierRejectsAlgorithmAndHeaderConfusion(t *testing.T) {
	fixture := newTokenFixture(t)
	claims := fixture.claims(t)
	payload := mustPayload(t, claims)
	tests := map[string]string{
		"deprecated EdDSA": fmt.Sprintf(`{"typ":%q,"alg":"EdDSA","kid":%q}`, Type, fixture.keyID),
		"none":             fmt.Sprintf(`{"typ":%q,"alg":"none","kid":%q}`, Type, fixture.keyID),
		"wrong type":       fmt.Sprintf(`{"typ":"at+jwt","alg":%q,"kid":%q}`, Algorithm, fixture.keyID),
		"unknown member":   fmt.Sprintf(`{"typ":%q,"alg":%q,"kid":%q,"jku":"https://evil.test"}`, Type, Algorithm, fixture.keyID),
		"duplicate alg":    fmt.Sprintf(`{"typ":%q,"alg":%q,"alg":%q,"kid":%q}`, Type, Algorithm, Algorithm, fixture.keyID),
	}
	for name, header := range tests {
		t.Run(name, func(t *testing.T) {
			token := signRaw(fixture.private, []byte(header), payload)
			if _, err := fixture.verifier(t).Verify(context.Background(), token, fixture.now); err != ErrInvalidToken {
				t.Fatalf("expected invalid token, got %v", err)
			}
		})
	}
}

func TestVerifierRejectsAmbiguousClaims(t *testing.T) {
	fixture := newTokenFixture(t)
	header := []byte(fmt.Sprintf(`{"typ":%q,"alg":%q,"kid":%q}`, Type, Algorithm, fixture.keyID))
	base := string(mustPayload(t, fixture.claims(t)))
	tests := map[string]string{
		"unknown":           strings.TrimSuffix(base, "}") + `,"role_ids":[]}`,
		"duplicate issuer":  strings.TrimSuffix(base, "}") + `,"iss":"http://localhost:8000/auth"}`,
		"array audience":    strings.Replace(base, `"aud":"http://localhost:8000/api"`, `"aud":["http://localhost:8000/api"]`, 1),
		"float iat":         strings.Replace(base, `"iat":2000000000`, `"iat":2e9`, 1),
		"noncanonical uuid": strings.Replace(base, "10000000-0000-4000-8000-000000000001", "{10000000-0000-4000-8000-000000000001}", 1),
	}
	for name, payload := range tests {
		t.Run(name, func(t *testing.T) {
			token := signRaw(fixture.private, header, []byte(payload))
			if _, err := fixture.verifier(t).Verify(context.Background(), token, fixture.now); err != ErrInvalidToken {
				t.Fatalf("expected invalid token, got %v", err)
			}
		})
	}
}

func TestVerifierEnforcesWireAndCanonicalSegments(t *testing.T) {
	fixture := newTokenFixture(t)
	verifier := fixture.verifier(t)
	invalid := []string{
		"", "a.b", "a.b.c.d", "=.e30.eA", strings.Repeat("a", MaximumWireSize+1),
	}
	for _, token := range invalid {
		if _, err := verifier.Verify(context.Background(), token, fixture.now); err != ErrInvalidToken {
			t.Fatalf("invalid compact token accepted: %q", token)
		}
	}
}

func mustPayload(t *testing.T, claims Claims) []byte {
	t.Helper()
	document, err := json.Marshal(claims.document())
	if err != nil {
		t.Fatal(err)
	}
	return document
}
