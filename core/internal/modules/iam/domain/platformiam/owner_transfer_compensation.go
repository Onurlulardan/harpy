package platformiam

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"
)

type OwnerTransferCompensation struct {
	Evidence
	OriginalOperationID        ID
	OriginalAckSHA256          [32]byte
	ProposalSHA256             [32]byte
	ProposalSignerKID          string
	ProposalSignatureSHA256    [32]byte
	ProposalExpiresAt          time.Time
	ExpectedOriginalGeneration int64
	StepUp                     StepUpEvidence
}

func (command OwnerTransferCompensation) Validate(now time.Time) error {
	if command.Evidence.Validate() != nil || !command.OriginalOperationID.Valid() ||
		command.OriginalOperationID == command.OperationID || command.OriginalAckSHA256 == [32]byte{} ||
		command.ProposalSHA256 == [32]byte{} || command.ProposalSignatureSHA256 == [32]byte{} ||
		!validKID(command.ProposalSignerKID) || command.ExpectedOriginalGeneration < 2 ||
		command.ProposalExpiresAt.Before(now) || command.ProposalExpiresAt.After(now.Add(7*24*time.Hour)) ||
		command.StepUp.Validate(now, OwnerTransferCompensationPurpose(command)) != nil {
		return ErrInvalidCommand
	}
	return nil
}

func OwnerTransferCompensationPurpose(command OwnerTransferCompensation) [32]byte {
	return sha256.Sum256([]byte(strings.Join([]string{
		"platform.owner.transfer.compensate", command.OperationID.String(),
		command.OriginalOperationID.String(), strconv.FormatInt(command.ExpectedOriginalGeneration+1, 10),
		hex.EncodeToString(command.OriginalAckSHA256[:]),
	}, "\n")))
}
