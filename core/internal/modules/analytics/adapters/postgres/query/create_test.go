package query

import (
	"context"
	"errors"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
	"testing"
)

func TestCreateFailsClosedWithoutBoundScope(t *testing.T) {
	_, _, err := (&store{}).Create(context.Background(), queryapp.CreateMutation{})
	if !errors.Is(err, queryapp.ErrUnavailable) {
		t.Fatal(err)
	}
}
