package platformiam

import "time"

type OperatorRevoke struct {
	TargetUserID                       ID
	TargetExpectedAuthorizationVersion int64
	Capability                         string
	Evidence                           Evidence
	StepUp                             StepUpEvidence
}

func (command OperatorRevoke) Validate(now time.Time) error {
	if !command.TargetUserID.Valid() || command.TargetUserID == command.Evidence.ActorID ||
		command.TargetExpectedAuthorizationVersion < 1 ||
		!CapabilityKnown(command.Capability) || command.Evidence.Validate() != nil ||
		command.StepUp.Validate(now, OperatorPurpose("revoke", command.TargetUserID, command.Capability)) != nil {
		return ErrInvalidCommand
	}
	return nil
}
