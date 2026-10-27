package authorization

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/google/uuid"
)

func TestRequirementIdentityAndAccessorsAreStable(t *testing.T) {
	requirement := mustRequirement(
		t, RequirementFieldFilter, "/include/0", "/include/0/where",
		uuid.MustParse(testFieldID), uuid.Nil,
	)
	same := mustRequirement(
		t, RequirementFieldFilter, "/include/0", "/include/0/where",
		uuid.MustParse(testFieldID), uuid.Nil,
	)
	identity, identityErr := requirement.ID()
	sameIdentity, sameErr := same.ID()
	kind, kindErr := requirement.Kind()
	fieldID, fieldFound, fieldErr := requirement.FieldID()
	relationID, relationFound, relationErr := requirement.RelationID()
	if identityErr != nil || sameErr != nil || identity == [32]byte{} ||
		identity != sameIdentity || kindErr != nil || kind != RequirementFieldFilter ||
		fieldErr != nil || !fieldFound || fieldID.String() != testFieldID ||
		relationErr != nil || relationFound || relationID != uuid.Nil {
		t.Fatalf("invalid accessors for stable requirement")
	}

	changed := mustRequirement(
		t, RequirementFieldHaving, "/include/0", "/include/0/where",
		uuid.MustParse(testFieldID), uuid.Nil,
	)
	changedIdentity, err := changed.ID()
	if err != nil || changedIdentity == identity {
		t.Fatal("authorization use did not bind requirement identity")
	}
}

func TestRequirementRejectsTamperAndRedacts(t *testing.T) {
	requirement := mustRequirement(
		t, RequirementExactCount, "", "/count", uuid.Nil, uuid.Nil,
	)
	tampered := requirement
	tampered.id = [32]byte{1}
	if tampered.Validate() == nil {
		t.Fatal("tampered requirement identity was accepted")
	}
	if fmt.Sprintf("%v", requirement) != authorizationRedaction ||
		requirement.LogValue().String() != authorizationRedaction {
		t.Fatal("requirement was not redacted")
	}
	if _, err := json.Marshal(requirement); err == nil {
		t.Fatal("requirement allowed direct JSON marshal")
	}
}
