package authentication

type IssuedChallenge interface {
	HMAC([]byte) ([32]byte, error)
	RevealOnce(func([]byte) error) error
	Destroy()
}

type PresentedChallenge interface {
	HMAC([]byte) ([32]byte, error)
	Wire() ([]byte, error)
	Destroy()
}

type WebAuthnAssertion interface {
	Bytes() ([]byte, error)
	Destroy()
}

type TOTPCode interface {
	Digits() ([6]byte, error)
	Destroy()
}

type RecoveryCode interface {
	KeyID() (string, error)
	HMAC([]byte) ([32]byte, error)
	Destroy()
}
