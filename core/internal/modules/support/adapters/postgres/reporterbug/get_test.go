package reporterbug

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

func TestGetMyBugFailsClosedWithoutDriver(t *testing.T) {
	_, err := (*Store)(nil).GetMyBug(context.Background(), ports.Access{}, domain.NilID, domain.NilID)
	if !errors.Is(err, ports.ErrUnavailable) {
		t.Fatalf("get my bug error = %v, want unavailable", err)
	}
}
