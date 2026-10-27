package v1cutover

import (
	"testing"

	app "github.com/Onurlulardan/harpy/core/internal/modules/notifications/application/v1cutover"
)

func TestPauseV1RejectsInvalidAuthority(t *testing.T) {
	if _, err := (&Store{}).PauseV1(t.Context(), app.Authority{}); err != app.ErrInvalid {
		t.Fatalf("PauseV1() error = %v", err)
	}
}
