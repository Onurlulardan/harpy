package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"time"
)

type runtimeHMACKey struct {
	file    string
	env     string
	purpose string
	marker  byte
}

func writeIdentityMaterial(dir string, now time.Time) ([]string, error) {
	lines := make([]string, 0, 24)
	cursor, err := writeCursorMaterial(dir)
	if err != nil {
		return nil, err
	}
	catalog, err := writeCatalogMaterial(dir)
	if err != nil {
		return nil, err
	}
	keyrings, err := writeIAMKeyrings(dir, now)
	if err != nil {
		return nil, err
	}
	passwords, err := writePasswordBlocklists(dir)
	if err != nil {
		return nil, err
	}
	origins, err := json.Marshal([]string{"http://127.0.0.1:8080"})
	if err != nil {
		return nil, err
	}
	lines = append(lines, cursor, catalog)
	lines = append(lines,
		runtimeLine("HARPY_IAM_ACCESS_TOKEN_ISSUER", "http://127.0.0.1:8080"),
		runtimeLine("HARPY_IAM_ACCESS_TOKEN_AUDIENCE", "http://127.0.0.1:8080/api"),
		runtimeLine("HARPY_IAM_ALLOWED_ORIGINS", string(origins)),
	)
	lines = append(lines, keyrings...)
	lines = append(lines, passwords)
	bridge, err := writeActorBridge(dir)
	if err != nil {
		return nil, err
	}
	invitation, err := writeInvitationMaterial(dir)
	if err != nil {
		return nil, err
	}
	lines = append(lines, bridge...)
	return append(lines, invitation...), nil
}

func writeCursorMaterial(dir string) (string, error) {
	current, err := markedMaterial(16)
	if err != nil {
		return "", err
	}
	previous, err := markedMaterial(17)
	if err != nil {
		return "", err
	}
	body := fmt.Sprintf(
		`{"version":1,"current":{"kid":"current-1","key":%q},"previous":{"kid":"previous-1","key":%q}}`,
		rawKey(current[:]), rawKey(previous[:]),
	)
	path, err := writeRuntimeFile(dir, "cursor-keyring.json", []byte(body))
	if err != nil {
		return "", err
	}
	return runtimeLine("HARPY_REST_QUERY_CURSOR_KEYRING_FILE", path), nil
}

func writeCatalogMaterial(dir string) (string, error) {
	public, _, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return "", err
	}
	body := fmt.Sprintf(
		`{"version":1,"keys":[{"key_id":"release-local","public_key":%q}]}`,
		rawKey(public),
	)
	path, err := writeRuntimeFile(dir, "catalog-package-keyring.json", []byte(body))
	if err != nil {
		return "", err
	}
	return runtimeLine("HARPY_CATALOG_PACKAGE_PUBLIC_KEYRING_FILE", path), nil
}

func writeIAMKeyrings(dir string, now time.Time) ([]string, error) {
	activated := now.Add(-24 * time.Hour)
	access, err := markedMaterial(1)
	if err != nil {
		return nil, err
	}
	accessBody := fmt.Sprintf(
		`{"version":1,"active":{"kid":%q,"seed":%q,"activated_at":%q,"sign_until":%q},"retired_public":[]}`,
		iamKeyID(access), rawKey(access[:]), iamStamp(activated),
		iamStamp(activated.Add(30*24*time.Hour)),
	)
	accessPath, err := writeRuntimeFile(dir, "access-token-keyring.json", []byte(accessBody))
	if err != nil {
		return nil, err
	}
	lines := []string{runtimeLine("HARPY_IAM_ACCESS_TOKEN_KEYRING_FILE", accessPath)}
	for _, item := range runtimeHMACKeys() {
		material, materialErr := markedMaterial(item.marker)
		if materialErr != nil {
			return nil, materialErr
		}
		body := fmt.Sprintf(
			`{"version":1,"purpose":%q,"active":{"kid":%q,"key":%q,"activated_at":%q,"issue_until":%q},"retained":[]}`,
			item.purpose, iamKeyID(material), rawKey(material[:]), iamStamp(activated),
			iamStamp(activated.Add(90*24*time.Hour)),
		)
		path, writeErr := writeRuntimeFile(dir, item.file, []byte(body))
		if writeErr != nil {
			return nil, writeErr
		}
		lines = append(lines, runtimeLine(item.env+"_FILE", path))
	}
	return lines, nil
}

func runtimeHMACKeys() []runtimeHMACKey {
	return []runtimeHMACKey{
		{"refresh-token-keyring.json", "HARPY_IAM_REFRESH_TOKEN_KEYRING", "refresh_token_v1", 2},
		{"csrf-keyring.json", "HARPY_IAM_CSRF_KEYRING", "csrf_v1", 3},
		{"login-directory-keyring.json", "HARPY_IAM_LOGIN_DIRECTORY_KEYRING", "login_directory_v1", 4},
		{"auth-throttle-keyring.json", "HARPY_IAM_AUTH_THROTTLE_KEYRING", "auth_throttle_v1", 5},
		{"session-network-keyring.json", "HARPY_IAM_SESSION_NETWORK_KEYRING", "session_network_v1", 6},
		{"one-time-token-keyring.json", "HARPY_IAM_ONE_TIME_TOKEN_KEYRING", "one_time_token_v1", 7},
		{"recovery-code-keyring.json", "HARPY_IAM_RECOVERY_CODE_KEYRING", "recovery_code_v1", 8},
	}
}
