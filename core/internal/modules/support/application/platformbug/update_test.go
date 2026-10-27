package platformbug

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

func TestUpdatePlatformBugFailsClosedWithoutKernel(t *testing.T) {
	_, err := (*Service)(nil).UpdatePlatformBug(context.Background(), ports.Access{}, domain.BugUpdate{})
	if !errors.Is(err, ports.ErrUnavailable) {
		t.Fatalf("update platform bug error = %v, want unavailable", err)
	}
}
