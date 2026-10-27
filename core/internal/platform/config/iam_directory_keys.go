package config

import "sort"

// IAMDirectoryKey is a redacted, defensive view of one directory HMAC key.
// Raw material is available only as a value copy for explicit composition.
type IAMDirectoryKey struct {
	iamKeyRedaction
	id       string
	material [32]byte
	active   bool
}

func (key IAMDirectoryKey) ID() (string, error) {
	if err := key.validate(); err != nil {
		return "", err
	}
	return key.id, nil
}

func (key IAMDirectoryKey) Material() ([32]byte, error) {
	if err := key.validate(); err != nil {
		return [32]byte{}, err
	}
	return key.material, nil
}

func (key IAMDirectoryKey) Active() (bool, error) {
	if err := key.validate(); err != nil {
		return false, err
	}
	return key.active, nil
}

func (key *IAMDirectoryKey) Destroy() {
	if key == nil {
		return
	}
	clear(key.material[:])
	key.id = ""
	key.active = false
}

func (key IAMDirectoryKey) validate() error {
	if !validIAMHMACKeyID(key.id) || iamAllZero(key.material[:]) {
		return errInvalidIAMHMACKeyring
	}
	return nil
}

// Keys returns active first, followed by retained keys in canonical ID order.
func (keyring IAMDirectoryKeyring) Keys() ([]IAMDirectoryKey, error) {
	if err := keyring.Validate(); err != nil || keyring.purpose != iamPurposeDirectory {
		return nil, errInvalidIAMHMACKeyring
	}
	ids := make([]string, 0, len(keyring.retained))
	for id := range keyring.retained {
		if id == keyring.active.kid {
			return nil, errInvalidIAMHMACKeyring
		}
		ids = append(ids, id)
	}
	sort.Strings(ids)
	keys := make([]IAMDirectoryKey, 0, len(ids)+1)
	keys = append(keys, IAMDirectoryKey{
		id: keyring.active.kid, material: keyring.active.material, active: true,
	})
	for _, id := range ids {
		keys = append(keys, IAMDirectoryKey{id: id, material: keyring.retained[id]})
	}
	return keys, nil
}
