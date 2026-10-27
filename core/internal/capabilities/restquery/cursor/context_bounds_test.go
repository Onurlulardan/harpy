package cursor

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestContextEnforcesPageSortCountAndUniqueFields(t *testing.T) {
	for _, size := range []int{0, maximumPageSize + 1} {
		fixture := baseContextFixture(t)
		fixture.pageSize = size
		if _, err := tryBuildContext(fixture); err == nil {
			t.Fatalf("page size %d accepted", size)
		}
	}
	fixture := baseContextFixture(t)
	fixture.sorts = nil
	if _, err := tryBuildContext(fixture); err == nil {
		t.Fatal("empty effective sort accepted")
	}
	fixture = baseContextFixture(t)
	fixture.sorts = append(fixture.sorts, fixture.sorts[0])
	if _, err := tryBuildContext(fixture); err == nil {
		t.Fatal("duplicate effective sort accepted")
	}
	fixture = baseContextFixture(t)
	fixture.sorts = testTextSorts(t, maximumSorts+1, 16)
	if _, err := tryBuildContext(fixture); err == nil {
		t.Fatal("more than six effective sorts accepted")
	}
}

func TestContextCanonicalizesAndDefensivelyCopiesInputs(t *testing.T) {
	fixture := baseContextFixture(t)
	fixture.mandatory = []byte(` { "z": 1, "a": true } `)
	inputMandatory := fixture.mandatory
	inputSorts := fixture.sorts
	context := buildContext(t, fixture)
	inputMandatory[3] = 'x'
	inputSorts[0].canonicalBound = 1
	canonical, err := context.MandatoryPredicateCanonical()
	if err != nil || string(canonical) != `{"a":true,"z":1}` {
		t.Fatalf("canonical predicate = %s, %v", canonical, err)
	}
	canonical[0] = 'x'
	again, _ := context.MandatoryPredicateCanonical()
	if string(again) != `{"a":true,"z":1}` {
		t.Fatal("mandatory predicate accessor leaked storage")
	}
	sorts, _ := context.Sorts()
	sorts[0].canonicalBound = 1
	stored, _ := context.Sorts()
	if stored[0].canonicalBound != 128 {
		t.Fatal("sort accessor leaked storage")
	}
}

func TestMandatoryPredicateAcceptsExact64KiBBoundary(t *testing.T) {
	fixture := baseContextFixture(t)
	fixture.mandatory = append([]byte{'"'}, bytes.Repeat(
		[]byte{'a'}, maximumMandatoryBytes-2,
	)...)
	fixture.mandatory = append(fixture.mandatory, '"')
	context, err := tryBuildContext(fixture)
	if err != nil {
		t.Fatalf("exact mandatory bound rejected: %v", err)
	}
	canonical, _ := context.MandatoryPredicateCanonical()
	if len(canonical) != maximumMandatoryBytes {
		t.Fatalf("canonical mandatory bytes = %d", len(canonical))
	}
}

func TestEncodeEnforcesLastValueAndPlaintextBounds(t *testing.T) {
	fixture := baseContextFixture(t)
	fixture.sorts = testTextSorts(t, 4, maximumCanonicalValueBytes)
	context := buildContext(t, fixture)
	codec, _ := newCodec(
		testKeyring(t, "current", 1, "previous", 2),
		bytes.NewReader(bytes.Repeat([]byte{1}, nonceBytes*3)),
		func() time.Time { return testNow },
	)
	if _, err := codec.Encode(context, nil); err == nil {
		t.Fatal("wrong last-value count accepted")
	}
	wrongKind := []LastValue{scalarLast(t, api.LogicalTypeInteger, "1")}
	wrongKind = append(wrongKind, NewNullLastValue(), NewNullLastValue(), NewNullLastValue())
	if _, err := codec.Encode(context, wrongKind); err == nil {
		t.Fatal("last-value logical type mismatch accepted")
	}
	oversized := strings.Repeat("a", maximumCanonicalValueBytes+1)
	values := []LastValue{scalarLast(t, api.LogicalTypeText, oversized)}
	values = append(values, NewNullLastValue(), NewNullLastValue(), NewNullLastValue())
	if _, err := codec.Encode(context, values); err == nil {
		t.Fatal("per-field canonical bound exceeded")
	}
	values = values[:0]
	for range fixture.sorts {
		values = append(values, scalarLast(
			t, api.LogicalTypeText, strings.Repeat("b", maximumCanonicalValueBytes),
		))
	}
	if _, err := codec.Encode(context, values); err == nil {
		t.Fatal("8 KiB plaintext limit exceeded")
	}
}

func testTextSorts(t testing.TB, count, bound int) []SortDefinition {
	ids := []string{
		"10000000-0000-4000-8000-000000000001", "10000000-0000-4000-8000-000000000002",
		"10000000-0000-4000-8000-000000000003", "10000000-0000-4000-8000-000000000004",
		"10000000-0000-4000-8000-000000000005", "10000000-0000-4000-8000-000000000006",
		"10000000-0000-4000-8000-000000000007",
	}
	result := make([]SortDefinition, count)
	for index := range result {
		result[index] = mustSort(t, ids[index], api.DirectionAscending, api.LogicalTypeText, bound)
	}
	return result
}
