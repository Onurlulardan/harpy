package config

import (
	"bytes"
	"encoding/json"
	"unicode/utf8"
)

const iamPasswordBlocklistsEnvironmentKey = "HARPY_IAM_PASSWORD_BLOCKLISTS"

type IAMPasswordBlocklists struct {
	iamKeyRedaction
	common      IAMPasswordBlocklistArtifact
	compromised IAMPasswordBlocklistArtifact
	valid       bool
}

func loadIAMPasswordBlocklists(values environmentValues) (IAMPasswordBlocklists, error) {
	secret, err := loadSecret(values, iamPasswordBlocklistsEnvironmentKey, true, secretFileLimit)
	if err != nil {
		return IAMPasswordBlocklists{}, err
	}
	blocklists, ok := parseIAMPasswordBlocklists(secret)
	if !ok {
		return IAMPasswordBlocklists{}, invalidIAMKeyring(
			iamPasswordBlocklistsEnvironmentKey, "iam-password-blocklists.v1",
		)
	}
	return blocklists, nil
}

func parseIAMPasswordBlocklists(secret Secret) (IAMPasswordBlocklists, bool) {
	document := []byte(secret.Value())
	if !utf8.Valid(document) || !json.Valid(document) ||
		rejectDuplicateJSONMembers(document) != nil {
		return IAMPasswordBlocklists{}, false
	}
	root, ok := decodeIAMObject(
		document, "version", "normalization", "digest", "common", "compromised",
	)
	if !ok || !blocklistHeaderValid(root) {
		return IAMPasswordBlocklists{}, false
	}
	common, ok := decodePasswordBlocklistMember(root["common"])
	if !ok {
		return IAMPasswordBlocklists{}, false
	}
	compromised, ok := decodePasswordBlocklistMember(root["compromised"])
	if !ok {
		return IAMPasswordBlocklists{}, false
	}
	result := IAMPasswordBlocklists{
		common: common, compromised: compromised, valid: true,
	}
	return result, result.Validate() == nil
}

func blocklistHeaderValid(root map[string]json.RawMessage) bool {
	var version int
	var normalization, digest string
	return json.Unmarshal(root["version"], &version) == nil && version == 1 &&
		json.Unmarshal(root["normalization"], &normalization) == nil &&
		normalization == "unicode_nfc_v1" &&
		json.Unmarshal(root["digest"], &digest) == nil && digest == "sha256"
}

func decodePasswordBlocklistMember(
	raw json.RawMessage,
) (IAMPasswordBlocklistArtifact, bool) {
	object, ok := decodeIAMObject(raw, "path", "sha256")
	if !ok {
		return IAMPasswordBlocklistArtifact{}, false
	}
	return decodeIAMPasswordBlocklistArtifact(object)
}

func (config IAMPasswordBlocklists) Validate() error {
	if !config.valid || config.common.Validate() != nil || config.compromised.Validate() != nil ||
		config.common.path.Value() == config.compromised.path.Value() ||
		bytes.Equal(config.common.digest[:], config.compromised.digest[:]) {
		return errInvalidIAMPasswordBlocklists
	}
	return nil
}

func (config IAMPasswordBlocklists) Common() (IAMPasswordBlocklistArtifact, error) {
	if err := config.Validate(); err != nil {
		return IAMPasswordBlocklistArtifact{}, err
	}
	return config.common, nil
}

func (config IAMPasswordBlocklists) Compromised() (IAMPasswordBlocklistArtifact, error) {
	if err := config.Validate(); err != nil {
		return IAMPasswordBlocklistArtifact{}, err
	}
	return config.compromised, nil
}
