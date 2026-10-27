package runtime

import "github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/adapters/postgres/executor"

func recordResult(result executor.Result) (Result, error) {
	page, err := result.Page()
	if err != nil {
		return Result{}, apiRuntimeError()
	}
	items, itemsErr := page.Records()
	hasMore, moreErr := page.HasMore()
	next, _, cursorErr := page.NextCursor()
	count, hasCount, countErr := result.ExactCount()
	if itemsErr != nil || moreErr != nil || cursorErr != nil || countErr != nil {
		return Result{}, apiRuntimeError()
	}
	var exact *int64
	if hasCount {
		exact = &count
	}
	return newResult(items, hasMore, next, exact), nil
}
