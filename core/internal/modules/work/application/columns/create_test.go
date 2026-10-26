package columns

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func TestCreateColumnRejectsEmptyInput(t *testing.T) {
	var service *Service
	_, err := service.CreateColumn(context.Background(), nil, ports.Mutation{}, CreateColumnInput{})
	if !errors.Is(err, ports.ErrInvalid) {
		t.Fatal(err)
	}
}
