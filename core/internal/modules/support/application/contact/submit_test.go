package contact

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/support/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/support/ports"
)

func TestSubmitContactFailsClosedWithoutKernel(t *testing.T) {
	err := (*Service)(nil).SubmitContact(context.Background(), domain.ContactInput{}, [32]byte{}, "", [32]byte{})
	if !errors.Is(err, ports.ErrUnavailable) {
		t.Fatalf("submit contact error = %v, want unavailable", err)
	}
}
