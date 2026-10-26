package httpapi

import (
	"fmt"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

func NewGlobalCommandAccessEvidence(
	principalID uuid.UUID,
	tokenVersion, authorizationVersion, authorizationClock int64,
	permission Permission,
	capability platformtx.GlobalCommandAccess,
) (Access, error) {
	access, err := NewGlobalAccess(principalID, authorizationClock, permission)
	if err != nil || tokenVersion < 1 || authorizationVersion < 1 {
		return Access{}, fmt.Errorf("billing global command access evidence is invalid")
	}
	if _, _, valid := capability.Capabilities(); !valid {
		return Access{}, fmt.Errorf("billing global command access evidence is invalid")
	}
	access.tokenVersion = tokenVersion
	access.authorizationVersion = authorizationVersion
	access.globalCommandAccess = capability
	access.globalMode = globalCommandAccess
	return access, nil
}

func (access Access) GlobalCommandAccess() (platformtx.GlobalCommandAccess, bool) {
	_, _, valid := access.globalCommandAccess.Capabilities()
	return access.globalCommandAccess, access.globalMode == globalCommandAccess && valid
}

func (access Access) GlobalReadbackSnapshotCapability() (
	platformtx.GlobalSnapshotCapability,
	bool,
) {
	_, snapshot, valid := access.globalCommandAccess.Capabilities()
	return snapshot, access.globalMode == globalCommandAccess && valid
}

func (access Access) GlobalCommandCapabilitiesFor(
	operation string,
) (platformtx.GlobalCommandCapability, platformtx.GlobalSnapshotCapability, bool) {
	commandAccess, available := access.GlobalCommandAccess()
	if !available || !commandAccess.Allows(
		access.principalID, operation, string(access.permission),
		access.authorizationClock, access.authorizationVersion,
	) {
		return platformtx.GlobalCommandCapability{}, platformtx.GlobalSnapshotCapability{}, false
	}
	return commandAccess.Capabilities()
}
