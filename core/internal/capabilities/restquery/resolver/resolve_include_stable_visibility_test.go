package resolver

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func TestResolveIncludeStableVisibilityFollowsProjectedFields(t *testing.T) {
	tests := []struct {
		name         string
		selectFields []string
		wantHidden   bool
		wantVisible  int
	}{
		{name: "automatic stable", selectFields: []string{"name"}, wantHidden: true, wantVisible: 1},
		{name: "explicit stable", selectFields: []string{"id"}, wantVisible: 1},
		{name: "default stable", selectFields: nil, wantVisible: 2},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			projection, visible := resolveIncludeStableProjection(t, test.selectFields)
			hidden, err := projection.Hidden()
			if err != nil || hidden != test.wantHidden || visible != test.wantVisible {
				t.Fatalf("stable hidden/visible = %v/%d, err = %v", hidden, visible, err)
			}
		})
	}
}

func resolveIncludeStableProjection(
	t *testing.T,
	selectFields []string,
) (resolved.Projection, int) {
	t.Helper()
	fixture := newQueryFixture(t, true)
	include, err := api.NewRequestInclude(api.RequestIncludeInput{
		Relation: "children", Select: selectFields,
	})
	if err != nil {
		t.Fatal(err)
	}
	page, _ := api.NewRequestPage(10, nil)
	request, err := api.NewRecordQueryRequest(api.RecordQueryRequestInput{
		Select: []string{"name"}, Include: []api.RequestInclude{include},
		Page: page, Count: api.CountModeNone,
	})
	if err != nil {
		t.Fatal(err)
	}
	spec, err := Resolve(request, fixture.set)
	if err != nil {
		t.Fatal(err)
	}
	record, _, _ := spec.Record()
	includes, _ := record.Includes()
	projections, _ := includes[0].Projections()
	visible := visibleResolvedProjections(projections)
	for _, projection := range projections {
		field, _ := projection.Field()
		identifier, _ := field.ID()
		if identifier.String() == fixtureTargetID {
			return projection, visible
		}
	}
	t.Fatal("stable projection missing")
	return resolved.Projection{}, 0
}
