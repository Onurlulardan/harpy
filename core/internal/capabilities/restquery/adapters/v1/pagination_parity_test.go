package v1query

import (
	"encoding/json"
	"os"
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
)

type paginationFixture struct {
	Name               string `json:"name"`
	RawQuery           string `json:"raw_query"`
	Page               int64  `json:"page"`
	Limit              int    `json:"limit"`
	PredecessorIndex   int64  `json:"predecessor_index"`
	HasPredecessorSeek bool   `json:"has_predecessor"`
}

func TestV1DynamicListPaginationParityFixtures(t *testing.T) {
	// Characterizes ../backend/shared/utils/query/query_params.go plus the
	// form-engine dynamic-list limit normalization before intentional V2 budgets.
	contents, err := os.ReadFile("testdata/v1_dynamic_list_pagination.json")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	var fixtures []paginationFixture
	if err := json.Unmarshal(contents, &fixtures); err != nil {
		t.Fatalf("decode fixture: %v", err)
	}
	decoder := mustDecoder(t)
	for _, fixture := range fixtures {
		t.Run(fixture.Name, func(t *testing.T) {
			result, err := decoder.Decode(fixture.RawQuery)
			if err != nil {
				t.Fatalf("Decode() error = %v", err)
			}
			plan, err := result.PagePlan()
			if err != nil {
				t.Fatalf("PagePlan() error = %v", err)
			}
			page, _ := plan.Page()
			limit, _ := plan.Limit()
			index, found, _ := plan.PredecessorIndex()
			if page != fixture.Page || limit != fixture.Limit ||
				index != fixture.PredecessorIndex || found != fixture.HasPredecessorSeek {
				t.Fatalf(
					"plan = (%d,%d,%d,%t), want (%d,%d,%d,%t)",
					page, limit, index, found, fixture.Page, fixture.Limit,
					fixture.PredecessorIndex, fixture.HasPredecessorSeek,
				)
			}
			request, _ := result.Request()
			if request.Page().Size() != fixture.Limit {
				t.Fatalf("canonical page size = %d, want %d", request.Page().Size(), fixture.Limit)
			}
			if _, hasAfter := request.Page().After(); hasAfter {
				t.Fatal("canonical QueryRequest unexpectedly contains a cursor")
			}
		})
	}
}

func TestV1DynamicListLimitIntentionalSafetyDifference(t *testing.T) {
	decoder := mustDecoder(t)
	for _, raw := range []string{"limit=0", "limit=-1", "limit=501", "limit=1001"} {
		_, err := decoder.Decode(raw)
		requireQueryError(
			t, err, restqueryv1.ErrorQueryBudgetExceeded,
			"/limit", restqueryv1.IssuePageSizeExceeded,
		)
	}
}

func TestV1PageCheckedArithmetic(t *testing.T) {
	decoder := mustDecoder(t)
	queries := []string{
		"page=9223372036854775807&limit=500",
		"page=18446744073709551615&limit=10",
		"page=999999999999999999999999999999999999&limit=10",
	}
	for _, raw := range queries {
		_, err := decoder.Decode(raw)
		requireQueryError(
			t, err, restqueryv1.ErrorQueryBudgetExceeded,
			"/page", restqueryv1.IssuePageOrdinalOverflow,
		)
	}
}

func TestMalformedV1PageIsNotSilentlyCoerced(t *testing.T) {
	_, err := mustDecoder(t).Decode("page=one")
	requireQueryError(
		t, err, restqueryv1.ErrorQueryInvalid,
		"/page", restqueryv1.IssueValueNotCanonical,
	)
}

func TestPaginationAliasFailurePointsToAcceptedWireMember(t *testing.T) {
	_, err := mustDecoder(t).Decode("per_page=501")
	requireQueryError(
		t, err, restqueryv1.ErrorQueryBudgetExceeded,
		"/per_page", restqueryv1.IssuePageSizeExceeded,
	)
}
