package main

import "testing"

func TestAPIEnvironmentAllowsBothLocalBrowserOrigins(t *testing.T) {
	set, _ := generatedSetForTest(t)
	environment := parseProfileEnvironment(t, set.files["api.env"])
	const expected = `["http://localhost:5173","http://127.0.0.1:5173"]`
	if environment["HARPY_IAM_ALLOWED_ORIGINS"] != expected {
		t.Fatalf("local allowed origins = %v, want %s",
			environment["HARPY_IAM_ALLOWED_ORIGINS"], expected)
	}
}
