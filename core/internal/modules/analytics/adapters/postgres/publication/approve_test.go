package publication

import (
	"context"
	"errors"
	app "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/publication"
	"testing"
)

func TestApproveFailsClosedWithoutBoundScope(t *testing.T) {
	_, _, err := (&store{}).Approve(context.Background(), app.TransitionMutation{})
	if !errors.Is(err, app.ErrUnavailable) {
		t.Fatal(err)
	}
}
