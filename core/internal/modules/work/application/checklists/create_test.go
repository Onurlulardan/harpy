package checklists

import (
	"context"
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func TestCreateChecklistRejectsEmptyInput(t *testing.T) {
	var service *Service
	_, err := service.CreateChecklistItem(context.Background(), nil, ports.Mutation{}, CreateChecklistInput{})
	if !errors.Is(err, ports.ErrInvalid) {
		t.Fatal(err)
	}
}
