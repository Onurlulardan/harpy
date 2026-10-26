package authorization

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

func TestPolicyDecisionVisibilityBound(t *testing.T) {
	visibilities := make([]Visibility, maximumPolicyVisibilities)
	for index := range visibilities {
		visibility, err := NewAllVisibility(
			sha256.Sum256([]byte(fmt.Sprintf("resource-%d", index))),
		)
		if err != nil {
			t.Fatal(err)
		}
		visibilities[index] = visibility
	}
	result, _ := NewDecisionResult(sha256.Sum256([]byte("requirement")), true)
	if _, err := NewPolicyDecision(
		sha256.Sum256([]byte("request")), sha256.Sum256([]byte("revision")),
		[]DecisionResult{result}, visibilities,
	); err != nil {
		t.Fatalf("exact visibility bound: %v", err)
	}
	overflow, _ := NewAllVisibility(sha256.Sum256([]byte("overflow")))
	visibilities = append(visibilities, overflow)
	if _, err := NewPolicyDecision(
		sha256.Sum256([]byte("request")), sha256.Sum256([]byte("revision")),
		[]DecisionResult{result}, visibilities,
	); err == nil {
		t.Fatal("visibility overflow was accepted")
	}
}
