package tasks

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func TestCreateTaskRejectsEmptyInput(t *testing.T) {
	var service *Service
	_, err := service.CreateTask(context.Background(), nil, ports.Mutation{}, CreateTaskInput{})
	if !errors.Is(err, ports.ErrInvalid) {
		t.Fatal(err)
	}
}
