package iam

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identifier"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

type InstallationDirectoryEvidence struct {
	KeyID        string
	LookupHMAC   [32]byte
	SourceSHA256 [32]byte
}

func PrepareInstallationDirectory(
	email string,
	keyring config.IAMDirectoryKeyring,
) (InstallationDirectoryEvidence, error) {
	normalized, err := identifier.NormalizeEmailV1(email)
	if err != nil {
		return InstallationDirectoryEvidence{}, ErrInstallationIdentityInvalid
	}
	keys, err := keyring.Keys()
	if err != nil || len(keys) == 0 {
		return InstallationDirectoryEvidence{}, ErrInstallationIdentityInvalid
	}
	keyID, idErr := keys[0].ID()
	material, materialErr := keys[0].Material()
	active, activeErr := keys[0].Active()
	for index := range keys {
		keys[index].Destroy()
	}
	if idErr != nil || materialErr != nil || activeErr != nil || !active {
		clear(material[:])
		return InstallationDirectoryEvidence{}, ErrInstallationIdentityInvalid
	}
	mac := hmac.New(sha256.New, material[:])
	clear(material[:])
	if _, err := mac.Write([]byte(normalized.Normalized())); err != nil {
		return InstallationDirectoryEvidence{}, fmt.Errorf(
			"prepare installation directory: %w", err,
		)
	}
	var lookup [32]byte
	copy(lookup[:], mac.Sum(nil))
	return InstallationDirectoryEvidence{KeyID: keyID, LookupHMAC: lookup,
		SourceSHA256: normalized.NormalizedSHA256()}, nil
}
