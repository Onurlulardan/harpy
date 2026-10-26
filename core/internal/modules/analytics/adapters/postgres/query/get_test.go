package query

import (
	"context"
	"errors"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
	"github.com/google/uuid"
	"testing"
)

func TestGetFailsClosedWithoutBoundScope(t *testing.T) {
	_, err := (&store{}).Get(context.Background(), uuid.Nil)
	if !errors.Is(err, queryapp.ErrUnavailable) {
		t.Fatal(err)
	}
}
