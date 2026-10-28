package main

import (
	"crypto/sha256"
	"time"

	passwordblocklistv1 "github.com/Onurlulardan/harpy/contracts/passwordblocklist/v1"
)

func addPasswordBlocklists(set *artifactSet, _ time.Time) error {
	common, err := newPasswordBlocklist(passwordblocklistv1.KindCommon)
	if err != nil {
		return err
	}
	defer clear(common)
	compromised, err := newPasswordBlocklist(passwordblocklistv1.KindCompromised)
	if err != nil {
		return err
	}
	defer clear(compromised)
	if err := set.add("iam-password-common.bin", common); err != nil {
		return err
	}
	if err := set.add("iam-password-compromised.bin", compromised); err != nil {
		return err
	}
	return addJSON(set, "iam-password-blocklists.json", map[string]any{
		"version": 1, "normalization": "unicode_nfc_v1", "digest": "sha256",
		"common": map[string]string{
			"path":   "/run/secrets/iam-password-common.bin",
			"sha256": sha256Hex(common),
		},
		"compromised": map[string]string{
			"path":   "/run/secrets/iam-password-compromised.bin",
			"sha256": sha256Hex(compromised),
		},
	})
}

func newPasswordBlocklist(kind passwordblocklistv1.Kind) ([]byte, error) {
	candidate, err := randomBytes(32)
	if err != nil {
		return nil, err
	}
	defer clear(candidate)
	record := sha256.Sum256(candidate)
	content, err := passwordblocklistv1.Encode(kind, [][32]byte{record})
	clear(record[:])
	return content, err
}
