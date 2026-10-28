package main

import "time"

type iamHMACTarget struct {
	file    string
	purpose string
}

var iamHMACTargets = []iamHMACTarget{
	{"iam-refresh-token-keyring.json", "refresh_token_v1"},
	{"iam-csrf-keyring.json", "csrf_v1"},
	{"iam-login-directory-keyring.json", "login_directory_v1"},
	{"iam-auth-throttle-keyring.json", "auth_throttle_v1"},
	{"iam-session-network-keyring.json", "session_network_v1"},
	{"iam-one-time-token-keyring.json", "one_time_token_v1"},
	{"iam-recovery-code-keyring.json", "recovery_code_v1"},
}

func addIAMKeyrings(set *artifactSet, now time.Time) error {
	activatedAt := now.Add(-time.Minute)
	signUntil := activatedAt.Add(30 * 24 * time.Hour)
	kid, err := randomBase64(12)
	if err != nil {
		return err
	}
	seed, err := randomBase64(32)
	if err != nil {
		return err
	}
	access := map[string]any{
		"version": 1,
		"active": map[string]any{
			"kid": kid, "seed": seed,
			"activated_at": canonicalTimestamp(activatedAt),
			"sign_until":   canonicalTimestamp(signUntil),
		},
		"retired_public": []any{},
	}
	if err := addJSON(set, "iam-access-token-keyring.json", access); err != nil {
		return err
	}
	set.trackExpiry(signUntil)
	return addIAMHMACKeyrings(set, activatedAt)
}

func addIAMHMACKeyrings(set *artifactSet, activatedAt time.Time) error {
	issueUntil := activatedAt.Add(90 * 24 * time.Hour)
	for _, target := range iamHMACTargets {
		kid, err := randomBase64(12)
		if err != nil {
			return err
		}
		key, err := randomBase64(32)
		if err != nil {
			return err
		}
		document := map[string]any{
			"version": 1, "purpose": target.purpose,
			"active": map[string]any{
				"kid": kid, "key": key,
				"activated_at": canonicalTimestamp(activatedAt),
				"issue_until":  canonicalTimestamp(issueUntil),
			},
			"retained": []any{},
		}
		if err := addJSON(set, target.file, document); err != nil {
			return err
		}
	}
	set.trackExpiry(issueUntil)
	return nil
}
