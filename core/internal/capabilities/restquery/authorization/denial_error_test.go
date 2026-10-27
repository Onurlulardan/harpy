package authorization

import (
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/google/uuid"
)

func TestQueryErrorForDenialsReturnsCanonicalFieldIssues(t *testing.T) {
	fieldID := uuid.MustParse(testFieldID)
	denied := []Requirement{
		mustRequirement(
			t, RequirementFieldFilter, "/where", "/where/2/field",
			fieldID, uuid.Nil,
		),
		mustRequirement(
			t, RequirementFieldSelect, "/select", "/select/0",
			fieldID, uuid.Nil,
		),
		mustRequirement(
			t, RequirementFieldSort, "/sort", "/sort/1/field",
			fieldID, uuid.Nil,
		),
	}
	queryError, found, err := queryErrorForDenials(denied)
	if err != nil || !found {
		t.Fatalf("queryErrorForDenials() found=%v error=%v", found, err)
	}
	if queryError.Code() != restqueryv1.ErrorQueryFieldForbidden {
		t.Fatalf("code = %s", queryError.Code())
	}
	issues := queryError.Issues()
	wantPaths := []string{"/select/0", "/sort/1/field", "/where/2/field"}
	if len(issues) != len(wantPaths) {
		t.Fatalf("issues = %#v", issues)
	}
	for index, path := range wantPaths {
		if issues[index].Path != path ||
			issues[index].Code != string(restqueryv1.IssueFieldForbidden) {
			t.Fatalf("issue %d = %#v", index, issues[index])
		}
	}
}

func TestQueryErrorForDenialsHidesGenericAndMixedDetails(t *testing.T) {
	fieldID := uuid.MustParse(testFieldID)
	denied := []Requirement{
		mustRequirement(
			t, RequirementFieldSelect, "/select", "/select/0",
			fieldID, uuid.Nil,
		),
		mustRequirement(
			t, RequirementExactCount, "", "/count",
			uuid.Nil, uuid.Nil,
		),
	}
	queryError, found, err := queryErrorForDenials(denied)
	if err != nil || !found {
		t.Fatalf("queryErrorForDenials() found=%v error=%v", found, err)
	}
	if queryError.Code() != restqueryv1.ErrorQueryForbidden ||
		len(queryError.Issues()) != 0 {
		t.Fatalf("generic denial leaked details: %s %#v", queryError.Code(), queryError.Issues())
	}
}

func TestQueryErrorForDenialsHandlesEmptyAndInvalidInput(t *testing.T) {
	if _, found, err := queryErrorForDenials(nil); err != nil || found {
		t.Fatalf("empty denials found=%v error=%v", found, err)
	}
	if _, found, err := queryErrorForDenials([]Requirement{{}}); err == nil || found {
		t.Fatalf("invalid denial found=%v error=%v", found, err)
	}
}

func TestFieldRequirementClassificationIsExhaustive(t *testing.T) {
	fieldKinds := map[RequirementKind]bool{
		RequirementFieldSelect: true, RequirementFieldFilter: true,
		RequirementFieldSort: true, RequirementFieldSearch: true,
		RequirementFieldGroup: true, RequirementFieldMeasure: true,
		RequirementFieldHaving: true,
	}
	for kind := RequirementResourceRead; kind <= RequirementAggregateSort; kind++ {
		if fieldRequirement(kind) != fieldKinds[kind] {
			t.Fatalf("fieldRequirement(%s) = %v", kind, fieldRequirement(kind))
		}
	}
}
