package adoption

import (
	"context"
	"errors"
	"testing"

	application "github.com/Onurlulardan/harpy/core/internal/modules/support/application/adoption"
)

func TestStartFailsClosedWithoutAdopterStore(t *testing.T) {
	_, _, err := (*Store)(nil).AdoptionStart(context.Background(), application.StartCommand{})
	if !errors.Is(err, application.ErrUnavailable) {
		t.Fatalf("start error = %v, want unavailable", err)
	}
}
