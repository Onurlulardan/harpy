package comments

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func TestCreateCommentRejectsEmptyInput(t *testing.T) {
	var service *Service
	_, err := service.CreateComment(context.Background(), nil, ports.Mutation{}, [16]byte{}, [16]byte{}, "")
	if !errors.Is(err, ports.ErrInvalid) {
		t.Fatal(err)
	}
}
