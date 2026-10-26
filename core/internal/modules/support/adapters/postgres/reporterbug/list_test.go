package reporterbug

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

func TestListMyBugsFailsClosedWithoutDriver(t *testing.T) {
	_, err := (*Store)(nil).ListMyBugs(context.Background(), ports.Access{}, domain.BugFilter{})
	if !errors.Is(err, ports.ErrUnavailable) {
		t.Fatalf("list my bugs error = %v, want unavailable", err)
	}
}
