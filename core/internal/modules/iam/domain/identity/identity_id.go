package identity

import (
	"encoding/hex"
	"errors"
)

var ErrInvalidIdentityID = errors.New("IAM identity ID is invalid")

type IdentityID [16]byte

func ParseIdentityID(value string) (IdentityID, error) {
	var resolved IdentityID
	if len(value) != 36 || value[8] != '-' || value[13] != '-' ||
		value[18] != '-' || value[23] != '-' {
		return resolved, ErrInvalidIdentityID
	}
	compact := make([]byte, 0, 32)
	for index := range value {
		if index == 8 || index == 13 || index == 18 || index == 23 {
			continue
		}
		if !identityHex(value[index]) {
			return IdentityID{}, ErrInvalidIdentityID
		}
		compact = append(compact, value[index])
	}
	if _, err := hex.Decode(resolved[:], compact); err != nil || resolved == (IdentityID{}) {
		return IdentityID{}, ErrInvalidIdentityID
	}
	return resolved, nil
}

func identityHex(value byte) bool {
	return value >= '0' && value <= '9' || value >= 'a' && value <= 'f'
}
