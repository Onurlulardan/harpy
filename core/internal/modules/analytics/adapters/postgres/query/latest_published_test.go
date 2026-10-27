package query

import (
	"context"
	"errors"
	"testing"

	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
	"github.com/google/uuid"
)

func TestLatestPublishedFailsClosedWithoutBoundScope(t *testing.T) {
	_, err := (&store{}).LatestPublished(context.Background(), uuid.Nil)
	if !errors.Is(err, queryapp.ErrUnavailable) {
		t.Fatal(err)
	}
}
