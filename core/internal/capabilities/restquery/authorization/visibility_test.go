package authorization

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"testing"
)

func TestVisibilityModesAndFilterAreImmutable(t *testing.T) {
	resource := sha256.Sum256([]byte("resource"))
	all, allErr := NewAllVisibility(resource)
	none, noneErr := NewNoVisibility(resource)
	predicate, predicateErr := NewPredicateVisibility(resource, mandatoryLeaf(t))
	if allErr != nil || noneErr != nil || predicateErr != nil {
		t.Fatalf("visibility constructors: %v %v %v", allErr, noneErr, predicateErr)
	}
	cases := []struct {
		visibility Visibility
		mode       VisibilityMode
		hasFilter  bool
	}{
		{all, VisibilityAll, false},
		{none, VisibilityNone, false},
		{predicate, VisibilityPredicate, true},
	}
	for _, test := range cases {
		identity, identityErr := test.visibility.ResourceRequirementSHA256()
		mode, modeErr := test.visibility.Mode()
		filter, found, filterErr := test.visibility.Filter()
		if identityErr != nil || identity != resource || modeErr != nil ||
			mode != test.mode || filterErr != nil || found != test.hasFilter {
			t.Fatalf("visibility %s accessors are invalid", test.mode)
		}
		if found {
			filter.validated = false
			if test.visibility.Validate() != nil {
				t.Fatal("returned filter mutated visibility")
			}
		}
	}
}

func TestVisibilityRejectsZeroAndRedacts(t *testing.T) {
	if _, err := NewAllVisibility([32]byte{}); err == nil {
		t.Fatal("zero resource requirement was accepted")
	}
	visibility, err := NewAllVisibility(sha256.Sum256([]byte("resource")))
	if err != nil {
		t.Fatal(err)
	}
	if fmt.Sprintf("%v", visibility) != authorizationRedaction ||
		visibility.LogValue().String() != authorizationRedaction {
		t.Fatal("visibility was not redacted")
	}
	if _, err := json.Marshal(visibility); err == nil {
		t.Fatal("visibility allowed direct JSON marshal")
	}
}
