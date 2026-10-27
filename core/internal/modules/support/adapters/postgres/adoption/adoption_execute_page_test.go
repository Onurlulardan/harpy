package adoption

import (
	"context"
	"errors"
	"testing"

	application "github.com/Onurlulardan/harpy/core/internal/modules/support/application/adoption"
)

func TestExecutePageRejectsUnknownSourceKind(t *testing.T) {
	_, err := (*Store)(nil).AdoptionExecutePage(context.Background(), application.ApplyPageCommand{})
	if !errors.Is(err, application.ErrInvalid) {
		t.Fatalf("execute page error = %v, want invalid", err)
	}
}
