package recoverykeymaterial

import (
	"bytes"
	"crypto/ecdh"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"regexp"
)

var canonicalUUIDPattern = regexp.MustCompile(
	`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`,
)

func Parse(document []byte, expected Expected) (Material, error) {
	if len(document) == 0 || len(document) > MaximumDocumentBytes || !validExpected(expected) {
		return Material{}, ErrInvalid
	}
	var wire materialWire
	decoder := json.NewDecoder(bytes.NewReader(document))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&wire) != nil || requireEOF(decoder) != nil ||
		!validWire(wire, expected) {
		return Material{}, ErrInvalid
	}
	canonical, err := canonicalValue(wire)
	if err != nil || !bytes.Equal(canonical, document) {
		return Material{}, ErrInvalid
	}
	drKEK, drValid := decodeKey(wire.Material.DRKEK.Bytes)
	recoveryPublic, recoveryValid := decodeKey(wire.Material.RecoveryPublicKey.Bytes)
	if !drValid || !recoveryValid {
		return Material{}, ErrInvalid
	}
	if _, err = ecdh.X25519().NewPublicKey(recoveryPublic[:]); err != nil {
		return Material{}, ErrInvalid
	}
	documentDigest := sha256.Sum256(canonical)
	return Material{
		purpose: wire.Purpose, providerKind: wire.ProviderKind,
		providerHandle: wire.ProviderHandle, environmentID: wire.EnvironmentID,
		generation: wire.Generation, drKeyID: wire.Material.DRKEK.KeyID,
		recoveryKeyID:  wire.Material.RecoveryPublicKey.KeyID,
		documentSHA256: hex.EncodeToString(documentDigest[:]),
		drKEK:          drKEK, recoveryPublic: recoveryPublic,
	}, nil
}

func validWire(value materialWire, expected Expected) bool {
	body, err := canonicalValue(value.Material)
	digest := sha256.Sum256(body)
	return err == nil && value.Schema == Schema && !value.Synthetic &&
		value.Purpose == expected.Purpose &&
		value.ProviderKind == expected.ProviderKind &&
		value.ProviderHandle == expected.ProviderHandle &&
		value.EnvironmentID == expected.EnvironmentID &&
		value.Generation == expected.Generation &&
		value.Material.DRKEK.Algorithm == "AES-256" &&
		value.Material.RecoveryPublicKey.Algorithm == "X25519" &&
		validKeyID(value.Material.DRKEK.KeyID) &&
		validKeyID(value.Material.RecoveryPublicKey.KeyID) &&
		value.Material.DRKEK.KeyID != value.Material.RecoveryPublicKey.KeyID &&
		value.MaterialSHA256 == hex.EncodeToString(digest[:])
}

func validExpected(value Expected) bool {
	return canonicalUUIDPattern.MatchString(value.EnvironmentID) &&
		value.EnvironmentID != "00000000-0000-0000-0000-000000000000" &&
		value.Purpose == PurposeComponentSealing && value.Generation > 0 &&
		validProvider(value.ProviderKind, value.ProviderHandle)
}

func decodeKey(value string) ([32]byte, bool) {
	var result [32]byte
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(value)
	if err != nil || len(decoded) != len(result) ||
		base64.RawURLEncoding.EncodeToString(decoded) != value {
		return result, false
	}
	copy(result[:], decoded)
	return result, result != [32]byte{}
}

func canonicalValue(value any) ([]byte, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	var normalized any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if decoder.Decode(&normalized) != nil || requireEOF(decoder) != nil {
		return nil, ErrInvalid
	}
	canonical, err := json.Marshal(normalized)
	if err != nil || len(canonical) > MaximumDocumentBytes ||
		len(canonical) < 2 || canonical[0] != '{' {
		return nil, ErrInvalid
	}
	return canonical, nil
}

func requireEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return ErrInvalid
	}
	return nil
}
