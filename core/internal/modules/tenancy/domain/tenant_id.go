package domain

import (
	"encoding/hex"
	"errors"
)

var ErrInvalidTenantID = errors.New("tenant ID must be a non-zero canonical UUID")

type TenantID [16]byte

func ParseTenantID(value string) (TenantID, error) {
	var resolved TenantID
	if len(value) != 36 || value[8] != '-' || value[13] != '-' ||
		value[18] != '-' || value[23] != '-' {
		return resolved, ErrInvalidTenantID
	}
	compact := make([]byte, 0, 32)
	for index := range value {
		if index == 8 || index == 13 || index == 18 || index == 23 {
			continue
		}
		if !lowerHex(value[index]) {
			return resolved, ErrInvalidTenantID
		}
		compact = append(compact, value[index])
	}
	if _, err := hex.Decode(resolved[:], compact); err != nil || resolved == (TenantID{}) {
		return TenantID{}, ErrInvalidTenantID
	}
	return resolved, nil
}

func lowerHex(value byte) bool {
	return value >= '0' && value <= '9' || value >= 'a' && value <= 'f'
}
