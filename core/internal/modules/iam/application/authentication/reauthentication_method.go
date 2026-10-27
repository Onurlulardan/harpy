package authentication

type ReauthenticationMethod string

const (
	ReauthenticationPassword ReauthenticationMethod = "password_v1"
	ReauthenticationWebAuthn ReauthenticationMethod = "webauthn_v1"
	ReauthenticationTOTP     ReauthenticationMethod = "totp_v1"
)

func (method ReauthenticationMethod) valid() bool {
	return method == ReauthenticationPassword ||
		method == ReauthenticationWebAuthn || method == ReauthenticationTOTP
}

func (method ReauthenticationMethod) accessMethod() AuthMethod {
	switch method {
	case ReauthenticationPassword:
		return AuthMethodPassword
	case ReauthenticationWebAuthn:
		return AuthMethodWebAuthn
	case ReauthenticationTOTP:
		return AuthMethodTOTP
	default:
		return ""
	}
}
