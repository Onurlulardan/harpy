package contact

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

func TestGetContactFailsClosedWithoutDriver(t *testing.T) {
	_, err := (*Store)(nil).GetContact(context.Background(), ports.Access{}, domain.NilID)
	if !errors.Is(err, ports.ErrUnavailable) {
		t.Fatalf("get contact error = %v, want unavailable", err)
	}
}
