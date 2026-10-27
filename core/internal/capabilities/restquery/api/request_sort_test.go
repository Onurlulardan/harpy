package api

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestRequestSortValidatesExactFieldAndDirection(t *testing.T) {
	sort, err := NewRequestSort("created_at", DirectionDescending)
	if err != nil {
		t.Fatal(err)
	}
	if sort.Field() != "created_at" || sort.Direction() != DirectionDescending {
		t.Fatalf("unexpected sort: %#v", sort)
	}
	for _, field := range []string{
		"", "_included", "field\x00name", "e\u0301", string([]byte{0xff}),
	} {
		if _, err := NewRequestSort(field, DirectionAscending); err == nil {
			t.Fatalf("invalid field %q accepted", field)
		}
	}
	if _, err := NewRequestSort("field", Direction(99)); err == nil {
		t.Fatal("invalid direction accepted")
	}
}

func TestRequestSortListIsUniqueAndBounded(t *testing.T) {
	one := mustRequestSort(t, "one")
	if _, err := cloneRequestSorts([]RequestSort{one, one}); err == nil {
		t.Fatal("duplicate sort field accepted")
	}
	values := make([]RequestSort, 6)
	for index := range values {
		values[index] = mustRequestSort(t, fmt.Sprintf("field_%d", index))
	}
	if _, err := cloneRequestSorts(values); err == nil {
		t.Fatal("six sort fields accepted")
	}
}

func TestRequestSortIsRedactedAndCannotMarshal(t *testing.T) {
	sort := mustRequestSort(t, "secret_field")
	for _, value := range []string{fmt.Sprint(sort), fmt.Sprintf("%#v", sort)} {
		if value != queryValueRedaction {
			t.Fatalf("sort formatting leaked: %q", value)
		}
	}
	if _, err := json.Marshal(sort); err == nil {
		t.Fatal("opaque sort silently marshaled")
	}
}

func mustRequestSort(t *testing.T, field string) RequestSort {
	t.Helper()
	value, err := NewRequestSort(field, DirectionAscending)
	if err != nil {
		t.Fatal(err)
	}
	return value
}
