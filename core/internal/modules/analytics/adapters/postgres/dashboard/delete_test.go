package dashboard

import (
	"context"
	"errors"
	app "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
	"testing"
)

func TestDeleteFailsClosedWithoutBoundScope(t *testing.T) {
	_, _, err := (&store{}).Delete(context.Background(), app.DeleteMutation{})
	if !errors.Is(err, app.ErrUnavailable) {
		t.Fatal(err)
	}
}
