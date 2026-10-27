package iam

import "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"

func validReauthenticationCompleteInput(
	input AuthHTTPReauthenticationCompleteInput,
) bool {
	if input.ChallengeID == [16]byte{} || len(input.Challenge) != 43 ||
		input.PurposeSHA256 == [32]byte{} || input.Origin == "" ||
		len(input.Origin) > 512 {
		return false
	}
	proofs := 0
	for _, proof := range [][]byte{input.Password, input.WebAuthnAssertion, input.TOTP} {
		if len(proof) != 0 {
			proofs++
		}
	}
	if proofs != 1 {
		return false
	}
	switch authentication.ReauthenticationMethod(input.Method) {
	case authentication.ReauthenticationPassword:
		return len(input.Password) <= 512
	case authentication.ReauthenticationWebAuthn:
		return len(input.WebAuthnAssertion) <= 64*1024
	case authentication.ReauthenticationTOTP:
		return len(input.TOTP) == 6
	default:
		return false
	}
}

func clearReauthenticationCompleteInput(input *AuthHTTPReauthenticationCompleteInput) {
	if input == nil {
		return
	}
	clear(input.Challenge)
	clear(input.Password)
	clear(input.WebAuthnAssertion)
	clear(input.TOTP)
}
