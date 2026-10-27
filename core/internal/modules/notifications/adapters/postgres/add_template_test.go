package postgres

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func TestAddTemplateRejectsInvalidDraft(t *testing.T) {
	if err := (&Store{}).AddTemplate(t.Context(), ports.TemplateDraft{}); err != ErrInvalid {
		t.Fatalf("AddTemplate() error = %v", err)
	}
}
