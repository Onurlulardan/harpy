package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"
)

const localWorkloadAudience = "//iam.googleapis.com/projects/123456789012/" +
	"locations/global/workloadIdentityPools/harpy-local/providers/local-runtime"

func addWorkloadIdentityArtifacts(set *artifactSet, now time.Time) error {
	masterKey, err := randomBytes(32)
	if err != nil {
		return err
	}
	defer clear(masterKey)
	if err := set.add("kms-master-key.bin", masterKey); err != nil {
		return err
	}
	return addLocalWorkloadIdentityArtifacts(set, now)
}

func addLocalWorkloadIdentityArtifacts(set *artifactSet, now time.Time) error {
	tokenKey, err := randomBytes(32)
	if err != nil {
		return err
	}
	defer clear(tokenKey)
	expiresAt := now.Add(30 * 24 * time.Hour)
	token, err := issueSubjectToken(
		now, expiresAt, tokenKey, localWorkloadAudience,
		"harpy-local", "harpy-local-runtime",
	)
	if err != nil {
		return err
	}
	if err := set.add("gcp-subject-token-hmac-key.bin", tokenKey); err != nil {
		return err
	}
	if err := set.add("gcp-subject-token.jwt", []byte(token)); err != nil {
		return err
	}
	if err := addJSON(set, "gcp-wif.json", map[string]any{
		"type":               "external_account",
		"audience":           localWorkloadAudience,
		"subject_token_type": "urn:ietf:params:oauth:token-type:jwt",
		"token_url":          "https://sts.googleapis.com/v1/token",
		"credential_source": map[string]string{
			"file": "/run/secrets/gcp-subject-token.jwt",
		},
	}); err != nil {
		return err
	}
	set.trackExpiry(expiresAt)
	return nil
}

func issueSubjectToken(
	now time.Time,
	expiresAt time.Time,
	key []byte,
	audience string,
	issuer string,
	subject string,
) (string, error) {
	header, err := marshalJSON(map[string]string{"alg": "HS256", "typ": "JWT"})
	if err != nil {
		return "", err
	}
	jti, err := randomBase64(16)
	if err != nil {
		return "", err
	}
	payload, err := marshalJSON(map[string]any{
		"aud": audience,
		"exp": expiresAt.Unix(),
		"iat": now.Unix(),
		"iss": issuer,
		"jti": jti,
		"nbf": now.Add(-time.Minute).Unix(),
		"sub": subject,
	})
	if err != nil {
		return "", err
	}
	encoded := base64.RawURLEncoding.EncodeToString(header) + "." +
		base64.RawURLEncoding.EncodeToString(payload)
	signer := hmac.New(sha256.New, key)
	if _, err := signer.Write([]byte(encoded)); err != nil {
		return "", fmt.Errorf("sign subject token: %w", err)
	}
	return encoded + "." + base64.RawURLEncoding.EncodeToString(signer.Sum(nil)), nil
}
