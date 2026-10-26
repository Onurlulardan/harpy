package tasks

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func TestReopenTaskRejectsInvalidID(t *testing.T) {
	var service *Service
	_, err := service.ReopenTask(context.Background(), nil, ports.Mutation{}, [16]byte{})
	if !errors.Is(err, ports.ErrInvalid) {
		t.Fatal(err)
	}
}
