package accesstoken

import (
	"bytes"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestClaimsRejectInvalidSecurityShapes(t *testing.T) {
	tests := map[string]func(*ClaimsInput){
		"nil subject": func(input *ClaimsInput) { input.Subject = uuid.Nil },
		"non-v4 jti": func(input *ClaimsInput) {
			input.JTI = uuid.MustParse("40000000-0000-1000-8000-000000000001")
		},
		"future auth":                 func(input *ClaimsInput) { input.AuthenticationTime = input.IssuedAt.Add(time.Second) },
		"zero token version":          func(input *ClaimsInput) { input.TokenVersion = 0 },
		"negative platform clock":     func(input *ClaimsInput) { input.PlatformAuthorizationClock = -1 },
		"zero security epoch":         func(input *ClaimsInput) { input.SecurityEpoch = [32]byte{} },
		"portal client operator mode": func(input *ClaimsInput) { input.Client = ClientPortal },
		"operator portal binding":     func(input *ClaimsInput) { input.PortalBinding = &input.Session },
		"unordered amr": func(input *ClaimsInput) {
			input.Methods = []AuthMethod{AuthMethodTOTP, AuthMethodPassword}
		},
		"baseline purpose":        func(input *ClaimsInput) { input.StepUpPurposeHash = make([]byte, 32) },
		"step-up without purpose": func(input *ClaimsInput) { input.Assurance = AssuranceStepUp },
		"subsecond iat":           func(input *ClaimsInput) { input.IssuedAt = input.IssuedAt.Add(time.Nanosecond) },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			fixture := newTokenFixture(t)
			mutate(&fixture.input)
			if _, err := NewClaims(fixture.profile, fixture.input); err != ErrInvalidClaims {
				t.Fatalf("expected invalid claims, got %v", err)
			}
		})
	}
}

func TestGenerateJTIUsesProvidedRandomness(t *testing.T) {
	identifier, err := GenerateJTI(bytes.NewReader(bytes.Repeat([]byte{1}, 16)))
	if err != nil {
		t.Fatal(err)
	}
	if identifier.Version() != 4 || identifier.Variant() != uuid.RFC4122 {
		t.Fatalf("unexpected JTI shape: %s", identifier)
	}
	if _, err := GenerateJTI(bytes.NewReader(nil)); err != ErrInvalidClaims {
		t.Fatalf("expected randomness failure, got %v", err)
	}
}

func TestClaimsAccessorsReturnCopies(t *testing.T) {
	fixture := newTokenFixture(t)
	fixture.input.Methods = []AuthMethod{AuthMethodPassword, AuthMethodTOTP}
	claims := fixture.claims(t)
	methods := claims.Methods()
	methods[0] = AuthMethodWebAuthn
	if claims.Methods()[0] != AuthMethodPassword {
		t.Fatal("methods accessor leaked mutable storage")
	}
}
