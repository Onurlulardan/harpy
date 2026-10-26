package resolver

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func TestResolveRecordBindsEffectiveStateAndIdentity(t *testing.T) {
	fixture := newQueryFixture(t, true)
	request := fixtureRecordRequest(t, nil)
	spec, err := Resolve(request, fixture.set)
	if err != nil || spec.Validate() != nil {
		if queryError, ok := err.(api.QueryError); ok {
			t.Fatalf("record resolution failed: %v", queryError.Issues())
		}
		t.Fatalf("record resolution failed: %v", err)
	}
	record, found, err := spec.Record()
	if err != nil || !found {
		t.Fatalf("record spec missing: %v", err)
	}
	assertRootRecordShape(t, record)
	assertIncludedRecordShape(t, record)
	hash, err := spec.IdentityHash()
	if err != nil {
		t.Fatal(err)
	}
	after := "opaque-cursor"
	second, err := Resolve(fixtureRecordRequest(t, &after), fixture.set)
	if err != nil {
		t.Fatal(err)
	}
	secondHash, err := second.IdentityHash()
	if err != nil || secondHash != hash {
		t.Fatal("page cursor changed the query identity")
	}
	size, sizeErr := spec.IdentitySize()
	secondSize, secondSizeErr := second.IdentitySize()
	if sizeErr != nil || secondSizeErr != nil || size < 1 || size > 64*1024 ||
		secondSize != size {
		t.Fatalf("identity size = %d, err = %v", size, sizeErr)
	}
}

func fixtureRecordRequest(t *testing.T, after *string) api.QueryRequest {
	t.Helper()
	where, err := api.NewRequestPredicate(
		"amount", api.OperatorGT, []api.Literal{mustTextLiteral(t, "1.20")},
	)
	if err != nil {
		t.Fatal(err)
	}
	includeWhere, _ := api.NewRequestPredicate(
		"name", api.OperatorContains, []api.Literal{mustTextLiteral(t, "Ada")},
	)
	includeSort, _ := api.NewRequestSort("name", api.DirectionDescending)
	include, err := api.NewRequestInclude(api.RequestIncludeInput{
		Relation: "children", Select: []string{"name"}, Where: &includeWhere,
		Sort: []api.RequestSort{includeSort},
	})
	if err != nil {
		t.Fatal(err)
	}
	nameSort, _ := api.NewRequestSort("name", api.DirectionAscending)
	idSort, _ := api.NewRequestSort("id", api.DirectionDescending)
	page, _ := api.NewRequestPage(25, after)
	search, _ := api.NewRequestSearch("Ada")
	request, err := api.NewRecordQueryRequest(api.RecordQueryRequestInput{
		Select: []string{"amount"}, Where: &where,
		Sort: []api.RequestSort{nameSort, idSort}, Page: page,
		Search: &search, Count: api.CountModeExact,
		Include: []api.RequestInclude{include},
	})
	if err != nil {
		t.Fatal(err)
	}
	return request
}

func assertRootRecordShape(t *testing.T, record resolved.RecordQuery) {
	t.Helper()
	projections, err := record.Projections()
	if err != nil || len(projections) != 2 {
		t.Fatalf("root projections = %d, err = %v", len(projections), err)
	}
	firstHidden, _ := projections[0].Hidden()
	secondHidden, _ := projections[1].Hidden()
	secondField, _ := projections[1].Field()
	secondID, _ := secondField.ID()
	if firstHidden || !secondHidden || secondID.String() != fixtureRootID {
		t.Fatal("root hidden stable projection is not canonical")
	}
	sorts, err := record.Sorts()
	if err != nil || len(sorts) != 2 {
		t.Fatalf("root sorts = %d, err = %v", len(sorts), err)
	}
	finalField, _ := sorts[1].Field()
	finalID, _ := finalField.ID()
	direction, _ := sorts[1].Direction()
	implicit, _ := sorts[1].Implicit()
	if finalID.String() != fixtureRootID || direction != api.DirectionDescending || implicit {
		t.Fatal("explicit stable descending sort was not preserved as the final tie")
	}
	count, _ := record.Count()
	search, _ := record.Search()
	if count != api.CountModeExact || len(search) != 1 {
		t.Fatal("record count or search state was lost")
	}
}

func assertIncludedRecordShape(t *testing.T, record resolved.RecordQuery) {
	t.Helper()
	includes, err := record.Includes()
	if err != nil || len(includes) != 1 {
		t.Fatalf("includes = %d, err = %v", len(includes), err)
	}
	include := includes[0]
	projections, _ := include.Projections()
	limit, _ := include.Limit()
	where, hasWhere, _ := include.Where()
	sorts, _ := include.Sorts()
	if len(projections) != 2 || limit != 50 || !hasWhere || where.Validate() != nil ||
		len(sorts) != 2 {
		t.Fatal("include effective state was not retained")
	}
	stableField, _ := projections[1].Field()
	stableID, _ := stableField.ID()
	hidden, _ := projections[1].Hidden()
	finalField, _ := sorts[1].Field()
	finalID, _ := finalField.ID()
	direction, _ := sorts[1].Direction()
	implicit, _ := sorts[1].Implicit()
	if stableID.String() != fixtureTargetID || !hidden || finalID != stableID ||
		direction != api.DirectionAscending || !implicit {
		t.Fatal("include implicit hidden stable field or tie sort is invalid")
	}
}
