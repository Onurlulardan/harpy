package v1cutover

import (
	"testing"

	app "github.com/Onurlulardan/harpy/core/internal/modules/notifications/application/v1cutover"
)

func TestCreateAuthorityRejectsUnknownOrigin(t *testing.T) {
	if _, err := (&Store{}).CreateAuthority(t.Context(), app.Origin("other")); err != app.ErrInvalid {
		t.Fatalf("CreateAuthority() error = %v", err)
	}
}
