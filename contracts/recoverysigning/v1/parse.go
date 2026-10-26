package recoverysigningv1

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

func Parse(document []byte) (Bundle, error) {
	canonical, err := canonicaljsonv1.CanonicalizeObject(document, MaximumBundleBytes)
	if err != nil || !bytes.Equal(canonical, document) {
		return Bundle{}, fmt.Errorf("recovery signing bundle is invalid")
	}
	return parseCanonical(canonical)
}

func parseCanonical(canonical []byte) (Bundle, error) {
	var wire bundleWire
	decoder := json.NewDecoder(bytes.NewReader(canonical))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&wire) != nil || requireJSONEOF(decoder) != nil ||
		wire.Schema != BundleSchema || wire.Generation < Generation ||
		wire.Algorithm != Algorithm || !validUUID(wire.EnvironmentID) ||
		len(wire.Keys) != len(orderedPurposes) {
		return Bundle{}, fmt.Errorf("recovery signing bundle is invalid")
	}
	generatedAt, err := parseTime(wire.GeneratedAt)
	if err != nil {
		return Bundle{}, fmt.Errorf("recovery signing bundle is invalid")
	}
	keys := make(map[Purpose]keyMaterial, len(wire.Keys))
	seenKeys := make(map[string]bool, len(wire.Keys)+len(wire.Retired))
	for index, value := range wire.Keys {
		material, valid := validateActiveKey(
			value, orderedPurposes[index], wire.Generation, generatedAt,
		)
		if !valid {
			return Bundle{}, fmt.Errorf("recovery signing bundle is invalid")
		}
		identity := string(material.publicKey[:])
		if seenKeys[identity] {
			return Bundle{}, fmt.Errorf("recovery signing bundle is invalid")
		}
		seenKeys[identity] = true
		keys[value.Purpose] = material
	}
	retired := make(map[string]keyMaterial, len(wire.Retired))
	for index, value := range wire.Retired {
		material, valid := validateRetiredKey(value, wire.Generation, generatedAt)
		identity := string(material.publicKey[:])
		if !valid || retired[material.id].id != "" || seenKeys[identity] ||
			(index > 0 && !retiredKeyBefore(wire.Retired[index-1], value)) {
			return Bundle{}, fmt.Errorf("recovery signing bundle is invalid")
		}
		seenKeys[identity] = true
		retired[material.id] = material
	}
	if (wire.Generation == Generation) != (len(retired) == 0) ||
		!validRetiredCoverage(wire.Retired, wire.Generation) {
		return Bundle{}, fmt.Errorf("recovery signing bundle is invalid")
	}
	digest := sha256.Sum256(canonical)
	return Bundle{wire: wire, keys: keys, retired: retired,
		canonical: append([]byte(nil), canonical...),
		digest:    hex.EncodeToString(digest[:]), createdAt: generatedAt}, nil
}

func decodePublicKey(value string) ([32]byte, bool) {
	var result [32]byte
	decoded, err := base64.RawURLEncoding.DecodeString(value)
	if err != nil || len(decoded) != len(result) {
		return result, false
	}
	copy(result[:], decoded)
	return result, result != [32]byte{}
}
