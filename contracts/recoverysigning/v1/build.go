package recoverysigningv1

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

func Build(
	environmentID string,
	generatedAt time.Time,
	material []PublicMaterial,
) (Bundle, error) {
	if !validUUID(environmentID) || !microsecondUTC(generatedAt) ||
		len(material) != len(orderedPurposes) {
		return Bundle{}, fmt.Errorf("recovery signing bundle input is invalid")
	}
	byPurpose := make(map[Purpose][32]byte, len(material))
	seenKeys := make(map[string]bool, len(material))
	for _, value := range material {
		if !validPurpose(value.Purpose) || value.PublicKey == [32]byte{} {
			return Bundle{}, fmt.Errorf("recovery signing bundle input is invalid")
		}
		if _, duplicate := byPurpose[value.Purpose]; duplicate {
			return Bundle{}, fmt.Errorf("recovery signing bundle input is invalid")
		}
		keyIdentity := string(value.PublicKey[:])
		if seenKeys[keyIdentity] {
			return Bundle{}, fmt.Errorf("recovery signing bundle input is invalid")
		}
		seenKeys[keyIdentity] = true
		byPurpose[value.Purpose] = value.PublicKey
	}
	wire := bundleWire{Schema: BundleSchema, EnvironmentID: environmentID,
		Generation: Generation, Algorithm: Algorithm,
		GeneratedAt: generatedAt.Format(time.RFC3339Nano)}
	for _, purpose := range orderedPurposes {
		public, found := byPurpose[purpose]
		if !found {
			return Bundle{}, fmt.Errorf("recovery signing bundle input is invalid")
		}
		wire.Keys = append(wire.Keys, keyWire{Purpose: purpose,
			KeyID:     makeKeyID(purpose, Generation, public),
			Algorithm: Algorithm, Status: "active",
			PublicKey: base64.RawURLEncoding.EncodeToString(public[:]),
			NotBefore: generatedAt.Format(time.RFC3339Nano)})
	}
	raw, err := json.Marshal(wire)
	if err != nil {
		return Bundle{}, fmt.Errorf("recovery signing bundle encode failed")
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(raw, MaximumBundleBytes)
	if err != nil {
		return Bundle{}, fmt.Errorf("recovery signing bundle encode failed")
	}
	return parseCanonical(canonical)
}
