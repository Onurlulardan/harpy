package config

import "errors"

var errInvalidIAMHMACKeyring = errors.New("IAM HMAC keyring is invalid")

type iamHMACKey struct {
	iamKeyRedaction
	kid      string
	material [32]byte
}

type iamHMACKeyring struct {
	iamKeyRedaction
	purpose  iamHMACPurpose
	active   iamHMACKey
	retained map[string][32]byte
	valid    bool
}

func (keyring iamHMACKeyring) Validate() error {
	if !keyring.valid || !keyring.purpose.valid() ||
		!validIAMHMACKeyID(keyring.active.kid) || iamAllZero(keyring.active.material[:]) ||
		len(keyring.retained) > 64 {
		return errInvalidIAMHMACKeyring
	}
	for kid, material := range keyring.retained {
		if !validIAMHMACKeyID(kid) || iamAllZero(material[:]) {
			return errInvalidIAMHMACKeyring
		}
	}
	return nil
}

func (keyring iamHMACKeyring) Purpose() (string, error) {
	if err := keyring.Validate(); err != nil {
		return "", err
	}
	return string(keyring.purpose), nil
}

func (keyring iamHMACKeyring) Active() (string, [32]byte, error) {
	if err := keyring.Validate(); err != nil {
		return "", [32]byte{}, err
	}
	return keyring.active.kid, keyring.active.material, nil
}

func (keyring iamHMACKeyring) HMACKey(kid string) ([32]byte, error) {
	if err := keyring.Validate(); err != nil || !validIAMHMACKeyID(kid) {
		return [32]byte{}, errInvalidIAMHMACKeyring
	}
	if kid == keyring.active.kid {
		return keyring.active.material, nil
	}
	material, exists := keyring.retained[kid]
	if !exists {
		return [32]byte{}, errInvalidIAMHMACKeyring
	}
	return material, nil
}

func (keyring iamHMACKeyring) keyMaterials() map[string][32]byte {
	materials := make(map[string][32]byte, len(keyring.retained)+1)
	materials[keyring.active.kid] = keyring.active.material
	for kid, material := range keyring.retained {
		materials[kid] = material
	}
	return materials
}

type IAMRefreshKeyring struct{ iamHMACKeyring }
type IAMCSRFKeyring struct{ iamHMACKeyring }
type IAMDirectoryKeyring struct{ iamHMACKeyring }
type IAMThrottleKeyring struct{ iamHMACKeyring }
type IAMSessionNetworkKeyring struct{ iamHMACKeyring }
type IAMOneTimeKeyring struct{ iamHMACKeyring }
type IAMRecoveryKeyring struct{ iamHMACKeyring }
