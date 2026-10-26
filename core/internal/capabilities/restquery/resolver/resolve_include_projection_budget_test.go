package resolver

import (
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestResolveIncludeProjectionBudgetCountsOnlyVisibleFields(t *testing.T) {
	tests := []struct {
		name string
		mode string
	}{
		{name: "automatic stable is hidden", mode: "automatic"},
		{name: "explicit stable is visible", mode: "explicit"},
		{name: "default stable is visible", mode: "default"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			accepted, err := projectionBudgetRequest(test.mode, 49)
			if err != nil {
				t.Fatalf("50 visible request rejected: %v", err)
			}
			catalogs := projectionBudgetCatalogs(t, defaultVisible(test.mode, 49))
			spec, err := Resolve(accepted, catalogs)
			if err != nil || spec.Validate() != nil {
				t.Fatalf("50 visible fields rejected: %v", err)
			}
			rejected, requestErr := projectionBudgetRequest(test.mode, 50)
			if test.mode != "default" {
				if requestErr == nil {
					t.Fatal("51 visible request was accepted")
				}
				return
			}
			if requestErr != nil {
				t.Fatalf("default projection request rejected early: %v", requestErr)
			}
			catalogs = projectionBudgetCatalogs(t, defaultVisible(test.mode, 50))
			assertResolutionError(
				t, rejected, catalogs, restqueryv1.ErrorQueryBudgetExceeded,
				"", restqueryv1.IssueSelectCountExceeded,
			)
		})
	}
}

func projectionBudgetRequest(mode string, targetVisible int) (api.QueryRequest, error) {
	selectFields := make([]string, 0, targetVisible)
	if mode == "explicit" {
		selectFields = append(selectFields, "id")
	}
	if mode != "default" {
		textCount := targetVisible
		if mode == "explicit" {
			textCount--
		}
		for index := 1; index <= textCount; index++ {
			selectFields = append(selectFields, budgetFieldName(index))
		}
	} else {
		selectFields = nil
	}
	include, err := api.NewRequestInclude(api.RequestIncludeInput{
		Relation: "children", Select: selectFields,
	})
	if err != nil {
		return api.QueryRequest{}, err
	}
	page, _ := api.NewRequestPage(10, nil)
	return api.NewRecordQueryRequest(api.RecordQueryRequestInput{
		Select: []string{"name"}, Include: []api.RequestInclude{include},
		Page: page, Count: api.CountModeNone,
	})
}

func defaultVisible(mode string, requested int) int {
	if mode == "default" {
		return requested
	}
	return 1
}
