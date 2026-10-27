package platformiam

import "time"

const StepUpACR = "urn:harpy:acr:step-up"

type Evidence struct {
	OperationID                  ID
	RequestSHA256                [32]byte
	ActorID                      ID
	ExpectedTokenVersion         int64
	ExpectedAuthorizationVersion int64
	ExpectedPlatformClock        int64
}

func (evidence Evidence) Validate() error {
	if !evidence.OperationID.Valid() || evidence.RequestSHA256 == [32]byte{} ||
		!evidence.ActorID.Valid() || evidence.ExpectedTokenVersion < 1 ||
		evidence.ExpectedAuthorizationVersion < 1 || evidence.ExpectedPlatformClock < 1 {
		return ErrInvalidCommand
	}
	return nil
}

type StepUpEvidence struct {
	PurposeSHA256 [32]byte
	AuthTime      time.Time
	ACR           string
	AMRPassword   bool
	AMRWebAuthn   bool
	AMRTOTP       bool
}

func (evidence StepUpEvidence) Validate(now time.Time, expected [32]byte) error {
	if evidence.PurposeSHA256 == [32]byte{} || evidence.PurposeSHA256 != expected ||
		evidence.AuthTime.IsZero() || evidence.AuthTime.After(now) ||
		now.Sub(evidence.AuthTime) > 10*time.Minute || evidence.ACR != StepUpACR ||
		!(evidence.AMRPassword || evidence.AMRWebAuthn || evidence.AMRTOTP) {
		return ErrInvalidCommand
	}
	return nil
}
