package v1query

import (
	"errors"
	"reflect"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestVerifiedManifestAliasesAreExplicitAndAuditable(t *testing.T) {
	decoder := mustDecoder(t)
	result, err := decoder.Decode(
		"q=Acme+Corp&sort[field]=legacy_status&" +
			"filters[legacy_status][operator]=same_as&" +
			"filters[legacy_status][value]=OPEN",
	)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	aliasIDs, _ := result.AliasIDs()
	want := []string{
		"field.status.v1", "operator.same-as.v1",
		"query.search.v1", "sort.status.v1",
	}
	if !reflect.DeepEqual(aliasIDs, want) {
		t.Fatalf("alias IDs = %#v, want %#v", aliasIDs, want)
	}
	request, _ := result.Request()
	search, found := request.Search()
	if !found || search.Term() != "Acme Corp" {
		t.Fatalf("search = %#v, found = %t", search, found)
	}
	if len(request.Sort()) != 1 || request.Sort()[0].Field() != "status" {
		t.Fatalf("sort = %#v", request.Sort())
	}
	where, _ := request.Where()
	requirePredicate(t, where, "status", api.OperatorEq, []string{"OPEN"})
}

func TestDecoderOwnsVerifiedManifestSnapshot(t *testing.T) {
	snapshot := testManifest()
	decoder, err := newDecoder(staticManifestSource{snapshot: snapshot})
	if err != nil {
		t.Fatalf("NewDecoder() error = %v", err)
	}
	delete(snapshot.FilterFields, "status")
	snapshot.DefaultLimit = 499
	result, err := decoder.Decode("filters[status]=OPEN")
	if err != nil {
		t.Fatalf("Decode() error after caller mutation = %v", err)
	}
	plan, _ := result.PagePlan()
	limit, _ := plan.Limit()
	if limit != 10 {
		t.Fatalf("owned default limit = %d, want 10", limit)
	}
}

func TestNewDecoderRejectsUnverifiedOrInvalidManifest(t *testing.T) {
	if _, err := newDecoder(nil); err == nil {
		t.Fatal("NewDecoder(nil) unexpectedly succeeded")
	}
	var typedNil *staticManifestSource
	if _, err := newDecoder(typedNil); err == nil {
		t.Fatal("NewDecoder(typed nil) unexpectedly succeeded")
	}
	sourceError := errors.New("signature rejected")
	if _, err := newDecoder(staticManifestSource{err: sourceError}); !errors.Is(err, sourceError) {
		t.Fatalf("source error = %v, want wrapped verifier error", err)
	}
	tests := []struct {
		name   string
		mutate func(*ManifestSnapshot)
	}{
		{"manifest ID", func(value *ManifestSnapshot) { value.ManifestID = "BAD" }},
		{"owner module", func(value *ManifestSnapshot) { value.OwnerModule = "Forms" }},
		{"domain", func(value *ManifestSnapshot) { value.Domain = "unknown" }},
		{"page policy", func(value *ManifestSnapshot) { value.MaximumPageSize = 501 }},
		{"missing authority", func(value *ManifestSnapshot) { value.FilterFields = nil }},
		{"built-in alias", func(value *ManifestSnapshot) {
			value.OperatorAliases["equals"] = OperatorBinding{
				Operator: "eq", AliasID: "operator.equals.v1",
			}
		}},
		{"ad hoc field alias", func(value *ManifestSnapshot) {
			value.FilterFields["old"] = FilterFieldBinding{
				Name: "new", Shape: FilterShapeScalar,
			}
		}},
		{"identity labeled as alias", func(value *ManifestSnapshot) {
			value.FilterFields["status"] = FilterFieldBinding{
				Name: "status", Shape: FilterShapeScalar, AliasID: "field.fake.v1",
			}
		}},
		{"canonical structured query key", func(value *ManifestSnapshot) {
			value.QueryKeyAliases["sort[field]"] = QueryKeyBinding{
				Key: "search", AliasID: "query.unreachable.v1",
			}
		}},
		{"unreachable filter key", func(value *ManifestSnapshot) {
			value.FilterFields["old[field]"] = FilterFieldBinding{
				Name: "status", Shape: FilterShapeScalar, AliasID: "field.unreachable.v1",
			}
		}},
		{"globally reused alias identity", func(value *ManifestSnapshot) {
			value.SortFields["legacy_status"] = SortFieldBinding{
				Name: "status", AliasID: "field.status.v1",
			}
		}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			snapshot := testManifest()
			test.mutate(&snapshot)
			if _, err := newDecoder(staticManifestSource{snapshot: snapshot}); err == nil {
				t.Fatal("invalid verified manifest unexpectedly succeeded")
			}
		})
	}
}

func TestAliasIDsAccessorReturnsCopy(t *testing.T) {
	result, err := mustDecoder(t).Decode("q=term")
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	first, _ := result.AliasIDs()
	first[0] = "changed"
	second, _ := result.AliasIDs()
	if second[0] != "query.search.v1" {
		t.Fatalf("stored alias ID changed to %q", second[0])
	}
}

func TestQueryAliasCannotPolluteCanonicalMember(t *testing.T) {
	_, err := mustDecoder(t).Decode("q=alias&search=canonical")
	requireQueryError(
		t, err, "QUERY_INVALID", "/search", "MEMBER_DUPLICATE",
	)
}
