package adoption

import (
	"context"
	"errors"
	"testing"

	application "github.com/Onurlulardan/harpy/core/internal/modules/support/application/adoption"
)

func TestCompleteFailsClosedWithoutAdopterStore(t *testing.T) {
	_, err := (*Store)(nil).AdoptionComplete(context.Background(), application.LeaseCommand{})
	if !errors.Is(err, application.ErrUnavailable) {
		t.Fatalf("complete error = %v, want unavailable", err)
	}
}
