package authentication

type InitialPasswordKeyring interface {
	Active() (string, [32]byte, error)
	HMACKey(string) ([32]byte, error)
}
