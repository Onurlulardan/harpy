package publication

import (
	"context"
	"errors"
	app "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/publication"
	"github.com/google/uuid"
	"testing"
)

func TestGetFailsClosedWithoutBoundScope(t *testing.T) {
	_, err := (&store{}).Get(context.Background(), uuid.Nil)
	if !errors.Is(err, app.ErrUnavailable) {
		t.Fatal(err)
	}
}
