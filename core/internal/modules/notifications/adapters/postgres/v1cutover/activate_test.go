package v1cutover

import (
	"testing"

	app "github.com/Onurlulardan/harpy/core/internal/modules/notifications/application/v1cutover"
)

func TestActivateV2RejectsInvalidAuthority(t *testing.T) {
	if _, err := (&Store{}).ActivateV2(t.Context(), app.Authority{}); err != app.ErrInvalid {
		t.Fatalf("ActivateV2() error = %v", err)
	}
}
