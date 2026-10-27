package authentication

type IssuedRefresh interface {
	Selector() ([16]byte, error)
	SecretHMAC([]byte) ([32]byte, error)
	RevealOnce(func([]byte) error) error
	Destroy()
}

type PresentedRefresh interface {
	Selector() ([16]byte, error)
	VerifyHMAC([]byte, [32]byte) (bool, error)
	Destroy()
}
