package reporterbug

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

func TestCreateBugFailsClosedWithoutDriver(t *testing.T) {
	_, _, err := (*Store)(nil).CreateBug(context.Background(), ports.BugSubmission{})
	if !errors.Is(err, ports.ErrUnavailable) {
		t.Fatalf("create bug error = %v, want unavailable", err)
	}
}
