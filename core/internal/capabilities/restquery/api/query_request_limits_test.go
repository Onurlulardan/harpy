package api

import (
	"fmt"
	"testing"
)

func TestRecordQueryRequestValidatesRootNamesAndUniqueness(t *testing.T) {
	valid := RecordQueryRequestInput{Page: mustRequestPage(t, 50, nil), Count: CountModeNone}
	for _, selectFields := range [][]string{
		{"field", "field"}, {""}, {"field\x00name"}, {"e\u0301"},
	} {
		input := valid
		input.Select = selectFields
		if _, err := NewRecordQueryRequest(input); err == nil {
			t.Fatalf("invalid select accepted: %#v", selectFields)
		}
	}
	input := valid
	input.Select = makeRequestNames(51, "field")
	if _, err := NewRecordQueryRequest(input); err == nil {
		t.Fatal("51 root select fields accepted")
	}
}

func TestRecordQueryRequestRequiresExplicitPageAndCount(t *testing.T) {
	for _, input := range []RecordQueryRequestInput{
		{Count: CountModeNone},
		{Page: mustRequestPage(t, 50, nil)},
		{Page: mustRequestPage(t, 50, nil), Count: CountMode(99)},
	} {
		if _, err := NewRecordQueryRequest(input); err == nil {
			t.Fatalf("invalid record mode accepted: %#v", input)
		}
	}
}

func TestRecordQueryRequestEnforcesProjectedFieldBudget(t *testing.T) {
	include := mustRequestInclude(t, RequestIncludeInput{
		Relation: "customer", Select: makeRequestNames(2, "child"),
	})
	input := RecordQueryRequestInput{
		Select: makeRequestNames(48, "root"), Include: []RequestInclude{include},
		Page: mustRequestPage(t, 50, nil), Count: CountModeNone,
	}
	if _, err := NewRecordQueryRequest(input); err != nil {
		t.Fatalf("50 projected fields rejected: %v", err)
	}
	input.Select = makeRequestNames(49, "root")
	if _, err := NewRecordQueryRequest(input); err == nil {
		t.Fatal("51 projected fields accepted")
	}
}

func TestRecordQueryRequestEnforcesRootIncludeEdges(t *testing.T) {
	input := RecordQueryRequestInput{
		Page: mustRequestPage(t, 50, nil), Count: CountModeNone,
	}
	for index := 0; index < 4; index++ {
		input.Include = append(input.Include, mustRequestInclude(t, RequestIncludeInput{
			Relation: fmt.Sprintf("relation_%d", index),
		}))
	}
	if _, err := NewRecordQueryRequest(input); err == nil {
		t.Fatal("four include edges accepted")
	}
}

func makeRequestNames(count int, prefix string) []string {
	result := make([]string, count)
	for index := range result {
		result[index] = fmt.Sprintf("%s_%d", prefix, index)
	}
	return result
}
