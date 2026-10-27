package projections

import (
	"context"
	"errors"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
	"testing"
)

func TestConsumeProjectionRejectsEmptyInput(t *testing.T) {
	var service *Service
	_, err := service.ConsumeProjectionEvent(context.Background(), nil, ports.Mutation{}, ports.ProjectionItem{})
	if !errors.Is(err, ports.ErrInvalid) {
		t.Fatal(err)
	}
}
