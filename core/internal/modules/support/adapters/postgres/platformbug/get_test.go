package platformbug

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

func TestGetPlatformBugFailsClosedWithoutDriver(t *testing.T) {
	_, err := (*Store)(nil).GetPlatformBug(context.Background(), ports.Access{}, domain.NilID, domain.NilID)
	if !errors.Is(err, ports.ErrUnavailable) {
		t.Fatalf("get platform bug error = %v, want unavailable", err)
	}
}
