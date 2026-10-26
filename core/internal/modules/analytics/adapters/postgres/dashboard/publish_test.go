package dashboard

import (
	"context"
	"errors"
	app "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
	"testing"
)

func TestPublishFailsClosedWithoutBoundScope(t *testing.T) {
	_, _, err := (&store{}).Publish(context.Background(), app.PublishMutation{})
	if !errors.Is(err, app.ErrUnavailable) {
		t.Fatal(err)
	}
}
