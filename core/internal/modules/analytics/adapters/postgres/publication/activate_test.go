package publication

import (
	"context"
	"errors"
	app "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/publication"
	"testing"
)

func TestActivateFailsClosedWithoutBoundScope(t *testing.T) {
	_, _, err := (&store{}).Activate(context.Background(), app.ActivateMutation{})
	if !errors.Is(err, app.ErrUnavailable) {
		t.Fatal(err)
	}
}
