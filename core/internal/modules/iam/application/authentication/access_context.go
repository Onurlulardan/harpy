package authentication

func (authenticator Authenticator) signAccess(
	input AccessClaims,
) (string, error) {
	wire, err := authenticator.access.Issue(input)
	if err != nil || wire == "" {
		return "", ErrAuthenticationCrypto
	}
	return wire, nil
}
