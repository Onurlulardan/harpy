package api

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

const (
	recoveryStepUpDomain   = "harpy/platform-recovery/step-up/v1\x00"
	operationsStepUpDomain = "harpy/platform-operations/step-up/v1\x00"
)

var ErrInvalidStepUpPurpose = errors.New("platform recovery step-up purpose is invalid")

// StepUpPurpose binds reauthentication to one recovery capability and UUID target.
func StepUpPurpose(
	tenantID string,
	capability string,
	targetID string,
) ([sha256.Size]byte, error) {
	domain, valid := stepUpDomain(capability)
	tenantBytes, _, tenantValid := purposeUUID(tenantID)
	_, canonicalTarget, targetValid := purposeUUID(targetID)
	if !tenantValid || !targetValid || !valid {
		return [sha256.Size]byte{}, ErrInvalidStepUpPurpose
	}
	hash := sha256.New()
	_, _ = hash.Write([]byte(domain))
	_, _ = hash.Write(tenantBytes[:])
	_, _ = hash.Write([]byte(capability))
	_, _ = hash.Write([]byte{0})
	_, _ = hash.Write([]byte(canonicalTarget))
	var result [sha256.Size]byte
	copy(result[:], hash.Sum(nil))
	return result, nil
}

func purposeUUID(value string) ([16]byte, string, bool) {
	var result [16]byte
	if len(value) != 36 || value[8] != '-' || value[13] != '-' ||
		value[18] != '-' || value[23] != '-' {
		return result, "", false
	}
	compact := strings.ReplaceAll(value, "-", "")
	decoded, err := hex.DecodeString(compact)
	if err != nil || len(decoded) != len(result) {
		return result, "", false
	}
	copy(result[:], decoded)
	if result == [16]byte{} {
		return result, "", false
	}
	canonicalHex := hex.EncodeToString(result[:])
	canonical := canonicalHex[:8] + "-" + canonicalHex[8:12] + "-" +
		canonicalHex[12:16] + "-" + canonicalHex[16:20] + "-" + canonicalHex[20:]
	return result, canonical, true
}

func stepUpDomain(capability string) (string, bool) {
	if capability == "platform.operations.permission_cache.flush" {
		return operationsStepUpDomain, true
	}
	return recoveryStepUpDomain, validRecoveryStepUpCapability(capability)
}

func validRecoveryStepUpCapability(value string) bool {
	if strings.TrimSpace(value) != value {
		return false
	}
	switch value {
	case "platform.recovery.enable", "platform.recovery.rotate",
		"platform.recovery.disable", "platform.recovery.anchor.rotate",
		"platform.recovery_sets.create", "platform.recovery_sets.manifest_download",
		"platform.recovery_sets.delete", "platform.recovery_sets.restore",
		"platform.restore_operations.plan_download":
		return true
	default:
		return false
	}
}
