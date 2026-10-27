package supporthttp

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
)

func pagedView(items []any, page domain.Page, total int64) map[string]any {
	totalPages := int64(0)
	if page.Size > 0 {
		totalPages = (total + int64(page.Size) - 1) / int64(page.Size)
	}
	return map[string]any{"items": items, "pagination": map[string]any{
		"page": page.Number, "page_size": page.Size, "total": total,
		"total_pages": totalPages,
	}}
}

func wireTime(value time.Time) string {
	return value.UTC().Truncate(time.Microsecond).Format("2006-01-02T15:04:05.000000Z")
}
