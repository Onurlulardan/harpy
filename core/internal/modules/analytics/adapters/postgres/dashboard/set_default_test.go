package dashboard

import (
	"context"
	"errors"
	app "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
	"testing"
)

func TestSetDefaultFailsClosedWithoutBoundScope(t *testing.T) {
	_, err := (&store{}).SetDefault(context.Background(), app.DefaultMutation{})
	if !errors.Is(err, app.ErrUnavailable) {
		t.Fatal(err)
	}
}
