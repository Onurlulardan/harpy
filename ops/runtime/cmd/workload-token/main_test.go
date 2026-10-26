package main

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRunIssuesShortLivedExecutableJWT(t *testing.T) {
	seed := bytes.Repeat([]byte{0x42}, ed25519.SeedSize)
	keyPath := filepath.Join(t.TempDir(), "api.seed")
	if err := os.WriteFile(keyPath, seed, 0o600); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 7, 23, 6, 0, 0, 0, time.UTC)
	var output bytes.Buffer
	err := run(
		[]string{"-key", keyPath, "-kid", "api", "-subject", "harpy-host-api"},
		validEnvironment, func() time.Time { return now },
		bytes.NewReader(bytes.Repeat([]byte{0x24}, 16)), &output,
	)
	if err != nil {
		t.Fatal(err)
	}
	var response executableResponse
	if err := json.Unmarshal(output.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if !response.Success || response.Version != 1 ||
		response.TokenType != jwtTokenType ||
		response.ExpirationTime != now.Add(5*time.Minute).Unix() {
		t.Fatalf("unexpected response: %+v", response)
	}
	parts := strings.Split(response.IDToken, ".")
	if len(parts) != 3 {
		t.Fatal("subject token is not a compact JWT")
	}
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || !ed25519.Verify(
		ed25519.NewKeyFromSeed(seed).Public().(ed25519.PublicKey),
		[]byte(parts[0]+"."+parts[1]), signature,
	) {
		t.Fatalf("subject token signature is invalid: %v", err)
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatal(err)
	}
	var claims struct {
		Audience string `json:"aud"`
		Expires  int64  `json:"exp"`
		IssuedAt int64  `json:"iat"`
		Subject  string `json:"sub"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		t.Fatal(err)
	}
	if claims.Audience != hostAudience || claims.IssuedAt != now.Unix() ||
		claims.Expires != now.Add(5*time.Minute).Unix() ||
		claims.Subject != "harpy-host-api" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestRunRejectsPurposeAndEnvironmentDrift(t *testing.T) {
	keyPath := filepath.Join(t.TempDir(), "seed")
	if err := os.WriteFile(
		keyPath, bytes.Repeat([]byte{1}, ed25519.SeedSize), 0o600,
	); err != nil {
		t.Fatal(err)
	}
	arguments := []string{
		"-key", keyPath, "-kid", "api", "-subject", "harpy-host-worker",
	}
	if err := run(
		arguments, validEnvironment, time.Now,
		bytes.NewReader(make([]byte, 16)), &bytes.Buffer{},
	); err == nil {
		t.Fatal("cross-purpose workload subject was accepted")
	}
	arguments[5] = "harpy-host-api"
	if err := run(
		arguments, func(string) string { return "" }, time.Now,
		bytes.NewReader(make([]byte, 16)), &bytes.Buffer{},
	); err == nil {
		t.Fatal("missing external-account environment was accepted")
	}
}

func validEnvironment(key string) string {
	switch key {
	case audienceEnvironment:
		return hostAudience
	case tokenTypeEnvironment:
		return jwtTokenType
	case "GOOGLE_EXTERNAL_ACCOUNT_INTERACTIVE":
		return "0"
	default:
		return ""
	}
}
