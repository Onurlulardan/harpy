package presentation

import (
	"encoding/hex"
	"strings"
)

func stableUUIDText(value string) bool {
	decoded, valid := decodeUUID(value)
	if !valid || decoded[8]&0xc0 != 0x80 {
		return false
	}
	for _, octet := range decoded {
		if octet != 0 {
			return true
		}
	}
	return false
}

func stableUUIDv7(value string) bool {
	decoded, valid := decodeUUID(value)
	return valid && decoded[6]>>4 == 7 && decoded[8]&0xc0 == 0x80
}

func decodeUUID(value string) ([16]byte, bool) {
	var result [16]byte
	if len(value) != 36 || value != strings.ToLower(value) || value[8] != '-' ||
		value[13] != '-' || value[18] != '-' || value[23] != '-' {
		return result, false
	}
	compact := strings.ReplaceAll(value, "-", "")
	decoded, err := hex.DecodeString(compact)
	if err != nil || len(decoded) != len(result) {
		return result, false
	}
	copy(result[:], decoded)
	return result, true
}
