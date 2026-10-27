package config

import (
	"strings"
	"testing"
	"time"
)

func TestIAMHTTPProfileLoadsExactPublishedValues(t *testing.T) {
	loaded, err := loadAPI(validAPIEnvironment())
	if err != nil {
		t.Fatal(err)
	}
	iamConfig, err := loaded.IAM()
	if err != nil {
		t.Fatal(err)
	}
	profile, err := iamConfig.HTTPProfile()
	if err != nil {
		t.Fatal(err)
	}
	if profile.Issuer() != "https://api.harpy.invalid" ||
		profile.Audience() != "https://api.harpy.invalid/api" ||
		!profile.AllowLocalHTTP() ||
		len(profile.AllowedOrigins()) != 1 ||
		profile.AllowedOrigins()[0] != "https://app.harpy.invalid" {
		t.Fatalf("unexpected IAM HTTP profile: %#v", profile)
	}
	origins := profile.AllowedOrigins()
	origins[0] = "https://mutated.invalid"
	if profile.AllowedOrigins()[0] != "https://app.harpy.invalid" {
		t.Fatal("allowed origins leaked mutable state")
	}
}

func TestIAMHTTPProfileAllowsLoopbackOnlyInExplicitLocalProfiles(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	local := withEnvironment(
		validAPIEnvironmentAt(now),
		iamAccessIssuerEnvironmentKey+"=http://localhost:8000",
		iamAccessAudienceEnvironmentKey+"=http://127.0.0.1:8000/api",
		iamAllowedOriginsEnvironmentKey+`=["http://[::1]:5173"]`,
	)
	loaded, err := loadAPIAt(local, now)
	if err != nil {
		t.Fatal(err)
	}
	iamConfig, _ := loaded.IAM()
	profile, _ := iamConfig.HTTPProfile()
	if !profile.AllowLocalHTTP() {
		t.Fatal("development did not explicitly enable loopback HTTP")
	}
	production := withEnvironment(local, "HARPY_ENVIRONMENT=production")
	if _, err := loadAPIAt(production, now); err == nil {
		t.Fatal("production accepted loopback HTTP profile")
	}
}

func TestIAMHTTPProfileRejectsMalformedOrAmbiguousValues(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	tests := []struct {
		name  string
		entry string
		key   string
	}{
		{"issuer userinfo", iamAccessIssuerEnvironmentKey + "=https://user@example.invalid", iamAccessIssuerEnvironmentKey},
		{"audience query", iamAccessAudienceEnvironmentKey + "=https://api.example.invalid?", iamAccessAudienceEnvironmentKey},
		{"nonloopback http", iamAccessAudienceEnvironmentKey + "=http://example.invalid/api", iamAccessAudienceEnvironmentKey},
		{"origin path", iamAllowedOriginsEnvironmentKey + `=["https://app.example.invalid/"]`, iamAllowedOriginsEnvironmentKey},
		{"duplicate origin", iamAllowedOriginsEnvironmentKey + `=["https://app.example.invalid","https://app.example.invalid"]`, iamAllowedOriginsEnvironmentKey},
		{"noncanonical json", iamAllowedOriginsEnvironmentKey + `=["https://app.example.invalid" ]`, iamAllowedOriginsEnvironmentKey},
		{"null origins", iamAllowedOriginsEnvironmentKey + "=null", iamAllowedOriginsEnvironmentKey},
		{"trailing document", iamAllowedOriginsEnvironmentKey + `=["https://app.example.invalid"]{}`, iamAllowedOriginsEnvironmentKey},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := loadAPIAt(withEnvironment(validAPIEnvironmentAt(now), test.entry), now)
			if err == nil || !strings.Contains(err.Error(), test.key) {
				t.Fatalf("expected stable %s error, got %v", test.key, err)
			}
		})
	}
}
