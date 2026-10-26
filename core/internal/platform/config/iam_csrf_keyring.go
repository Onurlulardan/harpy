package config

import "context"

func (keyring IAMCSRFKeyring) CSRFKey(
	ctx context.Context,
	kid string,
) ([]byte, error) {
	if ctx == nil || ctx.Err() != nil {
		return nil, errInvalidIAMHMACKeyring
	}
	material, err := keyring.HMACKey(kid)
	if err != nil {
		return nil, err
	}
	return append([]byte(nil), material[:]...), nil
}
