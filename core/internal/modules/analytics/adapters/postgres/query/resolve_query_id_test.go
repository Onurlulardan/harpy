package query

import (
	"context"
	"errors"
	"testing"

	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
	"github.com/google/uuid"
)

func TestResolveQueryIDFailsClosedWithoutBoundScope(t *testing.T) {
	_, err := (&store{}).ResolveQueryID(context.Background(), uuid.Nil)
	if !errors.Is(err, queryapp.ErrUnavailable) {
		t.Fatal(err)
	}
}
