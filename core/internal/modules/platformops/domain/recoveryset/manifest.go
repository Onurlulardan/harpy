package recoveryset

import (
	"bytes"
	"encoding/json"
	"io"
	"regexp"
	"strings"
)

const (
	ManifestSchema       = "environment-recovery-set-payload.v1"
	MaximumManifestBytes = 256 * 1024
)

var (
	digestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)
	uuidPattern   = regexp.MustCompile(
		`^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`,
	)
)

func (manifest ManifestPayload) CanonicalBytes() ([]byte, error) {
	if !manifest.valid() {
		return nil, ErrManifest
	}
	return canonicalObject(manifest, MaximumManifestBytes)
}

func ParseManifest(document []byte) (ManifestPayload, error) {
	if len(document) == 0 || len(document) > MaximumManifestBytes {
		return ManifestPayload{}, ErrManifest
	}
	var manifest ManifestPayload
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&manifest) != nil || decoder.Decode(new(any)) != io.EOF || !manifest.valid() {
		return ManifestPayload{}, ErrManifest
	}
	canonical, err := manifest.CanonicalBytes()
	if err != nil || !bytes.Equal(canonical, document) {
		return ManifestPayload{}, ErrManifest
	}
	return manifest, nil
}

func (manifest ManifestPayload) valid() bool {
	return manifest.Schema == ManifestSchema &&
		(manifest.Portability == SameEnvironment || manifest.Portability == Portable) &&
		manifest.Identity.valid() && validUUID(manifest.DatabaseSystemID) &&
		manifest.Components.valid() &&
		manifest.DatabaseArchiveSHA256 ==
			manifest.Components.DatabaseArchive.CiphertextSHA256 &&
		manifest.KeyCustodySHA256 ==
			manifest.Components.KeyCustody.CiphertextSHA256 &&
		manifest.OwnerAnchorSHA256 ==
			manifest.Components.OwnerAnchor.CiphertextSHA256 &&
		manifest.ProtectedHeadAckSHA256 ==
			manifest.Components.ProtectedAuthority.CiphertextSHA256 &&
		validDigests(manifest.DatabaseArchiveSHA256, manifest.DocumentsMerkleSHA256,
			manifest.KeyCustodySHA256, manifest.OwnerAnchorSHA256,
			manifest.ProtectedHeadAckSHA256, manifest.RecoveryTrustBundleSHA256) &&
		manifest.RecoverySigningGeneration > 0 && canonicalTime(manifest.CreatedAt)
}

func (components RecoveryComponents) valid() bool {
	return components.DatabaseArchive.valid() &&
		components.DocumentsInventory.valid() &&
		components.KeyCustody.valid() &&
		components.OwnerAnchor.valid() &&
		components.ProtectedAuthority.valid()
}

func (component ComponentIdentity) valid() bool {
	return len(component.ObjectKey) >= 1 && len(component.ObjectKey) <= 1024 &&
		!strings.ContainsAny(component.ObjectKey, "\x00\r\n") &&
		len(component.ObjectVersionID) >= 1 &&
		len(component.ObjectVersionID) <= 1024 &&
		!strings.ContainsAny(component.ObjectVersionID, "\x00\r\n") &&
		digestPattern.MatchString(component.CiphertextSHA256) &&
		component.CiphertextBytes > 0
}

func (identity EnvironmentIdentity) valid() bool {
	return validUUID(identity.EnvironmentID) && identity.DeploymentMode == ModeOnPremise &&
		validUUID(identity.PrimaryBindingID) && digestPattern.MatchString(identity.TenantSetSHA256) &&
		validUUID(identity.OwnerBindingID) && identity.OwnerBindingGeneration > 0 &&
		validUUID(identity.DatabaseLineageID)
}

func ValidateIdentity(identity EnvironmentIdentity) error {
	if !identity.valid() {
		return ErrManifest
	}
	return nil
}

func ValidateManifest(manifest ManifestPayload) error {
	if !manifest.valid() {
		return ErrManifest
	}
	return nil
}

func validUUID(value string) bool {
	return uuidPattern.MatchString(value)
}

func validDigests(values ...string) bool {
	for _, value := range values {
		if !digestPattern.MatchString(value) {
			return false
		}
	}
	return true
}
