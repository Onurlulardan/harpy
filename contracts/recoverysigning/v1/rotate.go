package recoverysigningv1

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
)

func Rotate(
	previous Bundle,
	generatedAt time.Time,
	material []PublicMaterial,
) (Bundle, error) {
	if previous.SHA256() == "" || !microsecondUTC(generatedAt) ||
		!generatedAt.After(previous.GeneratedAt()) || len(material) != len(orderedPurposes) {
		return Bundle{}, fmt.Errorf("recovery signing bundle rotation is invalid")
	}
	active, err := nextActiveKeys(previous, generatedAt, material)
	if err != nil {
		return Bundle{}, err
	}
	retired := append([]keyWire(nil), previous.wire.Retired...)
	for _, value := range previous.wire.Keys {
		signUntil := generatedAt.Format(time.RFC3339Nano)
		value.Status = "verify_only"
		value.SignUntil = &signUntil
		retired = append(retired, value)
	}
	wire := bundleWire{Schema: BundleSchema, EnvironmentID: previous.EnvironmentID(),
		Generation: previous.Generation() + 1, Algorithm: Algorithm,
		GeneratedAt: generatedAt.Format(time.RFC3339Nano), Keys: active, Retired: retired}
	raw, err := json.Marshal(wire)
	if err != nil {
		return Bundle{}, fmt.Errorf("recovery signing bundle rotation encode failed")
	}
	canonical, err := canonicaljsonv1.CanonicalizeObject(raw, MaximumBundleBytes)
	if err != nil {
		return Bundle{}, fmt.Errorf("recovery signing bundle rotation encode failed")
	}
	return parseCanonical(canonical)
}

func nextActiveKeys(
	previous Bundle,
	generatedAt time.Time,
	material []PublicMaterial,
) ([]keyWire, error) {
	byPurpose := make(map[Purpose][32]byte, len(material))
	seen := make(map[string]bool, len(material)+len(previous.keys)+len(previous.retired))
	for _, value := range previous.keys {
		seen[string(value.publicKey[:])] = true
	}
	for _, value := range previous.retired {
		seen[string(value.publicKey[:])] = true
	}
	for _, value := range material {
		identity := string(value.PublicKey[:])
		if !validPurpose(value.Purpose) || value.PublicKey == [32]byte{} ||
			seen[identity] || byPurpose[value.Purpose] != [32]byte{} {
			return nil, fmt.Errorf("recovery signing bundle rotation is invalid")
		}
		seen[identity] = true
		byPurpose[value.Purpose] = value.PublicKey
	}
	result := make([]keyWire, 0, len(orderedPurposes))
	for _, purpose := range orderedPurposes {
		publicKey := byPurpose[purpose]
		if publicKey == [32]byte{} {
			return nil, fmt.Errorf("recovery signing bundle rotation is invalid")
		}
		result = append(result, keyWire{Purpose: purpose,
			KeyID:     makeKeyID(purpose, previous.Generation()+1, publicKey),
			Algorithm: Algorithm, Status: "active",
			PublicKey: base64.RawURLEncoding.EncodeToString(publicKey[:]),
			NotBefore: generatedAt.Format(time.RFC3339Nano)})
	}
	return result, nil
}
