package recordkernel

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"strings"
	"testing"
)

func TestRelationPerFieldAndMutationLimitsAreIndependent(t *testing.T) {
	authority := fixtureActorAuthority()
	oneField := relationCatalog(1)
	if _, err := DecodeInput(oneField, ActionCreate, authority,
		relationDocument(1, MaximumRelationEdges)); err != nil {
		t.Fatalf("100 edges: %v", err)
	}
	if _, err := DecodeInput(oneField, ActionCreate, authority,
		relationDocument(1, MaximumRelationEdges+1)); !errors.Is(err, ErrBudgetExceeded) {
		t.Fatalf("101 edges: %v", err)
	}
	tenFields := relationCatalog(10)
	if _, err := DecodeInput(tenFields, ActionCreate, authority,
		relationDocument(10, MaximumRelationEdges)); err != nil {
		t.Fatalf("1000 targets: %v", err)
	}
	elevenFields := relationCatalog(11)
	document := relationMembers(11, MaximumRelationEdges)
	document["rel10"] = relationTargets(1, 1001)
	raw, _ := json.Marshal(document)
	if _, err := DecodeInput(elevenFields, ActionCreate, authority, raw); !errors.Is(err, ErrBudgetExceeded) {
		t.Fatalf("1001 targets: %v", err)
	}
}

func TestRecordRawAndCanonicalOneMiBBoundary(t *testing.T) {
	name := strings.Repeat("a", 36)
	catalog := textCatalog(name)
	empty := fmt.Sprintf(`{"%s":""}`, name)
	content := strings.Repeat("x", MaximumRecordBytes-len(empty))
	exact := []byte(fmt.Sprintf(`{"%s":"%s"}`, name, content))
	if len(exact) != MaximumRecordBytes {
		t.Fatalf("fixture bytes = %d", len(exact))
	}
	if _, err := DecodeInput(catalog, ActionCreate,
		fixtureActorAuthority(), exact); err != nil {
		t.Fatalf("1 MiB: %v", err)
	}
	over := append(append([]byte(nil), exact[:len(exact)-2]...), 'x', '"', '}')
	if _, err := DecodeInput(catalog, ActionCreate,
		fixtureActorAuthority(), over); !errors.Is(err, ErrInputInvalid) {
		t.Fatalf("1 MiB + 1: %v", err)
	}
}

func relationCatalog(count int) Catalog {
	fields := make([]Field, count)
	for index := range fields {
		fields[index] = Field{ID: fixtureV7(index + 10), PublicName: fmt.Sprintf("rel%02d", index),
			TypeID: TypeRelation, Role: RoleRelation,
			Nullable: true, WritePolicy: WriteActor, Generation: 1,
			Relation: &RelationProfile{Cardinality: "many"}}
	}
	return fixtureCatalog(fields)
}

func textCatalog(name string) Catalog {
	return fixtureCatalog([]Field{{ID: fixtureV7(9), PublicName: name,
		TypeID: TypeText, Role: RoleStored,
		Nullable: true, WritePolicy: WriteActor, Generation: 1}})
}

func fixtureCatalog(fields []Field) Catalog {
	byID, byName := indexFields(fields)
	return Catalog{tenantID: fixtureV7(1), publicationID: fixtureV7(4),
		publicationRevision: 1, formID: fixtureV7(2),
		formVersionID: fixtureV7(3), fields: fields, fieldsByID: byID,
		fieldsByName: byName, workflowActions: map[Action]struct{}{},
		limits: BudgetLimits{RecordInputBytes: MaximumRecordBytes,
			RelationTargetsPerMutation: MaximumRelationTargets,
			CalculationInstructions:    MaximumInstructions}, sha256: [32]byte{1}}
}

func relationDocument(fields, targets int) []byte {
	raw, _ := json.Marshal(relationMembers(fields, targets))
	return raw
}

func relationMembers(fields, targets int) map[string]any {
	result := make(map[string]any, fields)
	for index := 0; index < fields; index++ {
		result[fmt.Sprintf("rel%02d", index)] = relationTargets(targets, index*targets)
	}
	return result
}

func relationTargets(count, offset int) []string {
	result := make([]string, count)
	for index := range result {
		result[index] = fmt.Sprintf("71000000-0000-4000-8000-%012x", offset+index+1)
	}
	return result
}

func fixtureV7(value int) uuid.UUID {
	return uuid.MustParse(fmt.Sprintf("018f47f4-7b1b-7cc0-8000-%012x", value))
}

func fixtureActorAuthority() MutationAuthority {
	return NewActorAuthority()
}
