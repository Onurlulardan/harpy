package authentication

type refreshAccessContext struct {
	assurance Assurance
	methods   []AuthMethod
	stepUp    []byte
}

func (credential RefreshCredential) accessContext() (refreshAccessContext, error) {
	if credential.AMRRecovery {
		return refreshAccessContext{}, ErrAuthenticationStorage
	}
	methods := make([]AuthMethod, 0, 3)
	if credential.AMRPassword {
		methods = append(methods, AuthMethodPassword)
	}
	if credential.AMRWebAuthn {
		methods = append(methods, AuthMethodWebAuthn)
	}
	if credential.AMRTOTP {
		methods = append(methods, AuthMethodTOTP)
	}
	if len(methods) == 0 {
		return refreshAccessContext{}, ErrAuthenticationStorage
	}
	result := refreshAccessContext{methods: methods}
	switch credential.ACR {
	case "login", "mfa":
		if credential.HasStepUpPurpose {
			return refreshAccessContext{}, ErrAuthenticationStorage
		}
		result.assurance = AssuranceBaseline
	case "step_up":
		if !credential.HasStepUpPurpose {
			return refreshAccessContext{}, ErrAuthenticationStorage
		}
		result.assurance = AssuranceStepUp
		result.stepUp = append([]byte(nil), credential.StepUpPurposeSHA256[:]...)
	default:
		return refreshAccessContext{}, ErrAuthenticationStorage
	}
	return result, nil
}
