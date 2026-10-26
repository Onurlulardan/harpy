package api

import (
	"encoding/base32"
	"encoding/hex"
)

type UUID = [16]byte

var NilUUID UUID

var onboardingSchemaEncoding = base32.NewEncoding(
	"abcdefghijklmnopqrstuvwxyz234567",
).WithPadding(base32.NoPadding)

func UUIDString(value UUID) string {
	encoded := make([]byte, 32)
	hex.Encode(encoded, value[:])
	return string(encoded[:8]) + "-" + string(encoded[8:12]) + "-" +
		string(encoded[12:16]) + "-" + string(encoded[16:20]) + "-" + string(encoded[20:])
}

func onboardingSchemaName(value UUID) string {
	if value == NilUUID {
		return ""
	}
	return "t_" + onboardingSchemaEncoding.EncodeToString(value[:])
}

func OnboardingSchemaName(value UUID) (string, error) {
	name := onboardingSchemaName(value)
	if name == "" {
		return "", ErrOnboardingInvalid
	}
	return name, nil
}
