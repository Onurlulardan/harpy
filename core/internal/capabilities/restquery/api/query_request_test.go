package api

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestNewRecordQueryRequestBuildsImmutableValue(t *testing.T) {
	page := mustRequestPage(t, 50, nil)
	where := mustPredicate(t)
	search := mustRequestSearch(t, "Cafe\u0301")
	include := mustRequestInclude(t, RequestIncludeInput{
		Relation: "customer", Select: []string{"id"},
	})
	input := RecordQueryRequestInput{
		Select: []string{"id", "status"}, Where: &where,
		Sort: []RequestSort{mustRequestSort(t, "created_at")}, Page: page,
		Search: &search, Count: CountModeExact, Include: []RequestInclude{include},
	}
	request, err := NewRecordQueryRequest(input)
	if err != nil {
		t.Fatal(err)
	}
	input.Select[0] = "changed"
	input.Sort[0] = RequestSort{}
	input.Include[0] = RequestInclude{}
	where = RequestFilter{}
	search = RequestSearch{}
	if request.Mode() != QueryModeRecord || request.Count() != CountModeExact {
		t.Fatalf("unexpected mode/count: %s/%s", request.Mode(), request.Count())
	}
	storedSearch, found := request.Search()
	if request.Select()[0] != "id" || !found || storedSearch.Term() != "Caf\u00e9" {
		t.Fatal("request retained mutable input")
	}
	assertRecordRequestAccessorsAreDefensive(t, request)
}

func TestRecordQueryRequestAllowsAbsentOptionalValues(t *testing.T) {
	request, err := NewRecordQueryRequest(RecordQueryRequestInput{
		Page: mustRequestPage(t, 1, nil), Count: CountModeNone,
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, found := request.Where(); found {
		t.Fatal("absent where became present")
	}
	if _, found := request.Search(); found {
		t.Fatal("absent search became present")
	}
	if request.Select() == nil || request.Sort() == nil || request.Includes() == nil {
		t.Fatal("canonical empty collections must be non-nil")
	}
}

func TestQueryRequestIsRedactedAndCannotMarshal(t *testing.T) {
	input := RecordQueryRequestInput{
		Select: []string{"secret_field"}, Page: mustRequestPage(t, 50, nil),
		Count: CountModeNone,
	}
	request, err := NewRecordQueryRequest(input)
	if err != nil {
		t.Fatal(err)
	}
	for _, value := range []string{
		fmt.Sprint(request), fmt.Sprintf("%#v", request),
		fmt.Sprint(input), fmt.Sprintf("%#v", input),
	} {
		if value != queryValueRedaction {
			t.Fatalf("request formatting leaked: %q", value)
		}
	}
	for _, value := range []any{request, input, request.Mode()} {
		if _, err := json.Marshal(value); err == nil {
			t.Fatalf("%T silently marshaled", value)
		}
	}
}

func assertRecordRequestAccessorsAreDefensive(t *testing.T, request QueryRequest) {
	t.Helper()
	selectFields := request.Select()
	selectFields[0] = "changed"
	sorts := request.Sort()
	sorts[0] = RequestSort{}
	includes := request.Includes()
	includes[0] = RequestInclude{}
	where, _ := request.Where()
	children := where.Children()
	if len(children) > 0 {
		children[0] = RequestFilter{}
	}
	if request.Select()[0] != "id" || request.Sort()[0].Field() != "created_at" ||
		request.Includes()[0].Relation() != "customer" {
		t.Fatal("request accessor exposed mutable storage")
	}
}

func mustRequestPage(t *testing.T, size int, after *string) RequestPage {
	t.Helper()
	page, err := NewRequestPage(size, after)
	if err != nil {
		t.Fatal(err)
	}
	return page
}

func mustRequestSearch(t *testing.T, term string) RequestSearch {
	t.Helper()
	search, err := NewRequestSearch(term)
	if err != nil {
		t.Fatal(err)
	}
	return search
}
