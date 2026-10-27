package recurrence

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func TestUpdateRecurrenceRejectsEmptyInput(t *testing.T) {
	var service *Service
	_, err := service.UpdateRecurrence(
		context.Background(), nil, ports.Mutation{}, RecurrenceInput{}, 0,
	)
	if !errors.Is(err, ports.ErrInvalid) {
		t.Fatal(err)
	}
}
