package fcartv1

type DEK [DEKBytes]byte

func NewDEK(raw []byte) (DEK, error) {
	var key DEK
	if len(raw) != DEKBytes {
		return key, ErrInvalidDEK
	}
	copy(key[:], raw)
	if err := key.Validate(); err != nil {
		return DEK{}, err
	}
	return key, nil
}

func (key DEK) Validate() error {
	var zero DEK
	if key == zero {
		return ErrInvalidDEK
	}
	return nil
}
