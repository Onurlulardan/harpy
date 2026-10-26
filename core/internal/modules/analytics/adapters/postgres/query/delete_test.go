package query

import (
	"context"
	"errors"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
	"testing"
)

func TestDeleteFailsClosedWithoutBoundScope(t *testing.T) {
	_, _, err := (&store{}).Delete(context.Background(), queryapp.DeleteMutation{})
	if !errors.Is(err, queryapp.ErrUnavailable) {
		t.Fatal(err)
	}
}
