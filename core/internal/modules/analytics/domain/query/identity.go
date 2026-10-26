package query

import "bytes"

type ID struct{ value [16]byte }

func NewID(value [16]byte) (ID, error) {
	if !validV7(value) {
		return ID{}, ErrInvalid
	}
	return ID{value: value}, nil
}

func (id ID) Valid() bool         { return validV7(id.value) }
func (id ID) Bytes() [16]byte     { return id.value }
func (id ID) Equal(other ID) bool { return bytes.Equal(id.value[:], other.value[:]) }

type VersionID struct{ value [16]byte }

func NewVersionID(value [16]byte) (VersionID, error) {
	if !validV7(value) {
		return VersionID{}, ErrInvalid
	}
	return VersionID{value: value}, nil
}

func (id VersionID) Valid() bool     { return validV7(id.value) }
func (id VersionID) Bytes() [16]byte { return id.value }

func validV7(value [16]byte) bool {
	var zero [16]byte
	return value != zero && value[6]>>4 == 7 && value[8]>>6 == 2
}
