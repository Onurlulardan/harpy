package compiler

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
)

func TestContinuationIdentityIsNonzeroDeterministicAndTokenExact(t *testing.T) {
	_, _, base := fixtureSimpleQuery(t, "display", "display")
	none, noneErr := continuationIdentity(base)
	firstQuery := fixtureQueryAfter(t, base, "cursor-a")
	secondQuery := fixtureQueryAfter(t, base, "cursor-b")
	first, firstErr := continuationIdentity(firstQuery)
	firstAgain, againErr := continuationIdentity(firstQuery)
	second, secondErr := continuationIdentity(secondQuery)
	if noneErr != nil || firstErr != nil || againErr != nil || secondErr != nil ||
		none == [32]byte{} || first == [32]byte{} || second == [32]byte{} ||
		none == first || first != firstAgain || first == second {
		t.Fatal("execution continuation identity was not exact and domain-separated")
	}
}

func fixtureQueryAfter(
	testingT testing.TB, base resolved.RecordQuery, token string,
) resolved.RecordQuery {
	testingT.Helper()
	projections, _ := base.Projections()
	where, hasWhere, _ := base.Where()
	sorts, _ := base.Sorts()
	search, _ := base.Search()
	count, _ := base.Count()
	countPath, _ := base.CountPath()
	includes, _ := base.Includes()
	page, err := resolved.NewPage(1, &token)
	if err != nil {
		testingT.Fatal(err)
	}
	var filter *resolved.Filter
	if hasWhere {
		filter = &where
	}
	query, err := resolved.NewRecordQuery(
		projections, filter, sorts, search, page, count, countPath,
		count == api.CountModeExact, includes,
	)
	if err != nil {
		testingT.Fatal(err)
	}
	return query
}
