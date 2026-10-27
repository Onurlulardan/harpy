package platformiam

import (
	"crypto/sha256"
	"strings"
	"time"
)

type OperatorGrant struct {
	SubjectID, RoleID, SubjectRoleBindingID, RolePermissionBindingID ID
	TargetUserID                                                     ID
	TargetExpectedAuthorizationVersion                               int64
	Capability                                                       string
	Evidence                                                         Evidence
	StepUp                                                           StepUpEvidence
}

func (command OperatorGrant) Validate(now time.Time) error {
	if !command.SubjectID.Valid() || !command.RoleID.Valid() ||
		!command.SubjectRoleBindingID.Valid() || !command.RolePermissionBindingID.Valid() ||
		!command.TargetUserID.Valid() || command.TargetUserID == command.Evidence.ActorID ||
		command.TargetExpectedAuthorizationVersion < 1 ||
		!CapabilityKnown(command.Capability) || command.Evidence.Validate() != nil ||
		command.StepUp.Validate(now, OperatorPurpose("grant", command.TargetUserID, command.Capability)) != nil {
		return ErrInvalidCommand
	}
	return nil
}

func OperatorPurpose(operation string, target ID, capability string) [32]byte {
	return sha256.Sum256([]byte(strings.Join([]string{
		"platform.operators." + operation, target.String(), capability,
	}, "\n")))
}
