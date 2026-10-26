package dashboard

import (
	"context"
	"errors"
	app "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
	"github.com/google/uuid"
	"testing"
)

func TestGetFailsClosedWithoutBoundScope(t *testing.T) {
	_, err := (&store{}).Get(context.Background(), uuid.Nil)
	if !errors.Is(err, app.ErrUnavailable) {
		t.Fatal(err)
	}
}
