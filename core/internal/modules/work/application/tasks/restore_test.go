package tasks

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func TestRestoreTaskRejectsInvalidID(t *testing.T) {
	var service *Service
	_, err := service.RestoreTask(context.Background(), nil, ports.Mutation{}, [16]byte{})
	if !errors.Is(err, ports.ErrInvalid) {
		t.Fatal(err)
	}
}
