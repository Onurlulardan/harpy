package publication

type ID struct{ value [16]byte }

func NewID(value [16]byte) (ID, error) {
	if !validV7(value) {
		return ID{}, ErrInvalid
	}
	return ID{value: value}, nil
}

func (id ID) Valid() bool     { return validV7(id.value) }
func (id ID) Bytes() [16]byte { return id.value }

func validV7(value [16]byte) bool {
	var zero [16]byte
	return value != zero && value[6]>>4 == 7 && value[8]>>6 == 2
}

type PublicID struct{ value [16]byte }

func NewPublicID(value [16]byte) (PublicID, error) {
	var zero [16]byte
	if value == zero {
		return PublicID{}, ErrInvalid
	}
	return PublicID{value: value}, nil
}

func (id PublicID) Valid() bool     { var zero [16]byte; return id.value != zero }
func (id PublicID) Bytes() [16]byte { return id.value }
