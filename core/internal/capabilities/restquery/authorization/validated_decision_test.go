package authorization

import (
	"crypto/sha256"
	"testing"
)

func TestPolicyDecisionMustExactlyCoverRequest(t *testing.T) {
	request := mustPolicyRequest(t)
	var fieldID [sha256.Size]byte
	for _, requirement := range request.requirements {
		if requirement.kind == RequirementFieldFilter {
			fieldID = requirement.id
		}
	}
	decision := exactDecision(t, request, map[[sha256.Size]byte]bool{fieldID: true})
	validated, err := validatePolicyDecision(request, decision)
	if err != nil || len(validated.denied) != 1 ||
		validated.denied[0].kind != RequirementFieldFilter ||
		len(validated.visibilities) != 1 ||
		validated.contextRevision == [sha256.Size]byte{} {
		t.Fatalf("valid exact decision was rejected: %v", err)
	}

	resourceDenied := make(map[[sha256.Size]byte]bool)
	for _, requirement := range request.requirements {
		if requirement.kind == RequirementResourceRead {
			resourceDenied[requirement.id] = true
		}
	}
	decision = exactDecision(t, request, resourceDenied)
	validated, err = validatePolicyDecision(request, decision)
	if err != nil || len(validated.denied) != 1 || len(validated.visibilities) != 0 {
		t.Fatalf("resource denial contract: %v", err)
	}
}

func TestPolicyDecisionRejectsBindingResultAndVisibilityMismatch(t *testing.T) {
	request := mustPolicyRequest(t)
	exact := exactDecision(t, request, nil)
	wrongBinding, _ := NewPolicyDecision(
		sha256.Sum256([]byte("wrong")), exact.contextRevision,
		exact.results, exact.visibilities,
	)
	if _, err := validatePolicyDecision(request, wrongBinding); err == nil {
		t.Fatal("wrong policy request binding was accepted")
	}

	missingResult, _ := NewPolicyDecision(
		exact.requestBinding, exact.contextRevision,
		exact.results[:len(exact.results)-1], exact.visibilities,
	)
	if _, err := validatePolicyDecision(request, missingResult); err == nil {
		t.Fatal("partial policy result set was accepted")
	}

	missingVisibility, _ := NewPolicyDecision(
		exact.requestBinding, exact.contextRevision, exact.results, nil,
	)
	if _, err := validatePolicyDecision(request, missingVisibility); err == nil {
		t.Fatal("missing allowed-resource visibility was accepted")
	}

	var fieldResult DecisionResult
	for _, result := range exact.results {
		for _, requirement := range request.requirements {
			if result.requirement == requirement.id &&
				requirement.kind == RequirementFieldFilter {
				fieldResult = result
			}
		}
	}
	wrongVisibility, _ := NewAllVisibility(fieldResult.requirement)
	wrongTarget, _ := NewPolicyDecision(
		exact.requestBinding, exact.contextRevision, exact.results,
		[]Visibility{wrongVisibility},
	)
	if _, err := validatePolicyDecision(request, wrongTarget); err == nil {
		t.Fatal("non-resource visibility target was accepted")
	}
}
