package calendar

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func TestUpdateCalendarRejectsEmptyInput(t *testing.T) {
	var service *Service
	_, err := service.UpdateCalendarEvent(context.Background(), nil, ports.Mutation{}, [16]byte{}, CalendarEventInput{})
	if !errors.Is(err, ports.ErrInvalid) {
		t.Fatal(err)
	}
}
