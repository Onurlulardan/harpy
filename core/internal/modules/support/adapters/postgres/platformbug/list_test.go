package platformbug

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

func TestListPlatformBugsFailsClosedWithoutDriver(t *testing.T) {
	_, err := (*Store)(nil).ListPlatformBugs(context.Background(), ports.Access{}, domain.BugFilter{})
	if !errors.Is(err, ports.ErrUnavailable) {
		t.Fatalf("list platform bugs error = %v, want unavailable", err)
	}
}
