package boards

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func TestUpdateBoardRejectsEmptyInput(t *testing.T) {
	var service *Service
	_, err := service.UpdateBoard(context.Background(), nil, ports.Mutation{},
		[16]byte{}, UpdateBoardInput{})
	if !errors.Is(err, ports.ErrInvalid) {
		t.Fatal(err)
	}
}
