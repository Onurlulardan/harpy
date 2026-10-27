package platformiam

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"
)

type OwnerTransferAcknowledge struct {
	Evidence
	Acknowledgement AuthorityAcknowledgement
	StepUp          StepUpEvidence
}

func (command OwnerTransferAcknowledge) Validate(now time.Time) error {
	if command.Evidence.Validate() != nil ||
		command.Acknowledgement.OperationID != command.OperationID ||
		command.Acknowledgement.Validate(now) != nil ||
		command.StepUp.Validate(now, OwnerTransferAckPurpose(command)) != nil {
		return ErrInvalidCommand
	}
	return nil
}

func OwnerTransferAckPurpose(command OwnerTransferAcknowledge) [32]byte {
	return sha256.Sum256([]byte(strings.Join([]string{
		"platform.owner.transfer.ack", command.OperationID.String(),
		hex.EncodeToString(command.Acknowledgement.AckSHA256[:]),
	}, "\n")))
}
