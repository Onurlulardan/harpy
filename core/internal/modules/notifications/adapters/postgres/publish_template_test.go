package postgres

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/domain"
)

func TestPublishTemplateRejectsInvalidRevision(t *testing.T) {
	if err := (&Store{}).PublishTemplate(t.Context(), domain.ID{}, domain.ID{}, 0); err != ErrInvalid {
		t.Fatalf("PublishTemplate() error = %v", err)
	}
}
