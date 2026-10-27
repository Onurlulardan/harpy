package publication

import (
	"context"
	"errors"
	app "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/publication"
	"testing"
)

func TestSubmitFailsClosedWithoutBoundScope(t *testing.T) {
	_, _, err := (&store{}).Submit(context.Background(), app.TransitionMutation{})
	if !errors.Is(err, app.ErrUnavailable) {
		t.Fatal(err)
	}
}
