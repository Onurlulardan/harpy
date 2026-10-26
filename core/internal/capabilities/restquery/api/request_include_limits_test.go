package api

import "testing"

func TestRequestIncludeEnforcesDepthTwo(t *testing.T) {
	leaf := mustRequestInclude(t, RequestIncludeInput{Relation: "leaf"})
	depthTwo := mustRequestInclude(t, RequestIncludeInput{
		Relation: "parent", Include: []RequestInclude{leaf},
	})
	if depthTwo.depth != maximumIncludeDepth {
		t.Fatalf("depth = %d", depthTwo.depth)
	}
	if _, err := NewRequestInclude(RequestIncludeInput{
		Relation: "grandparent", Include: []RequestInclude{depthTwo},
	}); err == nil {
		t.Fatal("include depth three accepted")
	}
}

func TestRequestIncludeEnforcesThreeEdgesAndUniqueSiblings(t *testing.T) {
	one := mustRequestInclude(t, RequestIncludeInput{Relation: "one"})
	two := mustRequestInclude(t, RequestIncludeInput{Relation: "two"})
	three := mustRequestInclude(t, RequestIncludeInput{Relation: "three"})
	if _, err := cloneRequestIncludes([]RequestInclude{one, two, three}); err != nil {
		t.Fatalf("three edges rejected: %v", err)
	}
	if _, err := cloneRequestIncludes([]RequestInclude{one, two, three, one}); err == nil {
		t.Fatal("four edges accepted")
	}
	if _, err := cloneRequestIncludes([]RequestInclude{one, one}); err == nil {
		t.Fatal("duplicate sibling relation accepted")
	}
}

func TestRequestIncludeDoesNotResolveCatalogSemantics(t *testing.T) {
	limit := 2
	include, err := NewRequestInclude(RequestIncludeInput{
		Relation: "unknown_but_structurally_valid", Limit: &limit,
	})
	if err != nil {
		t.Fatalf("constructor assumed catalog responsibility: %v", err)
	}
	if include.Relation() != "unknown_but_structurally_valid" {
		t.Fatal("relation changed")
	}
}
