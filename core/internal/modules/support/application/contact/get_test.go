package contact

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

func TestGetContactFailsClosedWithoutKernel(t *testing.T) {
	_, err := (*Service)(nil).GetContact(context.Background(), ports.Access{}, domain.NilID)
	if !errors.Is(err, ports.ErrUnavailable) {
		t.Fatalf("get contact error = %v, want unavailable", err)
	}
}
