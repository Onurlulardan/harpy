package platformiam

import (
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"
)

type OwnerTransferPrepare struct {
	Evidence
	TargetUserID                       ID
	ActorExpectedRevision              int64
	TargetExpectedRevision             int64
	TargetExpectedTokenVersion         int64
	TargetExpectedAuthorizationVersion int64
	SuccessorSubjectID                 ID
	SuccessorBindingID                 ID
	RegistryBindingID                  ID
	ExpectedOwnerPolicyRevision        int64
	ExpectedHeadGeneration             int64
	ExpectedHeadSHA256                 [32]byte
	StateSHA256                        [32]byte
	ProposalSHA256                     [32]byte
	ProposalSignerKID                  string
	ProposalSignatureSHA256            [32]byte
	ProposalExpiresAt                  time.Time
	StepUp                             StepUpEvidence
}

func (command OwnerTransferPrepare) Validate(now time.Time) error {
	if command.Evidence.Validate() != nil || !command.TargetUserID.Valid() ||
		command.TargetUserID == command.ActorID || command.ActorExpectedRevision < 1 ||
		command.TargetExpectedRevision < 1 || command.TargetExpectedTokenVersion < 1 ||
		command.TargetExpectedAuthorizationVersion < 1 || !command.SuccessorSubjectID.Valid() ||
		!command.SuccessorBindingID.Valid() || !command.RegistryBindingID.Valid() ||
		command.ExpectedOwnerPolicyRevision < 1 || command.ExpectedHeadGeneration < 1 ||
		command.ExpectedHeadSHA256 == [32]byte{} || command.StateSHA256 == [32]byte{} ||
		command.StateSHA256 == command.ExpectedHeadSHA256 || command.ProposalSHA256 == [32]byte{} ||
		command.ProposalSignatureSHA256 == [32]byte{} || !validKID(command.ProposalSignerKID) ||
		command.ProposalExpiresAt.Before(now) || command.ProposalExpiresAt.After(now.Add(7*24*time.Hour)) ||
		command.StepUp.Validate(now, OwnerTransferPurpose(command)) != nil {
		return ErrInvalidCommand
	}
	return nil
}

func OwnerTransferPurpose(command OwnerTransferPrepare) [32]byte {
	return sha256.Sum256([]byte(strings.Join([]string{
		"platform.owner.transfer", command.OperationID.String(), command.TargetUserID.String(),
		strconv.FormatInt(command.ExpectedHeadGeneration+1, 10),
		hex.EncodeToString(command.StateSHA256[:]),
	}, "\n")))
}

func validKID(value string) bool {
	if len(value) < 1 || len(value) > 128 {
		return false
	}
	for _, character := range value {
		if !(character >= 'a' && character <= 'z') && !(character >= '0' && character <= '9') &&
			character != '.' && character != '_' && character != '-' {
			return false
		}
	}
	return true
}
