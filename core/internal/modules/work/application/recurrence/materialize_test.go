package recurrence

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func TestMaterializeRejectsEmptyInput(t *testing.T) {
	var service *Service
	_, err := service.MaterializeRecurrence(context.Background(), nil, ports.Mutation{}, [16]byte{}, time.Time{}, time.Time{}, 0)
	if !errors.Is(err, ports.ErrInvalid) {
		t.Fatal(err)
	}
}
