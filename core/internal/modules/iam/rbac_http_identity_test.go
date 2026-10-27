package iam

import (
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/accesstoken"
	"github.com/google/uuid"
)

func TestRBACStepUpRequiresExactFreshPurposeAndStrongMethod(t *testing.T) {
	now := time.Unix(2_000_000_000, 0).UTC()
	profile, err := accesstoken.NewProfile(
		"http://localhost:8000/auth", "http://localhost:8000/api", true,
	)
	if err != nil {
		t.Fatal(err)
	}
	command := "iam.rbac.assignment.update"
	target := uuid.New()
	purpose := rbacStepUpPurpose(command, target)
	claims := newRBACStepUpClaims(t, profile, now, purpose[:], now.Add(-time.Minute))
	application := &rbacHTTPApplication{now: func() time.Time { return now }}
	if evidence, valid := application.stepUp(claims, command, target); !valid || evidence.PurposeSHA256 != purpose || !evidence.AMRPassword {
		t.Fatalf("valid step-up rejected: %#v / %t", evidence, valid)
	}
	wrongTarget := uuid.New()
	if _, valid := application.stepUp(claims, command, wrongTarget); valid {
		t.Fatal("step-up accepted a purpose bound to another aggregate")
	}
	stale := newRBACStepUpClaims(t, profile, now, purpose[:], now.Add(-11*time.Minute))
	if _, valid := application.stepUp(stale, command, target); valid {
		t.Fatal("step-up accepted stale authentication evidence")
	}
}

func TestRBACCreatePurposeDoesNotBindServerGeneratedID(t *testing.T) {
	command := "iam.rbac.assignment.create"
	if rbacStepUpPurpose(command, uuid.New()) != rbacStepUpPurpose(command, uuid.Nil) {
		t.Fatal("create purpose unexpectedly depends on a server-generated ID")
	}
}

func newRBACStepUpClaims(
	t *testing.T,
	profile accesstoken.Profile,
	now time.Time,
	purpose []byte,
	authTime time.Time,
) accesstoken.Claims {
	t.Helper()
	claims, err := accesstoken.NewClaims(profile, accesstoken.ClaimsInput{
		Subject: uuid.New(), Session: uuid.New(), Tenant: uuid.New(),
		Client: accesstoken.ClientOperatorConsole, IssuedAt: now, JTI: uuid.New(),
		TokenVersion: 3, AuthorizationVersion: 5,
		PlatformAuthorizationClock: 7, SecurityEpoch: [32]byte{1},
		AccessMode: accesstoken.AccessModeOperator, AuthenticationTime: authTime,
		Assurance:         accesstoken.AssuranceStepUp,
		Methods:           []accesstoken.AuthMethod{accesstoken.AuthMethodPassword},
		StepUpPurposeHash: purpose,
	})
	if err != nil {
		t.Fatal(err)
	}
	return claims
}
