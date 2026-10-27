package transaction

import (
	"fmt"

	"github.com/google/uuid"
)

// GlobalCommandAccess is an opaque, indivisible command and readback proof.
// Both capabilities are minted from one IAM decision and one generation fence.
type GlobalCommandAccess struct {
	command  GlobalCommandCapability
	readback GlobalSnapshotCapability
	issued   bool
}

type GlobalCommandAccessIssuer struct{ valid bool }

func NewGlobalCommandAccessIssuer() GlobalCommandAccessIssuer {
	return GlobalCommandAccessIssuer{valid: true}
}

func (issuer GlobalCommandAccessIssuer) Issue(
	actorID uuid.UUID,
	operationID, permission string,
	authorizationClock, authorizationVersion int64,
	groupGeneration, billingGeneration int64,
) (GlobalCommandAccess, error) {
	command, err := NewGlobalCommandCapabilityIssuer().Issue(
		actorID, operationID, permission, authorizationClock,
		authorizationVersion, groupGeneration, billingGeneration,
	)
	if err != nil {
		return GlobalCommandAccess{}, fmt.Errorf("global command access is invalid")
	}
	readback, err := NewGlobalSnapshotCapabilityIssuer().Issue(
		actorID, operationID, permission, authorizationClock,
		authorizationVersion, groupGeneration, billingGeneration,
	)
	access := GlobalCommandAccess{command: command, readback: readback, issued: true}
	if err != nil || !issuer.valid || !access.valid() {
		return GlobalCommandAccess{}, fmt.Errorf("global command access is invalid")
	}
	return access, nil
}

func (access GlobalCommandAccess) Capabilities() (
	GlobalCommandCapability,
	GlobalSnapshotCapability,
	bool,
) {
	if !access.valid() {
		return GlobalCommandCapability{}, GlobalSnapshotCapability{}, false
	}
	return access.command, access.readback, true
}

func (access GlobalCommandAccess) valid() bool {
	command, snapshot := access.command, access.readback
	return access.issued && command.valid() && snapshot.valid() &&
		command.actorID == snapshot.actorID &&
		command.operationID == snapshot.operationID &&
		command.permission == snapshot.permission &&
		command.authorizationClock == snapshot.authorizationClock &&
		command.authorizationVersion == snapshot.authorizationVersion &&
		command.groupGeneration == snapshot.groupGeneration &&
		command.billingGeneration == snapshot.billingGeneration
}

func (access GlobalCommandAccess) Allows(
	actorID uuid.UUID,
	operationID, permission string,
	authorizationClock, authorizationVersion int64,
) bool {
	return access.valid() && access.command.actorID == actorID &&
		access.command.operationID == operationID && access.command.permission == permission &&
		access.command.authorizationClock == authorizationClock &&
		access.command.authorizationVersion == authorizationVersion
}
