package dashboard

import (
	"context"
	"errors"
	app "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
	"testing"
)

func TestUpdateFailsClosedWithoutBoundScope(t *testing.T) {
	_, _, err := (&store{}).Update(context.Background(), app.UpdateMutation{})
	if !errors.Is(err, app.ErrUnavailable) {
		t.Fatal(err)
	}
}
