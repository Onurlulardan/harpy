package reporterbug

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

func TestCreateBugFailsClosedWithoutKernel(t *testing.T) {
	_, err := (*Service)(nil).CreateBug(context.Background(), ports.Access{}, domain.BugInput{}, "", [32]byte{})
	if !errors.Is(err, ports.ErrUnavailable) {
		t.Fatalf("create bug error = %v, want unavailable", err)
	}
}
