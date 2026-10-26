package query

import (
	"context"
	"errors"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/query"
	"testing"
)

func TestPublishFailsClosedWithoutBoundScope(t *testing.T) {
	_, _, err := (&store{}).Publish(context.Background(), queryapp.PublishMutation{})
	if !errors.Is(err, queryapp.ErrUnavailable) {
		t.Fatal(err)
	}
}
