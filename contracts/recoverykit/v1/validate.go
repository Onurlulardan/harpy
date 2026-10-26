package recoverykitv1

import (
	"crypto/ecdh"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"regexp"
	"time"

	recoverysigningv1 "github.com/Onurlulardan/harpy/contracts/recoverysigning/v1"
	"github.com/google/uuid"
)

var keyIDPattern = regexp.MustCompile(`^[a-z][a-z0-9._:-]{0,127}$`)

func validUUID(value string) bool {
	parsed, err := uuid.Parse(value)
	return err == nil && parsed != uuid.Nil && parsed.String() == value
}

func validRecoveryKeys(keys []RecoveryKeyMaterial) bool {
	if len(keys) == 0 || len(keys) > 8 {
		return false
	}
	seen, active := make(map[string]bool, len(keys)), 0
	for _, key := range keys {
		if !keyIDPattern.MatchString(key.KeyID) || seen[key.KeyID] ||
			(key.State != "active" && key.State != "retained") ||
			key.PrivateKey == [32]byte{} {
			return false
		}
		if _, err := ecdh.X25519().NewPrivateKey(key.PrivateKey[:]); err != nil {
			return false
		}
		seen[key.KeyID] = true
		if key.State == "active" {
			active++
		}
	}
	return active == 1
}

func validatePlaintext(value plaintextWire) bool {
	generatedAt, err := time.Parse(canonicalTimeFormat, value.GeneratedAt)
	binding, bindingErr := hex.DecodeString(value.DeploymentBindingSHA256)
	ack, ackValid := decode(value.AckProofBase64URL, 32)
	bundleBytes, bundleErr := base64.RawURLEncoding.Strict().DecodeString(
		value.RecoveryTrustBundleBase64URL)
	bundle, parseErr := recoverysigningv1.Parse(bundleBytes)
	valid := err == nil && canonicalTime(generatedAt) && bindingErr == nil && len(binding) == 32 &&
		ackValid && len(ack) == 32 && bundleErr == nil && parseErr == nil &&
		value.Schema == PlaintextSchema && validUUID(value.EnvironmentID) &&
		validUUID(value.OperationID) && value.Generation > 0 &&
		bundle.EnvironmentID() == value.EnvironmentID && bundle.Generation() == value.Generation &&
		validateRecoveryKeyWire(value.RecoveryKeys) && validateSigningKeyWire(value.SigningKeys, bundle)
	clear(binding)
	clear(ack)
	clear(bundleBytes)
	if !valid || value.PreviousKitSHA256 == nil {
		return valid
	}
	previous, previousErr := hex.DecodeString(*value.PreviousKitSHA256)
	defer clear(previous)
	return previousErr == nil && len(previous) == 32
}

func validateRecoveryKeyWire(keys []recoveryKeyWire) bool {
	materials := make([]RecoveryKeyMaterial, len(keys))
	for index, key := range keys {
		private, valid := decode(key.PrivateKey, 32)
		if !valid || key.Algorithm != "X25519" {
			clear(private)
			return false
		}
		materials[index] = RecoveryKeyMaterial{KeyID: key.KeyID, State: key.State}
		copy(materials[index].PrivateKey[:], private)
		clear(private)
	}
	return validRecoveryKeys(materials)
}

func validateSigningKeyWire(keys []signingKeyWire, bundle recoverysigningv1.Bundle) bool {
	materials := make([]SigningKeyMaterial, len(keys))
	for index, key := range keys {
		private, valid := decode(key.PrivateKey, ed25519.PrivateKeySize)
		if !valid || key.Algorithm != "Ed25519" {
			clear(private)
			return false
		}
		materials[index] = SigningKeyMaterial{Purpose: key.Purpose,
			KeyID: key.KeyID, PrivateKey: ed25519.PrivateKey(private)}
	}
	return validSigningKeys(materials, bundle)
}
