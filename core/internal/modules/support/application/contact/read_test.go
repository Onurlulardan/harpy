package contact

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

func TestReadContactStatsFailsClosedWithoutKernel(t *testing.T) {
	_, err := (*Service)(nil).ReadContactStats(context.Background(), ports.Access{})
	if !errors.Is(err, ports.ErrUnavailable) {
		t.Fatalf("read contact stats error = %v, want unavailable", err)
	}
}
