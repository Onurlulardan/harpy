package config

import (
	"encoding/hex"
	"encoding/json"
	"errors"
)

var errInvalidIAMPasswordBlocklists = errors.New("IAM password blocklists are invalid")

type IAMPasswordBlocklistArtifact struct {
	iamKeyRedaction
	path   AbsolutePath
	digest [32]byte
	valid  bool
}

func (artifact IAMPasswordBlocklistArtifact) Validate() error {
	if !artifact.valid || artifact.path.Value() == "" || iamAllZero(artifact.digest[:]) {
		return errInvalidIAMPasswordBlocklists
	}
	return nil
}

func (artifact IAMPasswordBlocklistArtifact) Path() (AbsolutePath, error) {
	if err := artifact.Validate(); err != nil {
		return AbsolutePath{}, err
	}
	return artifact.path, nil
}

func (artifact IAMPasswordBlocklistArtifact) SHA256() ([32]byte, error) {
	if err := artifact.Validate(); err != nil {
		return [32]byte{}, err
	}
	return artifact.digest, nil
}

func decodeIAMPasswordBlocklistArtifact(
	raw map[string]json.RawMessage,
) (IAMPasswordBlocklistArtifact, bool) {
	var pathText, digestText string
	if json.Unmarshal(raw["path"], &pathText) != nil ||
		json.Unmarshal(raw["sha256"], &digestText) != nil {
		return IAMPasswordBlocklistArtifact{}, false
	}
	path, valid := newAbsolutePath(pathText)
	decoded, err := hex.DecodeString(digestText)
	if !valid || err != nil || len(digestText) != 64 || len(decoded) != 32 ||
		hex.EncodeToString(decoded) != digestText || iamAllZero(decoded) {
		clear(decoded)
		return IAMPasswordBlocklistArtifact{}, false
	}
	var digest [32]byte
	copy(digest[:], decoded)
	clear(decoded)
	return IAMPasswordBlocklistArtifact{path: path, digest: digest, valid: true}, true
}
