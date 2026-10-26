package labels

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func TestSetTaskLabelRejectsEmptyInput(t *testing.T) {
	var service *Service
	_, err := service.SetTaskLabel(context.Background(), nil, ports.Mutation{}, [16]byte{}, [16]byte{}, true)
	if !errors.Is(err, ports.ErrInvalid) {
		t.Fatal(err)
	}
}
