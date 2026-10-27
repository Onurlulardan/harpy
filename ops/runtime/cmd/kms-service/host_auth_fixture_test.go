package main

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"
)

func fixtureWorkloadRegistry() (
	map[string]workloadIdentity,
	ed25519.PrivateKey,
) {
	privateKey := ed25519.NewKeyFromSeed(
		bytes.Repeat([]byte{0x62}, ed25519.SeedSize),
	)
	return map[string]workloadIdentity{
		"api": {
			subject:   "harpy-host-api",
			publicKey: privateKey.Public().(ed25519.PublicKey),
		},
		"worker": {
			subject: "harpy-host-worker",
			publicKey: ed25519.NewKeyFromSeed(
				bytes.Repeat([]byte{0x73}, ed25519.SeedSize),
			).Public().(ed25519.PublicKey),
		},
	}, privateKey
}

func signFixtureSubject(
	t *testing.T,
	privateKey ed25519.PrivateKey,
	now time.Time,
) string {
	t.Helper()
	header, err := json.Marshal(map[string]string{
		"alg": "EdDSA", "kid": "api", "typ": "JWT",
	})
	if err != nil {
		t.Fatal(err)
	}
	payload, err := json.Marshal(map[string]any{
		"aud": hostAudience, "exp": now.Add(5 * time.Minute).Unix(),
		"iat": now.Unix(), "iss": "harpy-host", "jti": "subject-id",
		"nbf": now.Add(-30 * time.Second).Unix(), "sub": "harpy-host-api",
	})
	if err != nil {
		t.Fatal(err)
	}
	wire := base64.RawURLEncoding.EncodeToString(header) + "." +
		base64.RawURLEncoding.EncodeToString(payload)
	signature := ed25519.Sign(privateKey, []byte(wire))
	return wire + "." + base64.RawURLEncoding.EncodeToString(signature)
}
