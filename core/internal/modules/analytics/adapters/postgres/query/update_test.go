package query

import (
	"context"
	"errors"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
	"testing"
)

func TestUpdateFailsClosedWithoutBoundScope(t *testing.T) {
	_, _, err := (&store{}).Update(context.Background(), queryapp.UpdateMutation{})
	if !errors.Is(err, queryapp.ErrUnavailable) {
		t.Fatal(err)
	}
}
