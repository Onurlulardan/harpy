package contact

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

func TestSubmitContactFailsClosedWithoutDriver(t *testing.T) {
	_, _, err := (*Store)(nil).SubmitContact(context.Background(), ports.ContactSubmission{})
	if !errors.Is(err, ports.ErrUnavailable) {
		t.Fatalf("submit contact error = %v, want unavailable", err)
	}
}
