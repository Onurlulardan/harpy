package tasks

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func TestUpdateTaskRejectsInvalidInput(t *testing.T) {
	var service *Service
	_, err := service.UpdateTask(context.Background(), nil, ports.Mutation{}, UpdateInput{})
	if !errors.Is(err, ports.ErrInvalid) {
		t.Fatal(err)
	}
}
