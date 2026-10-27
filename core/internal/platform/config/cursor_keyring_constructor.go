package config

func NewCursorKey(kid string, material [cursorKeySize]byte) (CursorKey, error) {
	key := CursorKey{kid: kid, material: material, valid: true}
	if err := key.validate(); err != nil {
		return CursorKey{}, err
	}
	return key, nil
}

func NewCursorKeyring(current, previous CursorKey) (CursorKeyring, error) {
	keyring := CursorKeyring{current: current, previous: previous, valid: true}
	if err := keyring.Validate(); err != nil {
		return CursorKeyring{}, err
	}
	return keyring, nil
}
