package cryptography

import "github.com/go-webauthn/webauthn/protocol"

func webAuthnDiscoverable(
	outputs protocol.AuthenticationExtensionsClientOutputs,
) (bool, error) {
	wire, found := outputs["credProps"]
	if !found {
		return false, nil
	}
	properties, valid := wire.(map[string]any)
	if !valid {
		return false, ErrRegistrationDenied
	}
	resident, found := properties["rk"]
	if !found {
		return false, nil
	}
	value, valid := resident.(bool)
	if !valid {
		return false, ErrRegistrationDenied
	}
	return value, nil
}
