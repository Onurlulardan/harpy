package memberships

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func TestChangeMembershipRejectsInvalidInput(t *testing.T) {
	var service *Service
	_, err := service.ChangeMembership(context.Background(), nil, ports.Mutation{}, ChangeMembershipInput{})
	if !errors.Is(err, ports.ErrInvalid) {
		t.Fatal(err)
	}
}
