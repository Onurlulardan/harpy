package calendar

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func TestCreateCalendarRejectsEmptyInput(t *testing.T) {
	var service *Service
	_, err := service.CreateCalendarEvent(context.Background(), nil, ports.Mutation{}, CalendarEventInput{})
	if !errors.Is(err, ports.ErrInvalid) {
		t.Fatal(err)
	}
}
