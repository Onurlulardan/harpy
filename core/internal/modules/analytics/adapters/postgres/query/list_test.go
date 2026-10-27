package query

import (
	"context"
	"errors"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
	"testing"
)

func TestListFailsClosedWithoutBoundScope(t *testing.T) {
	_, err := (&store{}).List(context.Background(), 1, nil)
	if !errors.Is(err, queryapp.ErrUnavailable) {
		t.Fatal(err)
	}
}
