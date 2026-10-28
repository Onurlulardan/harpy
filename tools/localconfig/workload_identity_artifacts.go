package main

import (
	"crypto/ed25519"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"
)

const localWorkloadAudience = "//iam.googleapis.com/projects/123456789012/" +
	"locations/global/workloadIdentityPools/harpy-local/providers/local-runtime"
const hostWorkloadAudience = "//iam.googleapis.com/projects/123456789012/" +
	"locations/global/workloadIdentityPools/harpy-host/providers/host-runtime"

func addWorkloadIdentityArtifacts(set *artifactSet, now time.Time) error {
	masterKey, err := randomBytes(32)
	if err != nil {
		return err
	}
	defer clear(masterKey)
	if err := set.add("kms-master-key.bin", masterKey); err != nil {
		return err
	}
	if set.options.profile == profileForgeGuard {
		return addHostWorkloadIdentityArtifacts(set)
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

func addHostWorkloadIdentityArtifacts(set *artifactSet) error {
	accessKey, err := randomBytes(32)
	if err != nil {
		return err
	}
	defer clear(accessKey)
	if err := set.add("kms-access-token-signing-key.bin", accessKey); err != nil {
		return err
	}
	keys := make([]map[string]string, 0, 2)
	for _, workload := range []struct {
		name, subject string
	}{
		{name: "api", subject: "harpy-host-api"},
		{name: "worker", subject: "harpy-host-worker"},
	} {
		seed, seedErr := randomBytes(ed25519.SeedSize)
		if seedErr != nil {
			return seedErr
		}
		privateKey := ed25519.NewKeyFromSeed(seed)
		publicKey := privateKey.Public().(ed25519.PublicKey)
		keys = append(keys, map[string]string{
			"kid": workload.name, "subject": workload.subject,
			"public_key": base64.RawURLEncoding.EncodeToString(publicKey),
		})
		if err := set.add(workload.name+"-workload-signing-seed", seed); err != nil {
			clear(seed)
			clear(privateKey)
			return err
		}
		clear(seed)
		clear(privateKey)
		if err := addExecutableWIF(set, workload.name, workload.subject); err != nil {
			return err
		}
	}
	return addJSON(set, "host-workload-public-keys.json", map[string]any{
		"version": 1,
		"keys":    keys,
	})
}

func addExecutableWIF(set *artifactSet, name, subject string) error {
	command := "/app/workload-token -key /run/secrets/" + name +
		"-workload-signing-seed -kid " + name + " -subject " + subject
	return addJSON(set, name+"-gcp-wif.json", map[string]any{
		"type":               "external_account",
		"audience":           hostWorkloadAudience,
		"subject_token_type": "urn:ietf:params:oauth:token-type:jwt",
		"token_url":          "https://sts.googleapis.com/v1/token",
		"credential_source": map[string]any{
			"executable": map[string]any{
				"command":        command,
				"timeout_millis": 5000,
			},
		},
	})
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
