package dashboard

import (
	"context"
	"errors"
	app "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
	"testing"
)

func TestListFailsClosedWithoutBoundScope(t *testing.T) {
	_, err := (&store{}).List(context.Background(), 1, nil)
	if !errors.Is(err, app.ErrUnavailable) {
		t.Fatal(err)
	}
}
