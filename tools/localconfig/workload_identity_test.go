package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestGeneratedWorkloadIdentityIsCanonicalAndTokenIsVerifiable(t *testing.T) {
	set, now := generatedSetForTest(t)
	if len(set.files["kms-master-key.bin"]) != 32 ||
		len(set.files["gcp-subject-token-hmac-key.bin"]) != 32 {
		t.Fatal("local KMS or subject-token key is not 32 bytes")
	}
	var credential struct {
		Type             string `json:"type"`
		Audience         string `json:"audience"`
		SubjectTokenType string `json:"subject_token_type"`
		TokenURL         string `json:"token_url"`
		CredentialSource struct {
			File string `json:"file"`
		} `json:"credential_source"`
	}
	if err := json.Unmarshal(set.files["gcp-wif.json"], &credential); err != nil {
		t.Fatal(err)
	}
	if credential.Type != "external_account" ||
		credential.Audience != localWorkloadAudience ||
		credential.SubjectTokenType != "urn:ietf:params:oauth:token-type:jwt" ||
		credential.TokenURL != "https://sts.googleapis.com/v1/token" ||
		credential.CredentialSource.File != "/run/secrets/gcp-subject-token.jwt" {
		t.Fatal("WIF credential document differs from the runtime contract")
	}
	parts := strings.Split(string(set.files["gcp-subject-token.jwt"]), ".")
	if len(parts) != 3 {
		t.Fatal("subject token is not a compact JWT")
	}
	signer := hmac.New(sha256.New, set.files["gcp-subject-token-hmac-key.bin"])
	_, _ = signer.Write([]byte(parts[0] + "." + parts[1]))
	signature, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || !hmac.Equal(signature, signer.Sum(nil)) {
		t.Fatalf("subject-token signature is invalid: %v", err)
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatal(err)
	}
	var claims struct {
		Audience string `json:"aud"`
		Expires  int64  `json:"exp"`
		IssuedAt int64  `json:"iat"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		t.Fatal(err)
	}
	if claims.Audience != localWorkloadAudience || claims.IssuedAt != now.Unix() ||
		time.Unix(claims.Expires, 0).Sub(now) != 30*24*time.Hour {
		t.Fatal("subject-token claims differ from the local WIF contract")
	}
}
