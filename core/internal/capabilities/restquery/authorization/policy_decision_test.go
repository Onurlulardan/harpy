package authorization

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"testing"
)

func TestPolicyDecisionCanonicalizesAndDefensivelyCopies(t *testing.T) {
	firstID := sha256.Sum256([]byte("first"))
	secondID := sha256.Sum256([]byte("second"))
	first, _ := NewDecisionResult(firstID, true)
	second, _ := NewDecisionResult(secondID, false)
	visibility, _ := NewAllVisibility(firstID)
	decision, err := NewPolicyDecision(
		sha256.Sum256([]byte("request")),
		sha256.Sum256([]byte("revision")),
		[]DecisionResult{second, first},
		[]Visibility{visibility},
	)
	if err != nil {
		t.Fatal(err)
	}
	binding, bindingErr := decision.RequestBindingSHA256()
	revision, revisionErr := decision.ContextRevisionSHA256()
	results, resultsErr := decision.Results()
	visibilities, visibilityErr := decision.Visibilities()
	if bindingErr != nil || binding == [32]byte{} || revisionErr != nil ||
		revision == [32]byte{} || resultsErr != nil || len(results) != 2 || visibilityErr != nil ||
		len(visibilities) != 1 {
		t.Fatal("policy decision accessors are invalid")
	}
	left, _ := results[0].RequirementSHA256()
	right, _ := results[1].RequirementSHA256()
	if bytes.Compare(left[:], right[:]) >= 0 {
		t.Fatal("policy decision results are not canonical")
	}
	results[0] = DecisionResult{}
	visibilities[0] = Visibility{}
	if decision.Validate() != nil {
		t.Fatal("returned collections mutated policy decision")
	}
	if fmt.Sprintf("%v", decision) != authorizationRedaction ||
		decision.LogValue().String() != authorizationRedaction {
		t.Fatal("policy decision was not redacted")
	}
	if _, err := json.Marshal(decision); err == nil {
		t.Fatal("policy decision allowed direct JSON marshal")
	}
}

func TestPolicyDecisionRejectsDuplicatesAndTamper(t *testing.T) {
	if _, err := NewDecisionResult([32]byte{}, true); err == nil {
		t.Fatal("zero decision requirement was accepted")
	}
	id := sha256.Sum256([]byte("requirement"))
	result, _ := NewDecisionResult(id, true)
	if _, err := NewPolicyDecision(
		sha256.Sum256([]byte("request")), sha256.Sum256([]byte("revision")),
		[]DecisionResult{result, result}, nil,
	); err == nil {
		t.Fatal("duplicate decision result was accepted")
	}
	decision, _ := NewPolicyDecision(
		sha256.Sum256([]byte("request")), sha256.Sum256([]byte("revision")),
		[]DecisionResult{result}, nil,
	)
	decision.contextRevision = [32]byte{}
	if decision.Validate() == nil {
		t.Fatal("tampered context revision was accepted")
	}
}
