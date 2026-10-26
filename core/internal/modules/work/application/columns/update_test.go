package columns

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func TestUpdateColumnRejectsInvalidInput(t *testing.T) {
	var service *Service
	_, err := service.UpdateColumn(context.Background(), nil, ports.Mutation{}, UpdateInput{})
	if !errors.Is(err, ports.ErrInvalid) {
		t.Fatal(err)
	}
}
