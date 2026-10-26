package boards

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func TestCreateBoardRejectsEmptyInput(t *testing.T) {
	var service *Service
	_, err := service.CreateBoard(context.Background(), nil, ports.Mutation{}, CreateBoardInput{})
	if !errors.Is(err, ports.ErrInvalid) {
		t.Fatal(err)
	}
}
