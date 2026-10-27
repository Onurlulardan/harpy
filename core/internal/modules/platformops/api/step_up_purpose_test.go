package api

import (
	"encoding/base64"
	"testing"
)

func TestStepUpPurposeHasStableBrowserVector(t *testing.T) {
	tenantID := "11111111-1111-4111-8111-111111111111"
	targetID := "22222222-2222-4222-8222-222222222222"
	digest, err := StepUpPurpose(
		tenantID, "platform.recovery.enable", targetID)
	if err != nil {
		t.Fatal(err)
	}
	encoded := base64.RawURLEncoding.EncodeToString(digest[:])
	const expected = "wtK89l_CJ4HBUsIGY3_3F-77uD6k8B6NBdLlFSJEl4M"
	if encoded != expected {
		t.Fatalf("step-up vector drifted: %s", encoded)
	}
}

func TestStepUpPurposeRejectsUnknownCapability(t *testing.T) {
	_, err := StepUpPurpose(
		"11111111-1111-4111-8111-111111111111",
		"platform.recovery.read",
		"22222222-2222-4222-8222-222222222222",
	)
	if err == nil {
		t.Fatal("non-step-up recovery capability was accepted")
	}
}
