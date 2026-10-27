package api

import (
	"context"
	"testing"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	workmodule "github.com/Onurlulardan/harpy/core/internal/modules/work"
	"github.com/google/uuid"
)

type workIAMAccessFixture struct {
	decision iammodule.WorkAccessDecision
	err      error
}

func (fixture *workIAMAccessFixture) Verify(
	context.Context, iammodule.WorkAccessRequest,
) (iammodule.WorkAccessDecision, error) {
	return fixture.decision, fixture.err
}

func TestWorkAccessVerifierAcceptsJWTWithoutCohortHeaders(t *testing.T) {
	decision, err := iammodule.NewWorkAccessDecision(uuid.New(), uuid.New(), 1, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	verifier, err := newWorkAccessVerifier(&workIAMAccessFixture{decision: decision})
	if err != nil {
		t.Fatal(err)
	}
	access, err := verifier.Verify(context.Background(), workmodule.HTTPAccessRequest{
		Authority: "work.calendar.list", Operation: "listLegacyCalendarEvents",
		Compatibility: true,
	})
	if err != nil || access.PrincipalID() == uuid.Nil || access.ConsumerID() != "" {
		t.Fatalf("jwt compatibility access = %#v err=%v", access, err)
	}
}

func TestWorkAccessVerifierAcceptsCompatibilityEvidence(t *testing.T) {
	decision, err := iammodule.NewWorkAccessDecision(uuid.New(), uuid.New(), 1, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	verifier, err := newWorkAccessVerifier(&workIAMAccessFixture{decision: decision})
	if err != nil {
		t.Fatal(err)
	}
	access, err := verifier.Verify(context.Background(), workmodule.HTTPAccessRequest{
		Authority: "work.calendar.list", Operation: "listLegacyCalendarEvents",
		ConsumerIDValues: []string{"frontend-v3"}, BuildIDValues: []string{"frontend-v3-development"},
		Compatibility: true,
	})
	if err != nil || access.ConsumerID() != "frontend-v3" ||
		access.BuildID() != "frontend-v3-development" {
		t.Fatalf("compatibility access = %#v err=%v", access, err)
	}
}
