package permissionrequest

import "time"

const StepUpACR = "urn:harpy:acr:step-up"

type StepUpEvidence struct {
	PurposeSHA256 [32]byte
	AuthTime      time.Time
	ACR           string
	AMRPassword   bool
	AMRWebAuthn   bool
	AMRTOTP       bool
}

func (evidence StepUpEvidence) Validate() error {
	if evidence.PurposeSHA256 == [32]byte{} || evidence.AuthTime.IsZero() ||
		evidence.ACR != StepUpACR ||
		!(evidence.AMRPassword || evidence.AMRWebAuthn || evidence.AMRTOTP) {
		return ErrInvalid
	}
	return nil
}
