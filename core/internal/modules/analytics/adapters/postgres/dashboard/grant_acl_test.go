package dashboard

import (
	"context"
	"errors"
	app "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/dashboard"
	"testing"
)

func TestGrantACLFailsClosedWithoutBoundScope(t *testing.T) {
	_, _, err := (&store{}).GrantACL(context.Background(), app.ACLMutation{})
	if !errors.Is(err, app.ErrUnavailable) {
		t.Fatal(err)
	}
}
