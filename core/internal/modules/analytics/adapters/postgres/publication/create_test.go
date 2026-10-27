package publication

import (
	"context"
	"errors"
	app "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/publication"
	"testing"
)

func TestCreateFailsClosedWithoutBoundScope(t *testing.T) {
	_, _, err := (&store{}).Create(context.Background(), app.CreateMutation{})
	if !errors.Is(err, app.ErrUnavailable) {
		t.Fatal(err)
	}
}
