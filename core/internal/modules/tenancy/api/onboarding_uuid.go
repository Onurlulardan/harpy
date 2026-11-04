package api

import "encoding/hex"

type UUID = [16]byte

var NilUUID UUID

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
	return "tenant_" + hex.EncodeToString(value[:])
}

func OnboardingSchemaName(value UUID) (string, error) {
	name := onboardingSchemaName(value)
	if name == "" {
		return "", ErrOnboardingInvalid
	}
	return name, nil
}
