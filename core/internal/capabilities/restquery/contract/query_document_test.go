package contract

import (
	"bytes"
	"testing"
)

func TestQueryDocumentDistinguishesEmptyDocumentFromZero(t *testing.T) {
	t.Parallel()
	document, err := NewQueryDocument(QueryDocumentInput{})
	if err != nil {
		t.Fatal(err)
	}
	if err := document.Validate(); err != nil {
		t.Fatalf("empty object document is invalid: %v", err)
	}
	var zero QueryDocument
	if err := zero.Validate(); err == nil {
		t.Fatal("zero QueryDocument was accepted")
	}
}

func TestQueryDocumentPreservesEveryRootMember(t *testing.T) {
	t.Parallel()
	member := mustRawMember(t, `{"marker":true}`)
	document, err := NewQueryDocument(QueryDocumentInput{
		Select: member, Where: member, Sort: member, Page: member,
		Search: member, Count: member, Include: member, Aggregate: member,
	})
	if err != nil {
		t.Fatal(err)
	}
	accessors := []func() RawMember{
		document.Select, document.Where, document.Sort, document.Page,
		document.Search, document.Count, document.Include, document.Aggregate,
	}
	for _, accessor := range accessors {
		value, present := accessor().Value()
		if !present || !bytes.Equal(value, []byte(`{"marker":true}`)) {
			t.Fatalf("root member was not preserved: %q", value)
		}
	}
}

func TestQueryDocumentClonesInputAndAccessorMembers(t *testing.T) {
	t.Parallel()
	input := QueryDocumentInput{Select: mustRawMember(t, `["stable"]`)}
	document, err := NewQueryDocument(input)
	if err != nil {
		t.Fatal(err)
	}
	input.Select.value[2] = 'X'
	accessorCopy := document.Select()
	accessorCopy.value[2] = 'Y'
	value, _ := document.Select().Value()
	if string(value) != `["stable"]` {
		t.Fatalf("document storage was mutable: %s", value)
	}
}

func TestQueryDocumentRejectsInvalidMember(t *testing.T) {
	t.Parallel()
	_, err := NewQueryDocument(QueryDocumentInput{
		Where: RawMember{present: true, value: []byte(`{"secret":`)},
	})
	if err == nil {
		t.Fatal("invalid member was accepted")
	}
}

func mustRawMember(t *testing.T, value string) RawMember {
	t.Helper()
	member, err := NewRawMember([]byte(value))
	if err != nil {
		t.Fatal(err)
	}
	return member
}
