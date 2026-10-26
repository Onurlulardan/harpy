package v1query

import (
	"fmt"
	"strings"
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
)

func TestV1FilterBudgetsReturnExactStableIssues(t *testing.T) {
	t.Run("predicate count", func(t *testing.T) {
		snapshot := testManifest()
		parts := make([]string, 0, 51)
		for index := range 51 {
			field := fmt.Sprintf("field_%02d", index)
			snapshot.FilterFields[field] = FilterFieldBinding{
				Name: field, Shape: FilterShapeScalar,
			}
			parts = append(parts, fmt.Sprintf("filters[%s]=x", field))
		}
		decoder, err := newDecoder(staticManifestSource{snapshot: snapshot})
		if err != nil {
			t.Fatalf("NewDecoder() error = %v", err)
		}
		_, err = decoder.Decode(strings.Join(parts, "&"))
		requireQueryError(
			t, err, restqueryv1.ErrorQueryBudgetExceeded,
			"/filters", restqueryv1.IssuePredicateCountExceeded,
		)
	})

	t.Run("literal count", func(t *testing.T) {
		snapshot := testManifest()
		parts := make([]string, 0, 11)
		values := strings.Repeat("x%2C", 99) + "x"
		for index := range 11 {
			field := fmt.Sprintf("list_%02d", index)
			snapshot.FilterFields[field] = FilterFieldBinding{
				Name: field, Shape: FilterShapeScalar,
			}
			parts = append(parts,
				fmt.Sprintf(
					"filters[%s][operator]=in&filters[%s][value]=%s",
					field, field, values,
				),
			)
		}
		decoder, err := newDecoder(staticManifestSource{snapshot: snapshot})
		if err != nil {
			t.Fatalf("NewDecoder() error = %v", err)
		}
		_, err = decoder.Decode(strings.Join(parts, "&"))
		requireQueryError(
			t, err, restqueryv1.ErrorQueryBudgetExceeded,
			"/filters", restqueryv1.IssueLiteralCountExceeded,
		)
	})
}
