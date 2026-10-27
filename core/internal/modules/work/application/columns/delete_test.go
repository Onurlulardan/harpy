package columns

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func TestDeleteColumnRejectsInvalidInput(t *testing.T) {
	var service *Service
	_, err := service.DeleteColumn(context.Background(), nil, ports.Mutation{}, DeleteInput{})
	if !errors.Is(err, ports.ErrInvalid) {
		t.Fatalf("delete invalid result = %v", err)
	}
}
