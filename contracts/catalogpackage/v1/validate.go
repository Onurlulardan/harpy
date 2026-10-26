package catalogpackagev1

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"regexp"
	"strconv"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

var (
	releaseIDPattern = regexp.MustCompile(
		`^[0-9a-f]{8}-[0-9a-f]{4}-[1-8][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`,
	)
	stableIDPattern = regexp.MustCompile(
		`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`,
	)
	digestPattern   = regexp.MustCompile(`^[0-9a-f]{64}$`)
	keyIDPattern    = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,127}$`)
	ownerPattern    = regexp.MustCompile(`^[a-z][a-z0-9_]{0,62}$`)
	contractPattern = regexp.MustCompile(`^[a-z][a-z0-9.-]{2,63}$`)
	semverPattern   = regexp.MustCompile(
		`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$`,
	)
)

func (set ReleaseSet) Validate() error {
	if set.ContractVersion != ContractVersion ||
		!releaseIDPattern.MatchString(set.ReleaseID) ||
		len(set.Packages) < 1 || len(set.Packages) > 128 {
		return invalidSet()
	}
	manifestDigests := make(map[string]struct{}, len(set.Packages))
	componentIDs := make(map[string]struct{})
	var previous packageOrder
	for index, sealed := range set.Packages {
		order, err := validatePackage(sealed, manifestDigests, componentIDs)
		if err != nil {
			return err
		}
		if index > 0 && !previous.less(order) {
			return invalidSet()
		}
		previous = order
	}
	return nil
}

func validatePackage(
	sealed SealedPackage,
	manifestDigests, componentIDs map[string]struct{},
) (packageOrder, error) {
	canonical, err := canonicaljsonv1.CanonicalizeObject(
		sealed.Manifest, MaximumManifestBytes,
	)
	if err != nil || !bytes.Equal(canonical, sealed.Manifest) ||
		!digestPattern.MatchString(sealed.ManifestSHA256) ||
		!keyIDPattern.MatchString(sealed.KeyID) ||
		len(sealed.OwnerPayloads) < 1 || len(sealed.OwnerPayloads) > 256 {
		return packageOrder{}, invalidSet()
	}
	digest, err := canonicaljsonv1.DigestHex(sealed.Manifest, MaximumManifestBytes)
	signature, signatureErr := base64.RawURLEncoding.DecodeString(sealed.Signature)
	if err != nil || digest != sealed.ManifestSHA256 ||
		signatureErr != nil || len(signature) != ed25519.SignatureSize ||
		base64.RawURLEncoding.EncodeToString(signature) != sealed.Signature {
		return packageOrder{}, invalidSet()
	}
	if _, duplicate := manifestDigests[digest]; duplicate {
		return packageOrder{}, invalidSet()
	}
	manifestDigests[digest] = struct{}{}
	order, err := decodePackageOrder(sealed.Manifest)
	if err != nil {
		return packageOrder{}, err
	}
	previousComponentID := ""
	for _, payload := range sealed.OwnerPayloads {
		if err := validatePayload(payload, componentIDs); err != nil ||
			previousComponentID != "" &&
				previousComponentID >= payload.ComponentID {
			return packageOrder{}, invalidSet()
		}
		previousComponentID = payload.ComponentID
	}
	return order, nil
}

type packageOrder struct {
	id      string
	version [3]uint64
}

func decodePackageOrder(document []byte) (packageOrder, error) {
	var identity struct {
		PackageID string `json:"package_id"`
		Version   string `json:"version"`
	}
	if json.Unmarshal(document, &identity) != nil ||
		!stableIDPattern.MatchString(identity.PackageID) ||
		!semverPattern.MatchString(identity.Version) {
		return packageOrder{}, invalidSet()
	}
	matches := semverPattern.FindStringSubmatch(identity.Version)
	var result packageOrder
	result.id = identity.PackageID
	for index := range result.version {
		value, err := strconv.ParseUint(matches[index+1], 10, 64)
		if err != nil {
			return packageOrder{}, invalidSet()
		}
		result.version[index] = value
	}
	return result, nil
}

func (left packageOrder) less(right packageOrder) bool {
	if left.id != right.id {
		return left.id < right.id
	}
	for index := range left.version {
		if left.version[index] != right.version[index] {
			return left.version[index] < right.version[index]
		}
	}
	return false
}

func validatePayload(payload OwnerPayload, componentIDs map[string]struct{}) error {
	if !stableIDPattern.MatchString(payload.ComponentID) ||
		!ownerPattern.MatchString(payload.Owner) ||
		!contractPattern.MatchString(payload.Contract) ||
		!digestPattern.MatchString(payload.ContentSHA256) {
		return invalidSet()
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(
		payload.Document, MaximumPayloadBytes,
	)
	digest, digestErr := canonicaljsonv1.DigestHex(
		payload.Document, MaximumPayloadBytes,
	)
	if err != nil || !bytes.Equal(canonical, payload.Document) ||
		digestErr != nil || digest != payload.ContentSHA256 {
		return invalidSet()
	}
	if _, duplicate := componentIDs[payload.ComponentID]; duplicate {
		return invalidSet()
	}
	componentIDs[payload.ComponentID] = struct{}{}
	return nil
}

func invalidSet() error {
	return errInvalidPackageSet
}
