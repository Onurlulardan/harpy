package main

import "testing"

func TestAPIEnvironmentAllowsBothLocalBrowserOrigins(t *testing.T) {
	set, _ := generatedSetForTest(t)
	environment := parseProfileEnvironment(t, set.files["api.env"])
	const expected = localAllowedOrigins
	if environment["HARPY_IAM_ALLOWED_ORIGINS"] != expected {
		t.Fatalf("local allowed origins = %v, want %s",
			environment["HARPY_IAM_ALLOWED_ORIGINS"], expected)
	}
}
