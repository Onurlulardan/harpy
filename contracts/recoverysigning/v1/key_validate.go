package recoverysigningv1

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"
)

func validateActiveKey(
	value keyWire,
	purpose Purpose,
	generation int64,
	generatedAt time.Time,
) (keyMaterial, bool) {
	publicKey, valid := decodePublicKey(value.PublicKey)
	material := keyMaterial{id: value.KeyID, purpose: value.Purpose, publicKey: publicKey}
	return material, valid && value.Purpose == purpose &&
		value.KeyID == makeKeyID(purpose, generation, publicKey) &&
		value.Algorithm == Algorithm && value.Status == "active" &&
		value.NotBefore == generatedAt.Format(time.RFC3339Nano) &&
		value.SignUntil == nil && value.VerifyUntil == nil
}

func validateRetiredKey(
	value keyWire,
	activeGeneration int64,
	generatedAt time.Time,
) (keyMaterial, bool) {
	publicKey, valid := decodePublicKey(value.PublicKey)
	generation, generationOK := keyGeneration(value.KeyID, value.Purpose, publicKey)
	notBefore, notBeforeErr := parseTime(value.NotBefore)
	if value.SignUntil == nil {
		return keyMaterial{}, false
	}
	signUntil, signUntilErr := parseTime(*value.SignUntil)
	material := keyMaterial{id: value.KeyID, purpose: value.Purpose, publicKey: publicKey}
	return material, valid && validPurpose(value.Purpose) && generationOK &&
		generation > 0 && generation < activeGeneration &&
		value.Algorithm == Algorithm && value.Status == "verify_only" &&
		notBeforeErr == nil && signUntilErr == nil && notBefore.Before(signUntil) &&
		!signUntil.After(generatedAt) && value.VerifyUntil == nil
}

func makeKeyID(purpose Purpose, generation int64, publicKey [32]byte) string {
	digest := sha256.Sum256(publicKey[:])
	return string(purpose) + ".g" + strconv.FormatInt(generation, 10) + "." +
		hex.EncodeToString(digest[:8])
}

func keyGeneration(keyID string, purpose Purpose, publicKey [32]byte) (int64, bool) {
	prefix := string(purpose) + ".g"
	digest := sha256.Sum256(publicKey[:])
	suffix := "." + hex.EncodeToString(digest[:8])
	if !strings.HasPrefix(keyID, prefix) || !strings.HasSuffix(keyID, suffix) {
		return 0, false
	}
	value := strings.TrimSuffix(strings.TrimPrefix(keyID, prefix), suffix)
	generation, err := strconv.ParseInt(value, 10, 64)
	return generation, err == nil && strconv.FormatInt(generation, 10) == value
}

func digestPublicKey(publicKey [32]byte) string {
	digest := sha256.Sum256(publicKey[:])
	return hex.EncodeToString(digest[:])
}

func retiredKeyBefore(left, right keyWire) bool {
	leftPublic, leftOK := decodePublicKey(left.PublicKey)
	rightPublic, rightOK := decodePublicKey(right.PublicKey)
	leftGeneration, leftGenerationOK := keyGeneration(left.KeyID, left.Purpose, leftPublic)
	rightGeneration, rightGenerationOK := keyGeneration(right.KeyID, right.Purpose, rightPublic)
	if !leftOK || !rightOK || !leftGenerationOK || !rightGenerationOK {
		return false
	}
	if leftGeneration != rightGeneration {
		return leftGeneration < rightGeneration
	}
	return purposeIndex(left.Purpose) < purposeIndex(right.Purpose)
}

func purposeIndex(value Purpose) int {
	for index, purpose := range orderedPurposes {
		if value == purpose {
			return index
		}
	}
	return len(orderedPurposes)
}

func validRetiredCoverage(values []keyWire, activeGeneration int64) bool {
	if activeGeneration == Generation {
		return len(values) == 0
	}
	if len(values) < len(orderedPurposes) || len(values)%len(orderedPurposes) != 0 {
		return false
	}
	lastGeneration := int64(0)
	for index, value := range values {
		publicKey, valid := decodePublicKey(value.PublicKey)
		generation, generationOK := keyGeneration(value.KeyID, value.Purpose, publicKey)
		position := index % len(orderedPurposes)
		if !valid || !generationOK || value.Purpose != orderedPurposes[position] {
			return false
		}
		if position == 0 {
			if generation <= lastGeneration {
				return false
			}
			lastGeneration = generation
		} else if generation != lastGeneration {
			return false
		}
	}
	return lastGeneration == activeGeneration-1
}
