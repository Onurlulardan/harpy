package cryptography

import "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/authentication"

type TOTPCode struct{ secureMaterial }

func NewTOTPCode(value []byte) (*TOTPCode, error) {
	if len(value) != 6 {
		return nil, authentication.ErrAuthenticationContract
	}
	for _, digit := range value {
		if digit < '0' || digit > '9' {
			return nil, authentication.ErrAuthenticationContract
		}
	}
	material, err := newSecureMaterial(value, 6, 6)
	if err != nil {
		return nil, authentication.ErrAuthenticationContract
	}
	return &TOTPCode{secureMaterial: material}, nil
}

func (code *TOTPCode) Digits() ([6]byte, error) {
	value, err := code.copy()
	if err != nil {
		return [6]byte{}, authentication.ErrAuthenticationContract
	}
	defer clear(value)
	var result [6]byte
	copy(result[:], value)
	return result, nil
}
