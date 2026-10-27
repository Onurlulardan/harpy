package adoption

import (
	"context"
	"errors"
	"testing"

	application "github.com/Onurlulardan/harpy/core/internal/modules/support/application/adoption"
)

func TestRecordAdvanceFailsClosedWithoutAdopterStore(t *testing.T) {
	_, err := (*Store)(nil).AdoptionRecordAdvance(context.Background(), application.ApplyPageCommand{})
	if !errors.Is(err, application.ErrUnavailable) {
		t.Fatalf("advance error = %v, want unavailable", err)
	}
}
