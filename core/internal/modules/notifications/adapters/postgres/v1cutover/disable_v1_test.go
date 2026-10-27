package v1cutover

import (
	"testing"

	app "github.com/Onurlulardan/harpy/core/internal/modules/notifications/application/v1cutover"
)

func TestDisableV1RejectsInvalidEvidence(t *testing.T) {
	if _, err := (&Store{}).DisableV1(t.Context(), app.Authority{}, app.Evidence{}); err != app.ErrInvalid {
		t.Fatalf("DisableV1() error = %v", err)
	}
}
