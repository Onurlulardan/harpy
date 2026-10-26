package recurrence

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func TestCreateRecurrenceRejectsEmptyInput(t *testing.T) {
	var service *Service
	_, err := service.CreateRecurrence(context.Background(), nil, ports.Mutation{}, RecurrenceInput{})
	if !errors.Is(err, ports.ErrInvalid) {
		t.Fatal(err)
	}
}
