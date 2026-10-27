package profile

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/google/uuid"
)

func TestPublishedSchemaDefinitionIsOrderIndependentAndImmutable(t *testing.T) {
	first := mustField(t, baseFieldInput(fieldOneID, "nickname", LogicalString))
	second := mustField(t, baseFieldInput(fieldTwoID, "start_date", LogicalDate))
	left := mustSchema(t, schemaOneID, 1, second, first)
	right := mustSchema(t, schemaOneID, 1, first, second)
	if !bytes.Equal(left.DefinitionJSON(), right.DefinitionJSON()) ||
		left.DefinitionSHA256() != right.DefinitionSHA256() {
		t.Fatal("schema definition depends on caller field order")
	}
	returned := left.DefinitionJSON()
	returned[0] = '!'
	if !bytes.Equal(left.DefinitionJSON(), right.DefinitionJSON()) {
		t.Fatal("schema definition bytes escaped ownership")
	}
	expected := left.DefinitionSHA256()
	if _, err := NewPublishedSchema(SchemaInput{
		ID: schemaOneID, Resource: ResourceUser, Version: 1, State: SchemaActive,
		Fields: []Field{first, second}, ExpectedDefinitionSHA256: expected[:],
	}); err != nil {
		t.Fatalf("verified schema digest: %v", err)
	}
}

func TestPublishedSchemaRejectsIdentityAndDigestDrift(t *testing.T) {
	first := mustField(t, baseFieldInput(fieldOneID, "nickname", LogicalString))
	duplicateKey := mustField(t, baseFieldInput(fieldTwoID, "nickname", LogicalString))
	otherResourceInput := baseFieldInput(fieldThreeID, "legal_name", LogicalString)
	otherResourceInput.Resource = ResourceOrganization
	otherResource := mustField(t, otherResourceInput)
	badDigest := make([]byte, 32)
	tests := map[string]SchemaInput{
		"nil id":         {ID: uuid.Nil, Resource: ResourceUser, Version: 1, State: SchemaActive, Fields: []Field{}},
		"zero version":   {ID: schemaOneID, Resource: ResourceUser, State: SchemaActive, Fields: []Field{}},
		"nil fields":     {ID: schemaOneID, Resource: ResourceUser, Version: 1, State: SchemaActive},
		"duplicate id":   {ID: schemaOneID, Resource: ResourceUser, Version: 1, State: SchemaActive, Fields: []Field{first, first}},
		"duplicate key":  {ID: schemaOneID, Resource: ResourceUser, Version: 1, State: SchemaActive, Fields: []Field{first, duplicateKey}},
		"cross resource": {ID: schemaOneID, Resource: ResourceUser, Version: 1, State: SchemaActive, Fields: []Field{otherResource}},
		"digest drift":   {ID: schemaOneID, Resource: ResourceUser, Version: 1, State: SchemaActive, Fields: []Field{first}, ExpectedDefinitionSHA256: badDigest},
	}
	for name, input := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := NewPublishedSchema(input); err == nil {
				t.Fatal("invalid schema was accepted")
			}
		})
	}
}

func TestPublishedSchemaEnforcesActiveFieldHardLimit(t *testing.T) {
	fields := make([]Field, 0, MaximumActiveFields+1)
	for index := 1; index <= MaximumActiveFields+1; index++ {
		id := uuid.MustParse(fmt.Sprintf("%08x-0000-4000-8000-000000000001", index))
		fields = append(fields, mustField(t, baseFieldInput(id, "field_"+id.String()[:8], LogicalString)))
	}
	if _, err := NewPublishedSchema(SchemaInput{
		ID: schemaOneID, Resource: ResourceUser, Version: 1,
		State: SchemaActive, Fields: fields,
	}); err == nil {
		t.Fatal("active field hard limit was not enforced")
	}
}
