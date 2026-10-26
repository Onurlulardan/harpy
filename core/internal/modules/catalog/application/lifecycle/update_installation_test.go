package lifecycle

import (
	"context"
	"errors"
	"testing"
)

func TestUpdateInstallationRejectsMissingStore(t *testing.T) {
	_, err := UpdateInstallation(context.Background(), nil, InstallationMutation{})
	if !errors.Is(err, ErrUnavailable) {
		t.Fatal(err)
	}
}
